package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/tools"
)

// TavilySearchTool implements web search using Tavily API
type TavilySearchTool struct {
	APIKey     string
	MaxResults int
}

func NewTavilySearchTool(apiKey string, maxResults int) *TavilySearchTool {
	return &TavilySearchTool{
		APIKey:     apiKey,
		MaxResults: maxResults,
	}
}

func (t *TavilySearchTool) Name() string {
	return "tavily_search"
}

func (t *TavilySearchTool) Description() string {
	return "Search the web using Tavily API. Input should be a search query string. Returns relevant web search results with URLs, titles, and content snippets."
}

func (t *TavilySearchTool) Call(ctx context.Context, input string) (string, error) {
	if t.APIKey == "" {
		return "", fmt.Errorf("Tavily API key not set")
	}

	// Prepare Tavily API request
	requestBody := map[string]any{
		"api_key":             t.APIKey,
		"query":               input,
		"max_results":         t.MaxResults,
		"search_depth":        "advanced",
		"include_answer":      false,
		"include_raw_content": true,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make API request
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.tavily.com/search", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("tavily API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var result TavilyResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	// Format results
	var output strings.Builder
	output.WriteString(fmt.Sprintf("Found %d results for query: %s\n\n", len(result.Results), input))

	for i, r := range result.Results {
		output.WriteString(fmt.Sprintf("Result %d:\n", i+1))
		output.WriteString(fmt.Sprintf("Title: %s\n", r.Title))
		output.WriteString(fmt.Sprintf("URL: %s\n", r.URL))
		output.WriteString(fmt.Sprintf("Content: %s\n", r.Content))
		if r.Score > 0 {
			output.WriteString(fmt.Sprintf("Relevance Score: %.2f\n", r.Score))
		}
		output.WriteString("\n")
	}

	return output.String(), nil
}

// TavilyResponse represents the response from Tavily API
type TavilyResponse struct {
	Results []TavilyResult `json:"results"`
	Query   string         `json:"query"`
}

// TavilyResult represents a single search result from Tavily
type TavilyResult struct {
	Title      string  `json:"title"`
	URL        string  `json:"url"`
	Content    string  `json:"content"`
	Score      float64 `json:"score"`
	RawContent string  `json:"raw_content"`
}

// WebScraperTool scrapes web pages (simplified implementation)
type WebScraperTool struct{}

func NewWebScraperTool() *WebScraperTool {
	return &WebScraperTool{}
}

func (t *WebScraperTool) Name() string {
	return "web_scraper"
}

func (t *WebScraperTool) Description() string {
	return "Scrape content from a web page. Input should be a URL. Returns the text content of the page."
}

func (t *WebScraperTool) Call(ctx context.Context, input string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", input, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; GPTResearcher/1.0)")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP error: status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1024*1024)) // Limit to 1MB
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Simple text extraction (in production, use a proper HTML parser)
	content := string(body)

	// Remove HTML tags (very basic)
	content = strings.ReplaceAll(content, "<script", " <script")
	content = strings.ReplaceAll(content, "<style", " <style")

	return fmt.Sprintf("Content from %s (length: %d chars):\n%s", input, len(content), content[:min(len(content), 5000)]), nil
}

// SummarizerTool summarizes text using an LLM
type SummarizerTool struct {
	Model llms.Model
}

func NewSummarizerTool(model llms.Model) *SummarizerTool {
	return &SummarizerTool{Model: model}
}

func (t *SummarizerTool) Name() string {
	return "summarizer"
}

func (t *SummarizerTool) Description() string {
	return "Summarize text content. Input should be the text to summarize. Returns a concise summary."
}

func (t *SummarizerTool) Call(ctx context.Context, input string) (string, error) {
	prompt := fmt.Sprintf(`Please provide a concise summary of the following text, focusing on key facts and insights:

%s

Summary:`, input)

	messages := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeHuman, prompt),
	}

	resp, err := t.Model.GenerateContent(ctx, messages)
	if err != nil {
		return "", fmt.Errorf("failed to generate summary: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response from model")
	}

	return resp.Choices[0].Content, nil
}

// ToolRegistry holds all available tools
type ToolRegistry struct {
	SearchTool     *TavilySearchTool
	ScraperTool    *WebScraperTool
	SummarizerTool *SummarizerTool
}

// NewToolRegistry creates a new tool registry
func NewToolRegistry(config *Config, model llms.Model) *ToolRegistry {
	return &ToolRegistry{
		SearchTool:     NewTavilySearchTool(config.TavilyAPIKey, config.MaxSearchResults),
		ScraperTool:    NewWebScraperTool(),
		SummarizerTool: NewSummarizerTool(model),
	}
}

// GetTools returns all tools as a slice
func (tr *ToolRegistry) GetTools() []tools.Tool {
	return []tools.Tool{
		tr.SearchTool,
		tr.ScraperTool,
		tr.SummarizerTool,
	}
}
