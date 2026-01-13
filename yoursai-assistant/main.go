package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

var (
	dbPool *sql.DB
	dbOnce sync.Once
)

type Request struct {
	SessionId string `json:"sessionId"`
	Message   string `json:"message"`
}

func extractUserFromToken(authHeader string) (string, error) {
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return "", fmt.Errorf("missing or invalid authorization header")
	}
	
	token := strings.TrimPrefix(authHeader, "Bearer ")
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return "", fmt.Errorf("invalid JWT format")
	}
	
	// Decode payload (second part)
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", fmt.Errorf("failed to decode JWT payload")
	}
	
	var claims map[string]interface{}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return "", fmt.Errorf("failed to parse JWT claims")
	}
	
	// Extract user ID from 'sub' claim
	if sub, ok := claims["sub"].(string); ok {
		return sub, nil
	}
	
	return "", fmt.Errorf("user ID not found in token")
}

type Response struct {
	Reply     string `json:"reply"`
	SessionId string `json:"sessionId"`
}

func retryWithBackoff(fn func() (*http.Response, error), maxRetries int) (*http.Response, error) {
	retryDelays := []int{2, 5, 10} // configurable retry delays in seconds
	
	for i := 0; i < maxRetries; i++ {
		resp, err := fn()
		if err != nil {
			return nil, err
		}
		
		// Retry on rate limiting (429) and service unavailable (503)
		if resp.StatusCode != 429 && resp.StatusCode != 503 {
			return resp, nil
		}
		
		resp.Body.Close()
		
		if i == maxRetries-1 {
			return resp, nil // Return the last error response
		}
		
		// Use configurable retry delays
		delaySeconds := retryDelays[i]
		delay := time.Duration(delaySeconds) * time.Second
		log.Printf("API error %d, retrying in %v (attempt %d/%d)", resp.StatusCode, delay, i+1, maxRetries)
		time.Sleep(delay)
	}
	return nil, fmt.Errorf("max retries exceeded")
}

func getChatHistory(ctx context.Context, userId, sessionId string) ([]string, error) {
	cfg, _ := config.LoadDefaultConfig(ctx)
	db := dynamodb.NewFromConfig(cfg)

	out, err := db.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String("ChatHistory"),
		KeyConditionExpression: aws.String("user_id = :uid AND session_id = :sid"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":uid": &types.AttributeValueMemberS{Value: userId},
			":sid": &types.AttributeValueMemberS{Value: sessionId},
		},
		Limit:            aws.Int32(10),
		ScanIndexForward: aws.Bool(true), // oldest first for chronological order
	})
	if err != nil {
		return nil, err
	}

	history := []string{}
	for _, item := range out.Items {
		userMsg, userOk := item["userMessage"].(*types.AttributeValueMemberS)
		aiMsg, aiOk := item["aiReply"].(*types.AttributeValueMemberS)
		if userOk && aiOk {
			history = append(history,
				"User: "+userMsg.Value+"\nAI: "+aiMsg.Value)
		}
	}
	return history, nil
}

func saveChat(ctx context.Context, userId, sessionId, userMsg, aiMsg string) error {
	cfg, _ := config.LoadDefaultConfig(ctx)
	db := dynamodb.NewFromConfig(cfg)

	// Save new chat
	_, err := db.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String("ChatHistory"),
		Item: map[string]types.AttributeValue{
			"user_id":     &types.AttributeValueMemberS{Value: userId},
			"session_id":  &types.AttributeValueMemberS{Value: sessionId},
			"timestamp":   &types.AttributeValueMemberS{Value: time.Now().Format(time.RFC3339)},
			"userMessage": &types.AttributeValueMemberS{Value: userMsg},
			"aiReply":     &types.AttributeValueMemberS{Value: aiMsg},
		},
	})
	if err != nil {
		return err
	}

	// Clean up old chats to maintain limit
	cleanupOldChats(ctx, userId, sessionId, db)
	return nil
}

func cleanupOldChats(ctx context.Context, userId, sessionId string, db *dynamodb.Client) {
	// Get all chats for this user and session, ordered by timestamp
	out, err := db.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String("ChatHistory"),
		KeyConditionExpression: aws.String("user_id = :uid AND session_id = :sid"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":uid": &types.AttributeValueMemberS{Value: userId},
			":sid": &types.AttributeValueMemberS{Value: sessionId},
		},
		ScanIndexForward: aws.Bool(false), // newest first
	})
	if err != nil {
		return
	}

	// If we have more than MaxChatsPerSession, delete the oldest ones
	if len(out.Items) > MaxChatsPerSession {
		for i := MaxChatsPerSession; i < len(out.Items); i++ {
			item := out.Items[i]
			db.DeleteItem(ctx, &dynamodb.DeleteItemInput{
				TableName: aws.String("ChatHistory"),
				Key: map[string]types.AttributeValue{
					"user_id":    item["user_id"],
					"session_id": item["session_id"],
				},
			})
		}
	}
}

