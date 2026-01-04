package query_engine

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/smallnest/langgraphgo/showcases/DeepInsight/schema"
)

type TavilyResponse struct {
	Results []struct {
		Title         string  `json:"title"`
		URL           string  `json:"url"`
		Content       string  `json:"content"`
		Score         float64 `json:"score"`
		RawContent    string  `json:"raw_content"`
		PublishedDate string  `json:"published_date"`
	} `json:"results"`
	Images any `json:"images"` // Can be []string or []object
}

// SearchOptions holds optional search parameters
type SearchOptions struct {
	MaxResults    int
	SearchDepth   string
	Days          int
	IncludeImages bool
	Platform      string // For platform-specific searches
}

// DefaultSearchOptions returns default search options
func DefaultSearchOptions() SearchOptions {
	return SearchOptions{
		MaxResults:    10,
		SearchDepth:   "basic",
		Days:          0, // 0 means no date filter
		IncludeImages: false,
	}
}

// ExecuteSearch executes a search using Tavily API with enhanced options.
func ExecuteSearch(ctx context.Context, query string, toolName string, startDate, endDate string) ([]schema.SearchResult, error) {
	return ExecuteSearchWithOptions(ctx, query, toolName, startDate, endDate, DefaultSearchOptions())
}

// ExecuteSearchWithOptions executes a search with custom options.
func ExecuteSearchWithOptions(ctx context.Context, query string, toolName string, startDate, endDate string, opts SearchOptions) ([]schema.SearchResult, error) {
	apiKey := os.Getenv("TAVILY_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("TAVILY_API_KEY not set")
	}

	reqBody := map[string]any{
		"api_key": apiKey,
		"query":   query,
	}

	// Configure based on tool name
	switch toolName {
	case "basic_search_news":
		reqBody["search_depth"] = opts.SearchDepth
		reqBody["topic"] = "news"
		reqBody["max_results"] = opts.MaxResults
	case "deep_search_news":
		reqBody["search_depth"] = "advanced"
		reqBody["topic"] = "news"
		reqBody["max_results"] = opts.MaxResults
	case "search_news_last_24_hours":
		reqBody["search_depth"] = opts.SearchDepth
		reqBody["topic"] = "news"
		reqBody["days"] = 1
		reqBody["max_results"] = opts.MaxResults
	case "search_news_last_week":
		reqBody["search_depth"] = opts.SearchDepth
		reqBody["topic"] = "news"
		reqBody["days"] = 7
		reqBody["max_results"] = opts.MaxResults
	case "search_images_for_news":
		reqBody["search_depth"] = opts.SearchDepth
		reqBody["include_images"] = true
		reqBody["include_image_descriptions"] = true
		reqBody["max_results"] = opts.MaxResults
	case "search_news_by_date":
		reqBody["search_depth"] = "advanced"
		// Tavily supports date range in query
		if startDate != "" && endDate != "" {
			reqBody["query"] = fmt.Sprintf("%s after:%s before:%s", query, startDate, endDate)
		}
	default:
		reqBody["search_depth"] = opts.SearchDepth
		reqBody["topic"] = "general"
		reqBody["max_results"] = opts.MaxResults
	}

	// Apply days filter if specified and not already set
	if opts.Days > 0 && toolName != "search_news_last_24_hours" && toolName != "search_news_last_week" {
		reqBody["days"] = opts.Days
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request with timeout
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.tavily.com/search", strings.NewReader(string(jsonData)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Tavily API error (status %d): %s", resp.StatusCode, string(body))
	}

	var tavilyResp TavilyResponse
	if err := json.Unmarshal(body, &tavilyResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	var results []schema.SearchResult
	for _, r := range tavilyResp.Results {
		results = append(results, schema.SearchResult{
			Title:         r.Title,
			URL:           r.URL,
			Content:       r.Content,
			Score:         r.Score,
			RawContent:    r.RawContent,
			PublishedDate: r.PublishedDate,
		})
	}

	return results, nil
}

// ParseDateRange parses start and end date strings into time.Time
func ParseDateRange(startDate, endDate string) (start, end time.Time, err error) {
	if startDate != "" {
		start, err = time.Parse("2006-01-02", startDate)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid start date format (expected YYYY-MM-DD): %w", err)
		}
	}
	if endDate != "" {
		end, err = time.Parse("2006-01-02", endDate)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid end date format (expected YYYY-MM-DD): %w", err)
		}
	}
	return start, end, nil
}
