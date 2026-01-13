package main

const (
	// AI Models
	GeminiEmbeddingModel = "text-embedding-004"
	OpenAIEmbeddingModel = "text-embedding-3-small"
	
	// API URLs
	GeminiEmbeddingAPIURL = "https://generativelanguage.googleapis.com/v1beta/models/" + GeminiEmbeddingModel + ":embedContent"
	OpenAIEmbeddingAPIURL = "https://api.openai.com/v1/embeddings"

	// AWS Configuration
	SSMKeyPath = "/yoursai/gemini/apiKey"
	AWSRegion  = "us-east-1"

	// Chunking Configuration
	MaxTokensPerChunk = 500
	HTTPTimeout       = 10              // seconds
	APICallDelay      = 2000            // milliseconds between API calls
	MaxDocumentSize   = 5 * 1024 * 1024 // 5MB max document size

	// Database Parameter Store paths
	DBHostPath     = "/yoursai/db/host"
	DBUsernamePath = "/yoursai/db/username"
	DBPasswordPath = "/yoursai/db/password"
	DBDatabasePath = "/yoursai/db/database"
	DBPortPath     = "/yoursai/db/port"
)