func getParameter(ctx context.Context, name string) (string, error) {
	cfg, _ := config.LoadDefaultConfig(ctx)
	client := ssm.NewFromConfig(cfg)

	out, err := client.GetParameter(ctx, &ssm.GetParameterInput{
		Name:           aws.String(name),
		WithDecryption: aws.Bool(true),
	})
	if err != nil {
		return "", err
	}
	return *out.Parameter.Value, nil
}

func getDBPool(ctx context.Context) (*sql.DB, error) {
	var err error
	dbOnce.Do(func() {
		host, e1 := getParameter(ctx, DBHostPath)
		username, e2 := getParameter(ctx, DBUsernamePath)
		password, e3 := getParameter(ctx, DBPasswordPath)
		database, e4 := getParameter(ctx, DBDatabasePath)
		port, e5 := getParameter(ctx, DBPortPath)
		
		if e1 != nil || e2 != nil || e3 != nil || e4 != nil || e5 != nil {
			err = fmt.Errorf("failed to get DB parameters")
			return
		}

		connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=require",
			host, port, username, password, database)

		dbPool, err = sql.Open("postgres", connStr)
		if err != nil {
			return
		}

		// Configure connection pool
		dbPool.SetMaxOpenConns(MaxOpenConns)
		dbPool.SetMaxIdleConns(MaxIdleConns)
		dbPool.SetConnMaxLifetime(time.Duration(ConnMaxLifetime) * time.Minute)
		
		err = dbPool.Ping()
	})
	
	return dbPool, err
}

