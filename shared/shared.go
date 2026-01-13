package shared

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	_ "github.com/lib/pq"
)

func ExtractUserFromToken(authHeader string) (string, error) {
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return "", fmt.Errorf("missing or invalid authorization header")
	}
	
	token := strings.TrimPrefix(authHeader, "Bearer ")
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return "", fmt.Errorf("invalid JWT format")
	}
	
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", fmt.Errorf("failed to decode JWT payload")
	}
	
	var claims map[string]interface{}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return "", fmt.Errorf("failed to parse JWT claims")
	}
	
	if sub, ok := claims["sub"].(string); ok {
		return sub, nil
	}
	
	return "", fmt.Errorf("user ID not found in token")
}

func GetParameter(ctx context.Context, paramName string) (string, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return "", err
	}
	
	ssmClient := ssm.NewFromConfig(cfg)
	out, err := ssmClient.GetParameter(ctx, &ssm.GetParameterInput{
		Name:           aws.String(paramName),
		WithDecryption: aws.Bool(true),
	})
	if err != nil {
		return "", err
	}
	return *out.Parameter.Value, nil
}

func ConnectDB(ctx context.Context) (*sql.DB, error) {
	host, err := GetParameter(ctx, "/yoursai/db/host")
	if err != nil {
		return nil, err
	}
	
	username, err := GetParameter(ctx, "/yoursai/db/username")
	if err != nil {
		return nil, err
	}
	
	password, err := GetParameter(ctx, "/yoursai/db/password")
	if err != nil {
		return nil, err
	}
	
	dbname, err := GetParameter(ctx, "/yoursai/db/database")
	if err != nil {
		return nil, err
	}
	
	port, err := GetParameter(ctx, "/yoursai/db/port")
	if err != nil {
		return nil, err
	}

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=require",
		host, port, username, password, dbname)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

func GenerateEmbedding(ctx context.Context, text, apiKey string) ([]float64, error) {
	payload := map[string]interface{}{
		"model": "text-embedding-004",
		"content": map[string]interface{}{
			"parts": []map[string]string{
				{"text": text},
			},
		},
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "https://generativelanguage.googleapis.com/v1beta/models/text-embedding-004:embedContent?key="+apiKey, bytes.NewBuffer(body))
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

	embeddingMap, embOk := result["embedding"].(map[string]interface{})
	if !embOk {
		return nil, fmt.Errorf("invalid embedding response format")
	}
	
	values, valOk := embeddingMap["values"].([]interface{})
	if !valOk {
		return nil, fmt.Errorf("invalid embedding values format")
	}
	
	embedVec := make([]float64, len(values))
	for i, v := range values {
		if floatVal, ok := v.(float64); ok {
			embedVec[i] = floatVal
		} else {
			return nil, fmt.Errorf("invalid embedding value type at index %d", i)
		}
	}
	
	return embedVec, nil
}
