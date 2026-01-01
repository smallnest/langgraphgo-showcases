package config

import (
	"os"
	"strconv"
)

// Config 应用配置
type Config struct {
	// LLM配置
	LLMProvider    string
	LLMModel       string
	LLMAPIKey      string
	LLMBaseURL     string
	LLMTemperature float64
	LLMMaxTokens   int

	// 应用配置
	AppName    string
	AppVersion string
	Verbose    bool
	LogLevel   string

	// PDF处理配置
	MaxPDFSizeMB int

	// 分析配置
	DefaultDetailLevel string
	DefaultLanguage    string
	EnableHistory      bool

	// 备用模型列表
	FallbackModels []string
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	temperature, _ := strconv.ParseFloat(getEnv("LLM_TEMPERATURE", "0.3"), 64)
	maxTokens, _ := strconv.Atoi(getEnv("LLM_MAX_TOKENS", "4000"))
	maxPDFSize, _ := strconv.Atoi(getEnv("MAX_PDF_SIZE_MB", "20"))

	return &Config{
		// LLM配置
		LLMProvider:    getEnv("LLM_PROVIDER", "openai"),
		LLMModel:       getEnv("OPENAI_MODEL", "deepseek-v3"),
		LLMAPIKey:      getEnv("OPENAI_API_KEY", ""),
		LLMBaseURL:     getEnv("OPENAI_API_BASE", ""),
		LLMTemperature: temperature,
		LLMMaxTokens:   maxTokens,

		// 应用配置
		AppName:    "健康洞察代理",
		AppVersion: "1.0.0",
		Verbose:    getEnv("VERBOSE", "false") == "true",
		LogLevel:   getEnv("LOG_LEVEL", "info"),

		// PDF处理配置
		MaxPDFSizeMB: maxPDFSize,

		// 分析配置
		DefaultDetailLevel: getEnv("DETAIL_LEVEL", "Standard"),
		DefaultLanguage:    getEnv("LANGUAGE", "zh-CN"),
		EnableHistory:      getEnv("ENABLE_HISTORY", "true") == "true",

		// 备用模型
		FallbackModels: []string{
			"gpt-4",
			"gpt-3.5-turbo",
		},
	}
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// Validate 验证配置
func (c *Config) Validate() error {
	if c.LLMAPIKey == "" {
		return ErrMissingAPIKey
	}
	if c.MaxPDFSizeMB <= 0 {
		c.MaxPDFSizeMB = 20
	}
	if c.LLMTemperature < 0 || c.LLMTemperature > 2 {
		c.LLMTemperature = 0.3
	}
	if c.LLMMaxTokens <= 0 {
		c.LLMMaxTokens = 4000
	}
	return nil
}

// GetModelConfig 获取模型配置
func (c *Config) GetModelConfig() *ModelConfig {
	return &ModelConfig{
		Provider:       c.LLMProvider,
		PrimaryModel:   c.LLMModel,
		FallbackModels: c.FallbackModels,
		Temperature:    c.LLMTemperature,
		MaxTokens:      c.LLMMaxTokens,
		TopP:           0.95,
	}
}

// ModelConfig 模型配置
type ModelConfig struct {
	Provider       string
	PrimaryModel   string
	FallbackModels []string
	Temperature    float64
	MaxTokens      int
	TopP           float64
}
