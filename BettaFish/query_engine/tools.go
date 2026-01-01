package query_engine

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/smallnest/langgraphgo/showcases/BettaFish/schema"
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
	Images []string `json:"images"`
}

// ExecuteSearch executes a search using Tavily API.
func ExecuteSearch(ctx context.Context, query string, toolName string, startDate, endDate string) ([]schema.SearchResult, error) {
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
		reqBody["search_depth"] = "basic"
		reqBody["topic"] = "news"
		reqBody["max_results"] = 7
	case "deep_search_news":
		reqBody["search_depth"] = "advanced"
		reqBody["topic"] = "news"
		reqBody["max_results"] = 10
	case "search_news_last_24_hours":
		reqBody["search_depth"] = "basic"
		reqBody["topic"] = "news"
		reqBody["days"] = 1
		reqBody["max_results"] = 10
	case "search_news_last_week":
		reqBody["search_depth"] = "basic"
		reqBody["topic"] = "news"
		reqBody["days"] = 7
		reqBody["max_results"] = 10
	case "search_images_for_news":
		reqBody["search_depth"] = "basic"
		reqBody["include_images"] = true
		reqBody["max_results"] = 5
	case "search_news_by_date":
		reqBody["search_depth"] = "advanced" // Usually need advanced for historical
		// Tavily doesn't strictly support start/end date in the basic API payload in the same way as 'days',
		// but we can try to append it to the query or use specific fields if available.
		// For now, we'll append to query as a best effort if API doesn't support it directly.
		// Note: Tavily API has specific date params, but let's stick to simple implementation.
		reqBody["query"] = fmt.Sprintf("%s after:%s before:%s", query, startDate, endDate)
	default:
		reqBody["search_depth"] = "basic"
		reqBody["topic"] = "general"
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post("https://api.tavily.com/search", "application/json", strings.NewReader(string(jsonData)))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Tavily API error: %s", string(body))
	}

	var tavilyResp TavilyResponse
	if err := json.Unmarshal(body, &tavilyResp); err != nil {
		return nil, err
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
