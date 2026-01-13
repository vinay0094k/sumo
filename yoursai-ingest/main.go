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
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	_ "github.com/lib/pq"
)

type IngestRequest struct {
	DocumentName string `json:"documentName"`
	Text         string `json:"text"`
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

type IngestResponse struct {
	Message string `json:"message"`
	Chunks  int    `json:"chunks"`
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

func connectDB(ctx context.Context) (*sql.DB, error) {
	host, err := getParameter(ctx, DBHostPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get host: %v", err)
	}
	
	username, err := getParameter(ctx, DBUsernamePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get username: %v", err)
	}
	
	password, err := getParameter(ctx, DBPasswordPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get password: %v", err)
	}
	
	database, err := getParameter(ctx, DBDatabasePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get database: %v", err)
	}
	
	port, err := getParameter(ctx, DBPortPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get port: %v", err)
	}

	// Validate parameters
	if host == "" || username == "" || password == "" || database == "" || port == "" {
		return nil, fmt.Errorf("missing database parameters: host=%s, user=%s, db=%s, port=%s", host, username, database, port)
	}

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=require",
		host, port, username, password, database)
	
	log.Printf("Connecting to database with host=%s port=%s dbname=%s", host, port, database)
	
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open connection: %v", err)
	}
	
	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}
	
	return db, nil
}

// Chunk text into ~500 token chunks (roughly 375 words)
func chunkText(text string, maxTokens int) []string {
	words := strings.Fields(text)
	wordsPerChunk := maxTokens * 3 / 4 // Rough conversion: 1 token ≈ 0.75 words
	
	var chunks []string
	for i := 0; i < len(words); i += wordsPerChunk {
		end := i + wordsPerChunk
		if end > len(words) {
			end = len(words)
		}
		chunk := strings.Join(words[i:end], " ")
		chunks = append(chunks, chunk)
	}
	
	return chunks
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
	req, _ := http.NewRequest("POST", EmbeddingAPIURL+"?key="+apiKey, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: HTTPTimeout * time.Second}
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

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
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
	
	var req IngestRequest
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Headers: map[string]string{
				"Access-Control-Allow-Origin": "*",
				"Content-Type": "application/json",
			},
			Body: `{"error": "Invalid request body"}`,
		}, nil
	}

	// Get API key
	apiKey, err := getParameter(ctx, SSMKeyPath)
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

	// Connect to database
	conn, err := connectDB(ctx)
	if err != nil {
		log.Printf("DB connection failed: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Headers: map[string]string{
				"Access-Control-Allow-Origin": "*",
				"Content-Type": "application/json",
			},
			Body: `{"error": "Database connection failed"}`,
		}, nil
	}
	defer conn.Close()

	// Chunk the text into ~500 token chunks
	chunks := chunkText(req.Text, MaxTokensPerChunk)
	
	// Process each chunk
	for i, chunkText := range chunks {
		log.Printf("Processing chunk %d/%d for document: %s", i+1, len(chunks), req.DocumentName)
		
		// Generate embedding
		embeddingVector, err := generateEmbedding(ctx, chunkText, apiKey)
		if err != nil {
			log.Printf("Embedding generation failed for chunk %d: %v", i+1, err)
			continue
		}
		
		// Convert embedding to PostgreSQL array format
		embedStr := "["
		for j, v := range embeddingVector {
			if j > 0 {
				embedStr += ","
			}
			embedStr += fmt.Sprintf("%f", v)
		}
		embedStr += "]"
		
		// Insert into aiknowledge table with document name and user_id
		_, err = conn.Exec(
			"INSERT INTO aiknowledge (content, embedding, document_name, user_id) VALUES ($1, $2, $3, $4)",
			chunkText,
			embedStr,
			req.DocumentName,
			userId,
		)
		if err != nil {
			log.Printf("Vector storage failed for chunk %d: %v", i+1, err)
			continue
		}
	}

	response := IngestResponse{
		Message: fmt.Sprintf("Document '%s' ingested successfully", req.DocumentName),
		Chunks:  len(chunks),
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
