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

	"github.com/smallnest/langgraphgo-showcases/DeepInsight/schema"
	"github.com/smallnest/langgraphgo-showcases/DeepInsight/tool"
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
	case "wechat_search":
		// Use WeChat official account article search
		wechatTool, err := tool.NewWeChatSearch(tool.WithWeChatCount(opts.MaxResults))
		if err != nil {
			return nil, fmt.Errorf("failed to create WeChat search tool: %w", err)
		}
		result, err := wechatTool.Call(ctx, query)
		if err != nil {
			return nil, fmt.Errorf("WeChat search failed: %w", err)
		}
		// Parse the WeChat search result and convert to schema.SearchResult
		return parseWeChatResult(result)
	case "basic_search":
		reqBody["search_depth"] = opts.SearchDepth
		reqBody["topic"] = "general"
		reqBody["max_results"] = opts.MaxResults
	case "deep_search":
		reqBody["search_depth"] = "advanced"
		reqBody["topic"] = "general"
		reqBody["max_results"] = opts.MaxResults
	case "search_last_24_hours":
		reqBody["search_depth"] = opts.SearchDepth
		reqBody["topic"] = "general"
		reqBody["days"] = 1
		reqBody["max_results"] = opts.MaxResults
	case "search_last_week":
		reqBody["search_depth"] = opts.SearchDepth
		reqBody["topic"] = "general"
		reqBody["days"] = 7
		reqBody["max_results"] = opts.MaxResults
	case "search_images":
		reqBody["search_depth"] = opts.SearchDepth
		reqBody["include_images"] = true
		reqBody["include_image_descriptions"] = true
		reqBody["max_results"] = opts.MaxResults
	case "search_by_date":
		reqBody["search_depth"] = "advanced"
		reqBody["topic"] = "general"
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
	if opts.Days > 0 && toolName != "search_last_24_hours" && toolName != "search_last_week" {
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

// parseWeChatResult parses WeChat search result string and converts to schema.SearchResult
func parseWeChatResult(wechatResult string) ([]schema.SearchResult, error) {
	var results []schema.SearchResult

	// Parse the formatted WeChat search result
	// Format example:
	// Found 5 articles for: nano banana pro
	//
	// 1. Title: 【免费】Nano Banana Pro 教程一:生成你的头像
	//    Publish Time: 2024-01-01
	//    URL: https://weixin.sogou.com/link?...
	//    Content: ...
	//
	// 2. Title: ...

	lines := strings.Split(wechatResult, "\n")
	var currentResult *schema.SearchResult

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Found") && strings.Contains(line, "articles for:") {
			// Skip the header line
			continue
		}
		if strings.HasPrefix(line, "No articles found") {
			// No results
			return []schema.SearchResult{}, nil
		}

		// Match article index: "1. Title: ..."
		if len(line) > 0 && line[0] >= '1' && line[0] <= '9' && strings.Contains(line, ". Title:") {
			// Save previous result if exists
			if currentResult != nil && currentResult.Title != "" {
				results = append(results, *currentResult)
			}
			// Start new result
			parts := strings.SplitN(line, ". Title: ", 2)
			currentResult = &schema.SearchResult{}
			if len(parts) == 2 {
				currentResult.Title = strings.TrimSpace(parts[1])
			}
		} else if currentResult != nil {
			if strings.HasPrefix(line, "Publish Time:") {
				currentResult.PublishedDate = strings.TrimSpace(strings.TrimPrefix(line, "Publish Time:"))
			} else if strings.HasPrefix(line, "URL:") {
				currentResult.URL = strings.TrimSpace(strings.TrimPrefix(line, "URL:"))
			} else if strings.HasPrefix(line, "Content:") {
				content := strings.TrimSpace(strings.TrimPrefix(line, "Content:"))
				currentResult.Content = content
				currentResult.RawContent = content
			}
		}
	}

	// Don't forget the last result
	if currentResult != nil && currentResult.Title != "" {
		results = append(results, *currentResult)
	}

	return results, nil
}
