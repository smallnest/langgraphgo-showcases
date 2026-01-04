package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"
)

// RetryConfig configures retry behavior
type RetryConfig struct {
	MaxRetries    int           // Maximum number of retry attempts
	InitialDelay  time.Duration // Initial delay before first retry
	MaxDelay      time.Duration // Maximum delay between retries
	BackoffFactor float64       // Multiplier for delay after each retry
}

// DefaultRetryConfig returns sensible default retry settings
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:    3,
		InitialDelay:  1 * time.Second,
		MaxDelay:      30 * time.Second,
		BackoffFactor: 2.0,
	}
}

// RetryableFunc is a function that can be retried
type RetryableFunc func() error

// RetryWithContext executes a function with retry logic
func RetryWithContext(ctx context.Context, config RetryConfig, fn RetryableFunc) error {
	var lastErr error
	delay := config.InitialDelay

	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		// Check if we should retry this error
		if !IsRetryableError(err) {
			return fmt.Errorf("non-retryable error: %w", err)
		}

		if attempt < config.MaxRetries {
			log.Printf("[Retry] Attempt %d/%d failed: %v. Retrying in %v...", attempt+1, config.MaxRetries, err, delay)
			time.Sleep(delay)

			// Exponential backoff with jitter
			delay = time.Duration(float64(delay) * config.BackoffFactor)
			if delay > config.MaxDelay {
				delay = config.MaxDelay
			}
		}
	}

	return fmt.Errorf("max retries (%d) exceeded, last error: %w", config.MaxRetries, lastErr)
}

// IsRetryableError determines if an error should trigger a retry
func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()

	// Network-related errors (typically retryable)
	retryablePatterns := []string{
		"connection refused",
		"connection reset",
		"timeout",
		"deadline exceeded",
		"temporary failure",
		"rate limit",
		"too many requests",
		"service unavailable",
		"gateway timeout",
		"bad gateway",
		"EOF",
		"broken pipe",
	}

	for _, pattern := range retryablePatterns {
		if strings.Contains(strings.ToLower(errStr), pattern) {
			return true
		}
	}

	return false
}

// JSONParseError wraps JSON parsing errors with context
type JSONParseError struct {
	Content string
	Err     error
}

func (e *JSONParseError) Error() string {
	return fmt.Sprintf("JSON parse error: %w. Content preview: %s", e.Err, truncateString(e.Content, 200))
}

func (e *JSONParseError) Unwrap() error {
	return e.Err
}

// SafeJSONUnmarshal safely unmarshals JSON with detailed error reporting
func SafeJSONUnmarshal(data []byte, v any) error {
	if !json.Valid(data) {
		return &JSONParseError{
			Content: string(data),
			Err:     fmt.Errorf("invalid JSON format"),
		}
	}

	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(v); err != nil {
		return &JSONParseError{
			Content: string(data),
			Err:     err,
		}
	}

	return nil
}

// CleanJSONContent cleans common JSON formatting issues from LLM responses
func CleanJSONContent(content string) string {
	// Remove markdown code blocks
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	// Fix common LLM JSON issues
	content = strings.ReplaceAll(content, "\u201C", "\"") // Left double quotation mark
	content = strings.ReplaceAll(content, "\u201D", "\"") // Right double quotation mark
	content = strings.ReplaceAll(content, "\u2018", "'")  // Left single quotation mark
	content = strings.ReplaceAll(content, "\u2019", "'")  // Right single quotation mark
	content = strings.ReplaceAll(content, "\u2026", "...") // Ellipsis

	// Remove control characters except newlines and tabs
	var result strings.Builder
	for _, r := range content {
		if r == '\n' || r == '\t' || r >= ' ' {
			result.WriteRune(r)
		}
	}

	return result.String()
}

// GenerateJSONWithRetry generates JSON from LLM with retry and repair logic
func GenerateJSONWithRetry(
	ctx context.Context,
	llm LLMProvider,
	systemPrompt, userContent string,
	output any,
	config RetryConfig,
) error {
	var lastErr error

	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Generate content from LLM
		content, err := llm.GenerateContent(ctx, systemPrompt, userContent)
		if err != nil {
			if IsRetryableError(err) && attempt < config.MaxRetries {
				log.Printf("[JSON Retry] LLM call failed (attempt %d/%d): %v", attempt+1, config.MaxRetries, err)
				time.Sleep(config.InitialDelay * time.Duration(attempt+1))
				continue
			}
			return fmt.Errorf("LLM generation failed: %w", err)
		}

		// Clean JSON content
		content = CleanJSONContent(content)

		// Try to parse
		if err := SafeJSONUnmarshal([]byte(content), output); err != nil {
			lastErr = err

			if attempt < config.MaxRetries {
				log.Printf("[JSON Retry] JSON parse failed (attempt %d/%d): %v", attempt+1, config.MaxRetries, err)

				// Try to repair common JSON issues
				if repaired := TryRepairJSON(content); repaired != "" {
					if repairErr := SafeJSONUnmarshal([]byte(repaired), output); repairErr == nil {
						log.Printf("[JSON Retry] Successfully repaired JSON on attempt %d", attempt+1)
						return nil
					}
				}

				time.Sleep(config.InitialDelay * time.Duration(attempt+1))
				continue
			}

			return fmt.Errorf("JSON parsing failed after %d attempts: %w", config.MaxRetries+1, lastErr)
		}

		return nil
	}

	return lastErr
}

// TryRepairJSON attempts to repair common JSON formatting issues
func TryRepairJSON(content string) string {
	// Try to extract JSON from markdown if present
	if strings.Contains(content, "```json") {
		if start := strings.Index(content, "```json"); start != -1 {
			start += 7
			if end := strings.Index(content[start:], "```"); end != -1 {
				return strings.TrimSpace(content[start : start+end])
			}
		}
	}

	// Try to extract JSON from triple backticks
	if strings.Contains(content, "```") {
		if start := strings.Index(content, "```"); start != -1 {
			start += 3
			if end := strings.Index(content[start:], "```"); end != -1 {
				candidate := strings.TrimSpace(content[start : start+end])
				if json.Valid([]byte(candidate)) {
					return candidate
				}
			}
		}
	}

	// Try to find JSON object boundaries
	content = strings.TrimSpace(content)
	if strings.HasPrefix(content, "{") && strings.HasSuffix(content, "}") {
		// Try to balance braces
		depth := 0
		for i, c := range content {
			if c == '{' {
				depth++
			} else if c == '}' {
				depth--
				if depth == 0 {
					return content[:i+1]
				}
			}
		}
	}

	return ""
}

// LLMProvider interface for LLM calls
type LLMProvider interface {
	GenerateContent(ctx context.Context, systemPrompt, userContent string) (string, error)
}

// truncateString truncates a string to a maximum length
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// ErrorHandler provides centralized error handling
type ErrorHandler struct {
	LogErrors bool
}

// NewErrorHandler creates a new error handler
func NewErrorHandler(logErrors bool) *ErrorHandler {
	return &ErrorHandler{LogErrors: logErrors}
}

// Handle handles an error with logging and optional transformation
func (h *ErrorHandler) Handle(err error, context string) error {
	if err == nil {
		return nil
	}

	if h.LogErrors {
		log.Printf("[Error] %s: %v", context, err)
	}

	return fmt.Errorf("%s: %w", context, err)
}

// Wrap wraps an error with additional context
func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
}