func generateEmbedding(ctx context.Context, text string, apiKey string) ([]float64, error) {
	payload := map[string]interface{}{
		"model": EmbeddingModel,
		"content": map[string]interface{}{
			"parts": []map[string]string{
				{"text": text},
			},
		},
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequestWithContext(ctx, "POST", EmbeddingAPIURL+"?key="+apiKey, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	embedding := result["embedding"].(map[string]interface{})["values"].([]interface{})
	embedVec := make([]float64, len(embedding))
	for i, v := range embedding {
		embedVec[i] = v.(float64)
	}
	
	return embedVec, nil
}

func vectorSearch(ctx context.Context, db *sql.DB, embedding []float64, userId string) ([]string, error) {
	// Use PostgreSQL array parameter directly - no manual string construction
	query := `SELECT content, document_name FROM aiknowledge WHERE user_id = $2 ORDER BY embedding <-> $1 LIMIT 3`
	rows, err := db.Query(query, embedding, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []string
	for rows.Next() {
		var content, docName string
		if err := rows.Scan(&content, &docName); err != nil {
			continue
		}
		
		// Include document name in the result
		if docName != "" {
			results = append(results, fmt.Sprintf("[From: %s] %s", docName, content))
		} else {
			results = append(results, content)
		}
	}
	
	return results, nil
}

func getGeminiKey(ctx context.Context) (string, error) {
	cfg, _ := config.LoadDefaultConfig(ctx)
	client := ssm.NewFromConfig(cfg)

	out, err := client.GetParameter(ctx, &ssm.GetParameterInput{
		Name:           aws.String("yoursAI_apiKey"),
		WithDecryption: aws.Bool(true),
	})
	if err != nil {
		return "", err
	}
	return *out.Parameter.Value, nil
}

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.Printf("Received API Gateway request")
	
	// Extract user ID from Authorization header
	authHeader := request.Headers["Authorization"]
	if authHeader == "" {
		authHeader = request.Headers["authorization"] // case-insensitive fallback
	}
	
	userId, err := extractUserFromToken(authHeader)
	if err != nil {
		log.Printf("Authentication failed: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: 401,
			Headers: map[string]string{
				"Access-Control-Allow-Origin": "*",
				"Content-Type": "application/json",
			},
			Body: `{"error": "Authentication required"}`,
		}, nil
	}
	
	var req Request
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		log.Printf("Error parsing request body: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Headers: map[string]string{
				"Access-Control-Allow-Origin": "*",
				"Content-Type": "application/json",
			},
			Body: `{"error": "Invalid request body"}`,
		}, nil
	}
	
	// Generate session ID if not provided (security fix)
	sessionId := req.SessionId
	if sessionId == "" {
		sessionId = uuid.New().String()
	}
	
	// Structured logging for observability
	fmt.Printf(`{"session":"%s","msg":"%s","time":"%s"}`+"\n",
		sessionId, req.Message, time.Now().Format(time.RFC3339))
	
	// Validate message length
	if len(req.Message) > MaxMessageLength {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Headers: map[string]string{
				"Access-Control-Allow-Origin": "*",
				"Content-Type": "application/json",
			},
			Body: `{"reply": "Message too long. Please keep it under 2000 characters."}`,
		}, nil
	}
	
	// Handle greeting messages
	lowerMsg := strings.ToLower(strings.TrimSpace(req.Message))
	greetings := []string{"hi", "hello", "hey", "good morning", "good afternoon", "good evening"}
	for _, greeting := range greetings {
		if lowerMsg == greeting || lowerMsg == greeting+"!" {
			response := Response{
				Reply:     "Hello, How can I help you today?",
				SessionId: sessionId,
			}
			responseBody, _ := json.Marshal(response)
			return events.APIGatewayProxyResponse{
				StatusCode: 200,
				Headers: map[string]string{
					"Access-Control-Allow-Origin": "*",
					"Content-Type": "application/json",
				},
				Body: string(responseBody),
			}, nil
		}
	}
	
	apiKey, err := getGeminiKey(ctx)
	if err != nil {
		log.Printf("Error getting API key: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Headers: map[string]string{
				"Access-Control-Allow-Origin": "*",
				"Content-Type": "application/json",
			},
			Body: `{"error": "Failed to get API key"}`,
		}, nil
	}

	// Get conversation history for context
	history, _ := getChatHistory(ctx, userId, sessionId)

	conversation := ""
	for _, h := range history {
		conversation += h + "\n"
	}

	// Perform vector search for relevant knowledge (only for substantial queries)
	db, err := getDBPool(ctx)
	var vectorContext string
	useRag := len(req.Message) > 30
	
	if err == nil && useRag {
		// Add delay before API call
		time.Sleep(time.Duration(APICallDelay) * time.Millisecond)
		
		// Generate embedding for user question
		embedding, err := generateEmbedding(ctx, req.Message, apiKey)
		if err == nil {
			// Search for similar content
			searchResults, err := vectorSearch(ctx, db, embedding, userId)
			if err == nil && len(searchResults) > 0 {
				vectorContext = "\n\nRelevant Knowledge:\n"
				for _, result := range searchResults {
					vectorContext += "- " + result + "\n"
				}
			}
		}
	}

	// Add delay before main API call
	time.Sleep(time.Duration(APICallDelay) * time.Millisecond)

	// Build final prompt in correct order: System → Conversation → Documents → User
	finalPrompt := SystemPrompt
	if conversation != "" {
		finalPrompt += "\n\nConversation so far:\n" + conversation
	}
	if vectorContext != "" {
		finalPrompt += vectorContext
	}
	finalPrompt += "\nUser: " + req.Message

	payload := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]string{
					{"text": finalPrompt},
				},
			},
		},
		"generationConfig": map[string]interface{}{
			"maxOutputTokens": MaxOutputTokens,
			"temperature":     Temperature,
		},
	}

	body, _ := json.Marshal(payload)
	log.Printf("Sending request to AI API")

	client := &http.Client{Timeout: 30 * time.Second}
	
	// Use retry logic for rate limiting
	resp, err := retryWithBackoff(func() (*http.Response, error) {
		// Create fresh request each time to avoid body consumption issues
		httpReq, _ := http.NewRequestWithContext(ctx, "POST", GeminiAPIURL+"?key="+apiKey, bytes.NewBuffer(body))
		httpReq.Header.Set("Content-Type", "application/json")
		return client.Do(httpReq)
	}, 3)
	
	if err != nil {
		log.Printf("Error calling Gemini API: %v", "Request failed - timeout or network error")
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Headers: map[string]string{
				"Access-Control-Allow-Origin": "*",
				"Content-Type": "application/json",
			},
			Body: `{"error": "Failed to call AI service"}`,
		}, nil
	}
	defer resp.Body.Close()

	log.Printf("Got response from AI API with status code: %d", resp.StatusCode)

	// Fail-safe for AI API downtime
	if resp.StatusCode != 200 {
		log.Printf("AI API returned non-200 status: %d", resp.StatusCode)
		response := Response{
			Reply:     "AI service is temporarily unavailable. Try again.",
			SessionId: sessionId,
		}
		responseBody, _ := json.Marshal(response)
		return events.APIGatewayProxyResponse{
			StatusCode: 200,
			Headers: map[string]string{
				"Access-Control-Allow-Origin": "*",
				"Content-Type": "application/json",
			},
			Body: string(responseBody),
		}, nil
	}

	var geminiResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&geminiResp); err != nil {
		log.Printf("Error decoding response: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Headers: map[string]string{
				"Access-Control-Allow-Origin": "*",
				"Content-Type": "application/json",
			},
			Body: `{"error": "Failed to decode AI response"}`,
		}, nil
	}

	// Check if response has expected structure
	candidates, ok := geminiResp["candidates"].([]interface{})
	if !ok || len(candidates) == 0 {
		log.Printf("No candidates in response")
		return events.APIGatewayProxyResponse{
			StatusCode: 200,
			Headers: map[string]string{
				"Access-Control-Allow-Origin": "*",
				"Content-Type": "application/json",
			},
			Body: `{"reply": "No response from AI"}`,
		}, nil
	}

	candidate, candidateOk := candidates[0].(map[string]interface{})
	if !candidateOk {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Headers: map[string]string{
				"Access-Control-Allow-Origin": "*",
				"Content-Type": "application/json",
			},
			Body: `{"error": "Invalid AI response format"}`,
		}, nil
	}
	
	content, contentOk := candidate["content"].(map[string]interface{})
	parts, partsOk := content["parts"].([]interface{})
	if !contentOk || !partsOk || len(parts) == 0 {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Headers: map[string]string{
				"Access-Control-Allow-Origin": "*",
				"Content-Type": "application/json",
			},
			Body: `{"error": "Invalid AI response format"}`,
		}, nil
	}
	
	partMap, partOk := parts[0].(map[string]interface{})
	reply, replyOk := partMap["text"].(string)
	if !partOk || !replyOk {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Headers: map[string]string{
				"Access-Control-Allow-Origin": "*",
				"Content-Type": "application/json",
			},
			Body: `{"error": "Invalid AI response format"}`,
		}, nil
	}

	// Check if response was truncated due to token limit
	finishReason, hasFinishReason := candidate["finishReason"].(string)
	if hasFinishReason && finishReason == "MAX_TOKENS" {
		// Auto-continue the response
		continuePrompt := SystemPrompt
		if conversation != "" {
			continuePrompt += "\n\nConversation so far:\n" + conversation
		}
		continuePrompt += "\nUser: " + req.Message + "\nAI: " + reply + "\nUser: Continue exactly from the last word. Do not repeat. Complete the previous response."
		
		continuePayload := map[string]interface{}{
			"contents": []map[string]interface{}{
				{
					"parts": []map[string]string{
						{"text": continuePrompt},
					},
				},
			},
			"generationConfig": map[string]interface{}{
				"maxOutputTokens": MaxOutputTokens,
				"temperature":     Temperature,
			},
		}

		continueBody, _ := json.Marshal(continuePayload)
		continueReq, _ := http.NewRequestWithContext(ctx, "POST", GeminiAPIURL+"?key="+apiKey, bytes.NewBuffer(continueBody))
		continueReq.Header.Set("Content-Type", "application/json")

		continueResp, err := client.Do(continueReq)
		if err == nil && continueResp.StatusCode == 200 {
			defer continueResp.Body.Close()
			var continueGeminiResp map[string]interface{}
			if json.NewDecoder(continueResp.Body).Decode(&continueGeminiResp) == nil {
				if continueCandidates, ok := continueGeminiResp["candidates"].([]interface{}); ok && len(continueCandidates) > 0 {
					continueCandidate := continueCandidates[0].(map[string]interface{})
					continueContent := continueCandidate["content"].(map[string]interface{})
					continueParts := continueContent["parts"].([]interface{})
					continueText := continueParts[0].(map[string]interface{})["text"].(string)
					reply += continueText
				}
			}
		}
	}

	// Save chat to DynamoDB
	err = saveChat(ctx, userId, sessionId, req.Message, reply)
	if err != nil {
		log.Printf("DynamoDB error: %v", err)
	}

	log.Printf("Successfully processed request, returning response")
	
	response := Response{
		Reply:     reply,
		SessionId: sessionId,
	}
	responseBody, _ := json.Marshal(response)
	
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Access-Control-Allow-Origin": "*",
			"Content-Type": "application/json",
		},
		Body: string(responseBody),
	}, nil
}

func main() {
	lambda.Start(handler)
}
