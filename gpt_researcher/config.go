package main

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds all configuration parameters for the GPT Researcher system
type Config struct {
	// LLM Configuration
	OpenAIAPIKey string
	Model        string
	ReportModel  string
	SummaryModel string

	// Search Configuration
	TavilyAPIKey     string
	MaxSearchResults int
	MaxSourcesToUse  int

	// Research Configuration
	MaxQuestions  int
	MaxIterations int
	ReportType    string // "research_report", "outline_report", "resource_report", etc.
	ReportFormat  string // "markdown", "pdf", "docx"

	// Output Configuration
	OutputDir        string
	SaveIntermediate bool
	Verbose          bool
}

// NewConfig creates a new configuration with defaults and environment variable overrides
func NewConfig() *Config {
	config := &Config{
		// Defaults
		Model:            getEnvOrDefault("GPT_MODEL", "deepseek-v3"),
		ReportModel:      getEnvOrDefault("REPORT_MODEL", "deepseek-v3"),
		SummaryModel:     getEnvOrDefault("SUMMARY_MODEL", "deepseek-v3"),
		MaxSearchResults: getEnvIntOrDefault("MAX_SEARCH_RESULTS", 10),
		MaxSourcesToUse:  getEnvIntOrDefault("MAX_SOURCES_TO_USE", 20),
		MaxQuestions:     getEnvIntOrDefault("MAX_QUESTIONS", 5),
		MaxIterations:    getEnvIntOrDefault("MAX_ITERATIONS", 5),
		ReportType:       getEnvOrDefault("REPORT_TYPE", "research_report"),
		ReportFormat:     getEnvOrDefault("REPORT_FORMAT", "markdown"),
		OutputDir:        getEnvOrDefault("OUTPUT_DIR", "./output"),
		SaveIntermediate: getEnvBoolOrDefault("SAVE_INTERMEDIATE", false),
		Verbose:          getEnvBoolOrDefault("VERBOSE", true),
	}

	// Required API keys
	config.OpenAIAPIKey = os.Getenv("OPENAI_API_KEY")
	if config.OpenAIAPIKey == "" {
		fmt.Println("Warning: OPENAI_API_KEY not set")
	}

	config.TavilyAPIKey = os.Getenv("TAVILY_API_KEY")
	if config.TavilyAPIKey == "" {
		fmt.Println("Warning: TAVILY_API_KEY not set")
	}

	return config
}

// getEnvOrDefault gets an environment variable or returns a default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvIntOrDefault gets an integer environment variable or returns a default value
func getEnvIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvBoolOrDefault gets a boolean environment variable or returns a default value
func getEnvBoolOrDefault(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// String returns a string representation of the configuration
func (c *Config) String() string {
	return fmt.Sprintf(`GPT Researcher Configuration:
  Model: %s
  Report Model: %s
  Summary Model: %s
  Max Search Results: %d
  Max Sources To Use: %d
  Max Questions: %d
  Max Iterations: %d
  Report Type: %s
  Report Format: %s
  Output Directory: %s
  Verbose: %v`,
		c.Model,
		c.ReportModel,
		c.SummaryModel,
		c.MaxSearchResults,
		c.MaxSourcesToUse,
		c.MaxQuestions,
		c.MaxIterations,
		c.ReportType,
		c.ReportFormat,
		c.OutputDir,
		c.Verbose,
	)
}
