// Package main implements the AI PDF Chatbot - a Go port of the original TypeScript/Node.js project.
// It uses LangGraphGo for graph orchestration and provides both CLI and HTTP server modes.
package main

import (
	"log"
	"os"
)

// Config holds the application configuration.
type Config struct {
	// Server configuration
	ServerPort string
	ServerHost string

	// LLM configuration
	OpenAIKey string
	Model     string

	// OpenAI API configuration
	OpenAIBaseURL string

	// Vector store configuration
	VectorStoreType string // "memory" or "supabase"
	SupabaseURL     string
	SupabaseKey     string

	// LangSmith tracing (optional)
	LangChainAPIKey  string
	LangChainProject string

	// Graph configuration
	IngestionGraphID string
	RetrievalGraphID string
	ChunkSize        int
	ChunkOverlap     int
	TopK             int // Number of documents to retrieve
}

// LoadConfig loads configuration from environment variables with sensible defaults.
func LoadConfig() Config {
	cfg := Config{
		ServerHost:       getEnv("SERVER_HOST", "0.0.0.0"),
		ServerPort:       getEnv("SERVER_PORT", "8080"),
		OpenAIKey:        os.Getenv("OPENAI_API_KEY"),
		Model:            getEnv("OPENAI_MODEL", "gpt-4o-mini"),
		OpenAIBaseURL:    getEnv("OPENAI_BASE_URL", ""),
		VectorStoreType:  getEnv("VECTOR_STORE_TYPE", "memory"),
		SupabaseURL:      os.Getenv("SUPABASE_URL"),
		SupabaseKey:      os.Getenv("SUPABASE_SERVICE_ROLE_KEY"),
		LangChainAPIKey:  os.Getenv("LANGCHAIN_API_KEY"),
		LangChainProject: getEnv("LANGCHAIN_PROJECT", "pdf-chatbot"),
		IngestionGraphID: getEnv("LANGGRAPH_INGESTION_ASSISTANT_ID", "ingestion_graph"),
		RetrievalGraphID: getEnv("LANGGRAPH_RETRIEVAL_ASSISTANT_ID", "retrieval_graph"),
		ChunkSize:        1000,
		ChunkOverlap:     200,
		TopK:             5,
	}
	return cfg
}

// ValidateConfig validates that required configuration is present.
func ValidateConfig(cfg Config) {
	if cfg.OpenAIKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is required")
	}

	if cfg.VectorStoreType == "supabase" {
		if cfg.SupabaseURL == "" {
			log.Fatal("SUPABASE_URL environment variable is required when using supabase vector store")
		}
		if cfg.SupabaseKey == "" {
			log.Fatal("SUPABASE_SERVICE_ROLE_KEY environment variable is required when using supabase vector store")
		}
	}
}

// getEnv retrieves an environment variable or returns a default value.
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
