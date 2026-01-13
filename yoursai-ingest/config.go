package main

const (
	// AI Models
	EmbeddingModel  = "text-embedding-004"
	EmbeddingAPIURL = "https://generativelanguage.googleapis.com/v1beta/models/" + EmbeddingModel + ":embedContent"

	// AWS Configuration
	SSMKeyPath = "yoursAI_apiKey"
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
