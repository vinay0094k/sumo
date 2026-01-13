package main

const (
	// AI Models
	EmbeddingModel     = "text-embedding-004"
	EmbeddingAPIURL    = "https://generativelanguage.googleapis.com/v1beta/models/" + EmbeddingModel + ":embedContent"
	
	// AWS Configuration
	SSMKeyPath         = "yoursAI_apiKey"
	AWSRegion          = "us-east-1"
	
	// Chunking Configuration
	MaxTokensPerChunk  = 500
	HTTPTimeout        = 10 // seconds
	
	// Database Parameter Store paths
	DBHostPath         = "/yoursai/db/host"
	DBUsernamePath     = "/yoursai/db/username"
	DBPasswordPath     = "/yoursai/db/password"
	DBDatabasePath     = "/yoursai/db/database"
	DBPortPath         = "/yoursai/db/port"
)
