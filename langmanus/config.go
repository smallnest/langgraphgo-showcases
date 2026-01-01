package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all configuration for LangManus
type Config struct {
	// LLM Configuration
	OpenAIAPIKey     string
	OpenAIBaseURL    string
	OpenAIModel      string
	OpenAIModelSmall string // For simpler tasks
	Temperature      float32

	// Search Configuration
	SearchAPIKey string // Tavily or similar search API
	SearchEngine string // "tavily", "serp", etc.

	// Code Execution
	EnableCodeExecution bool
	CodeTimeout         int // seconds

	// Browser Configuration
	EnableBrowser bool
	BrowserURL    string

	// Agent Configuration
	MaxIterations int
	Verbose       bool

	// Concurrency
	MaxConcurrentTasks int
}

// NewConfig creates a new configuration from environment variables
func NewConfig() *Config {
	// Try to load .env file (ignore error if file doesn't exist)
	_ = godotenv.Load()

	config := &Config{
		OpenAIAPIKey:        getEnv("OPENAI_API_KEY", ""),
		OpenAIBaseURL:       getEnv("OPENAI_BASE_URL", "https://qianfan.baidubce.com/v2"),
		OpenAIModel:         getEnv("OPENAI_MODEL", "deepseek-v3"),
		OpenAIModelSmall:    getEnv("OPENAI_MODEL_SMALL", "deepseek-v3"),
		Temperature:         getEnvFloat32("TEMPERATURE", 0.7),
		SearchAPIKey:        getEnv("SEARCH_API_KEY", ""),
		SearchEngine:        getEnv("SEARCH_ENGINE", "tavily"),
		EnableCodeExecution: getEnvBool("ENABLE_CODE_EXECUTION", true),
		CodeTimeout:         getEnvInt("CODE_TIMEOUT", 60),
		EnableBrowser:       getEnvBool("ENABLE_BROWSER", false),
		BrowserURL:          getEnv("BROWSER_URL", ""),
		MaxIterations:       getEnvInt("MAX_ITERATIONS", 15),
		Verbose:             getEnvBool("VERBOSE", true),
		MaxConcurrentTasks:  getEnvInt("MAX_CONCURRENT_TASKS", 3),
	}

	return config
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.OpenAIAPIKey == "" {
		return fmt.Errorf("OPENAI_API_KEY is required")
	}

	if c.SearchAPIKey == "" && c.SearchEngine != "" {
		fmt.Println("Warning: SEARCH_API_KEY not set, search functionality may be limited")
	}

	return nil
}

// String returns a string representation of the configuration
func (c *Config) String() string {
	return fmt.Sprintf(`LangManus Configuration:
  OpenAI Model: %s
  OpenAI Model (Small): %s
  Base URL: %s
  Temperature: %.2f
  Search Engine: %s
  Code Execution: %t
  Code Timeout: %d seconds
  Browser Enabled: %t
  Max Iterations: %d
  Max Concurrent Tasks: %d
  Verbose: %t
`,
		c.OpenAIModel,
		c.OpenAIModelSmall,
		c.OpenAIBaseURL,
		c.Temperature,
		c.SearchEngine,
		c.EnableCodeExecution,
		c.CodeTimeout,
		c.EnableBrowser,
		c.MaxIterations,
		c.MaxConcurrentTasks,
		c.Verbose,
	)
}

// Helper functions to get environment variables with defaults

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvFloat32(key string, defaultValue float32) float32 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 32); err == nil {
			return float32(floatValue)
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
