package tool

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// TavilySearch is a tool for searching using Tavily API.
// Tavily is a search API specifically designed for AI agents.
type TavilySearch struct {
	APIKey    string
	UserAgent string
	Timeout   time.Duration
	MaxResults int
	SearchDepth string // "basic" or "advanced"
}

// TavilyOption is a function type for configuring TavilySearch.
type TavilyOption func(*TavilySearch)

// WithTavilyAPIKey sets the Tavily API key.
func WithTavilyAPIKey(apiKey string) TavilyOption {
	return func(t *TavilySearch) {
		t.APIKey = apiKey
	}
}

// WithTavilyUserAgent sets a custom User-Agent header.
func WithTavilyUserAgent(ua string) TavilyOption {
	return func(t *TavilySearch) {
		t.UserAgent = ua
	}
}

// WithTavilyTimeout sets the request timeout.
func WithTavilyTimeout(timeout time.Duration) TavilyOption {
	return func(t *TavilySearch) {
		t.Timeout = timeout
	}
}

// WithTavilyMaxResults sets the maximum number of results to return.
func WithTavilyMaxResults(maxResults int) TavilyOption {
	return func(t *TavilySearch) {
		t.MaxResults = maxResults
	}
}

// WithTavilySearchDepth sets the search depth ("basic" or "advanced").
func WithTavilySearchDepth(depth string) TavilyOption {
	return func(t *TavilySearch) {
		if depth == "basic" || depth == "advanced" {
			t.SearchDepth = depth
		}
	}
}

// NewTavilySearch creates a new TavilySearch tool.
func NewTavilySearch(opts ...TavilyOption) (*TavilySearch, error) {
	t := &TavilySearch{
		UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/137.0.0.0 Safari/537.36 Edg/137.0.0.0",
		Timeout:   30 * time.Second,
		MaxResults: 10,
		SearchDepth: "basic",
	}

	for _, opt := range opts {
		opt(t)
	}

	// Check if API key is set
	if t.APIKey == "" {
		// Try to get from environment variable
		// Note: We'll check this at runtime, not during initialization
	}

	return t, nil
}

// Name returns the name of the tool.
func (t *TavilySearch) Name() string {
	return "Tavily_Search"
}

// Description returns the description of the tool.
func (t *TavilySearch) Description() string {
	return "Tavily AI 搜索引擎。一个专为 AI 代理设计的强大搜索 API，提供准确、实时的搜索结果。" +
		"支持网络搜索、新闻搜索，并能自动提取和总结网页内容。" +
		"需要 Tavily API Key（可通过环境变量 TAVILY_API_KEY 设置）。"
}

// Call executes the search and returns formatted results.
func (t *TavilySearch) Call(ctx context.Context, input string) (string, error) {
	results, err := t.Search(ctx, input, t.MaxResults)
	if err != nil {
		return "", fmt.Errorf("搜索失败: %w", err)
	}
	return FormatResults(results, input), nil
}

// Search implements the SearchTool interface.
func (t *TavilySearch) Search(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	// Create a context with timeout if not already set
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, t.Timeout)
		defer cancel()
	}

	// Get API key - first from struct, then from environment variable
	apiKey := t.APIKey
	if apiKey == "" {
		apiKey = os.Getenv("TAVILY_API_KEY")
	}
	if apiKey == "" {
		return nil, fmt.Errorf("Tavily API key not set. Please set TAVILY_API_KEY environment variable or provide it via WithTavilyAPIKey option")
	}

	// Prepare request payload
	requestPayload := map[string]interface{}{
		"api_key":      apiKey,
		"query":        query,
		"search_depth": t.SearchDepth,
		"max_results":  limit,
	}

	// Add optional parameters for better results
	requestPayload["include_answer"] = true
	requestPayload["include_raw_content"] = false
	requestPayload["include_images"] = false

	// Marshal to JSON
	jsonData, err := json.Marshal(requestPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.tavily.com/search", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", t.UserAgent)

	// Execute request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Tavily API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var tavilyResponse TavilyResponse
	if err := json.NewDecoder(resp.Body).Decode(&tavilyResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Convert to SearchResult format
	results := make([]SearchResult, 0, len(tavilyResponse.Results))
	for _, item := range tavilyResponse.Results {
		// Use content if available, otherwise use snippet
		content := item.Content
		if content == "" {
			content = item.Snippet
		}

		results = append(results, SearchResult{
			Title:       item.Title,
			URL:         item.URL,
			Author:      item.Author,
			PublishTime: formatDate(item.PublishedDate),
			Summary:     item.Snippet,
			Content:     content,
			Source:      "Tavily",
		})
	}

	return results, nil
}

// TavilyResponse represents the response from Tavily API.
type TavilyResponse struct {
	Answer  string `json:"answer"`
	Query   string `json:"query"`
	Results []TavilyResult `json:"results"`
}

// TavilyResult represents a single search result from Tavily.
type TavilyResult struct {
	Title         string `json:"title"`
	URL           string `json:"url"`
	Content       string `json:"content"`
	Score         float64 `json:"score"`
	RawContent    string `json:"raw_content"`
	Snippet       string `json:"snippet"`
	PublishedDate string `json:"published_date"`
	Author        string `json:"author"`
}

// formatDate formats a date string in a more readable way.
func formatDate(dateStr string) string {
	if dateStr == "" {
		return ""
	}

	// Parse the date (Tavily returns ISO 8601 format)
	t, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		return dateStr
	}

	// Return in a more readable format
	return t.Format("2006-01-02")
}

// setDefaultHeaders sets common HTTP headers for requests.
func (t *TavilySearch) setDefaultHeaders(req *http.Request) {
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", t.UserAgent)
}
