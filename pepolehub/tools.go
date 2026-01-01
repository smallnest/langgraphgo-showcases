package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/tmc/langchaingo/llms/openai"
)

// Real External Services

// LinkedInData, SearchResult, ScrapedContent, WebSummary are defined in types.go

func fetchLinkedInProfile(url string) (*LinkedInData, error) {
	// Attempt to fetch via Tavily (Search for profile content)
	// Since direct scraping is hard, we search for the specific URL to see if Tavily has indexed it.
	
tavily, err := NewTavilyClient()
	if err != nil {
		return nil, err
	}

	// Search specifically for this URL to get cached content
	results, err := tavily.Search(context.Background(), url, 1)
	if err != nil {
		return nil, fmt.Errorf("failed to search linkedin: %w", err)
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("profile not found in search index")
	}

	res := results[0]
	
	// Create a "best effort" profile object
	return &LinkedInData{
		LinkedinUrl:    url,
		LinkedinId:     "unknown", // Cannot extract easily
		FirstName:      "Unknown", // Needs parsing
		LastName:       "Unknown", // Needs parsing
		FullName:       res.Title, // Use Page Title
		Headline:       "Extracted from Web Search",
		About:          res.Content, // Use the snippet/content from search
		Location:       "Unknown",
		Connections:    0,
	},
	nil
}

func generateSearchQuery(personName, linkedinUrl string) (string, error) {
	// Use LLM to generate query
	llm, err := openai.New()
	if err != nil {
		return "", fmt.Errorf("failed to init llm: %w", err)
	}

	prompt := fmt.Sprintf("Generate a Google search query to find information about %s (LinkedIn: %s). Return ONLY the query string, nothing else.", personName, linkedinUrl)
	
	resp, err := llm.Call(context.Background(), prompt)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(resp), nil
}

type PersonSearchOptions struct {
	MaxResults int
	Context    string
}

func searchGoogleForPerson(personName, linkedinUrl string, options PersonSearchOptions) ([]SearchResult, error) {
	tavily, err := NewTavilyClient()
	if err != nil {
		return nil, err
	}

	// Construct query
	query := fmt.Sprintf("%s %s", personName, options.Context)
	if options.Context == "" {
		query = fmt.Sprintf("%s %s", personName, linkedinUrl)
	}

	results, err := tavily.Search(context.Background(), query, options.MaxResults)
	if err != nil {
		return nil, err
	}

	var searchResults []SearchResult
	for i, r := range results {
		searchResults = append(searchResults, SearchResult{
			Title:  r.Title,
			Url:    r.URL,
			Snippet: r.Content,
			Rank:   i + 1,
			Source: "tavily",
		})
	}

	return searchResults, nil
}

type ScrapeOptions struct {}

func scrapeUrls(urls []string, options *ScrapeOptions) ([]ScrapedContent, error) {
	var results []ScrapedContent
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	for _, u := range urls {
		// Basic http get
		req, err := http.NewRequest("GET", u, nil)
		if err != nil {
			continue
		}
		// Mock User-Agent to avoid some blocks
		req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.114 Safari/537.36")

		resp, err := client.Do(req)
		if err != nil {
			// If failed, we might skip or record error
			results = append(results, ScrapedContent{
				Url:       u,
				Error:     err.Error(),
				FetchedAt: time.Now().Unix(),
				Status:    0,
			})
			continue
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		content := string(body)

		// Very basic HTML text extraction (stripping tags would be better, but keeping raw for now or minimal cleanup)
		// For true implementation, we'd use goquery to extract text.
		// Let's assume we just store the body, or a truncated version.
		if len(content) > 50000 {
			content = content[:50000] // Truncate
		}

		results = append(results, ScrapedContent{
			Url:       u,
			Content:   content,
			Bytes:     len(content),
			FetchedAt: time.Now().Unix(),
			Status:    resp.StatusCode,
		})
	}
	return results, nil
}

func summarizeWebContent(url, content, personName string) (*WebSummary, error) {
	llm, err := openai.New()
	if err != nil {
		return nil, err
	}

	// Truncate content for LLM context window
	if len(content) > 10000 {
		content = content[:10000]
	}

	prompt := fmt.Sprintf(`Summarize the following web content specifically regarding "%s".
Content from %s:
%s

Output format:
Summary: <text>
KeyPoints: <point1>, <point2>, ...
MentionsPerson: <true/false>
Sentiment: <positive/neutral/negative>
`, personName, url, content)

	resp, err := llm.Call(context.Background(), prompt)
	if err != nil {
		return nil, err
	}

	// Simple parsing of the response (Robust parsing would be better)
	// Expecting the LLM to follow format somewhat.
	
	summary := "Generated summary"
	keyPoints := []string{}
	mentions := false
	sentiment := "neutral"

	lines := strings.Split(resp, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Summary:") {
			summary = strings.TrimPrefix(line, "Summary:")
		} else if strings.HasPrefix(line, "KeyPoints:") {
			kp := strings.TrimPrefix(line, "KeyPoints:")
			keyPoints = strings.Split(kp, ",")
			for i := range keyPoints {
				keyPoints[i] = strings.TrimSpace(keyPoints[i])
			}
		} else if strings.HasPrefix(line, "MentionsPerson:") {
			val := strings.TrimPrefix(line, "MentionsPerson:")
			if strings.ToLower(strings.TrimSpace(val)) == "true" {
				mentions = true
			}
		} else if strings.HasPrefix(line, "Sentiment:") {
			sentiment = strings.TrimSpace(strings.TrimPrefix(line, "Sentiment:"))
		}
	}

	return &WebSummary{
		Url:            url,
		Summary:        summary,
		KeyPoints:      keyPoints,
		MentionsPerson: mentions,
		Confidence:     0.8,
		Sentiment:      sentiment,
		Source:         "web",
	},
	nil
}

type ResearchDataBundle struct {
	PersonName    string
	LinkedinUrl   string
	LinkedinData  *LinkedInData
	WebSummaries  []WebSummary
	SearchResults []SearchResult
	Metadata      map[string]interface{}
}

type ResearchReportResult struct {
	Report string
}

func generateResearchReport(bundle ResearchDataBundle) (*ResearchReportResult, error) {
	llm, err := openai.New()
	if err != nil {
		return nil, err
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Research data for %s (%s):\n\n", bundle.PersonName, bundle.LinkedinUrl))
	
	if bundle.LinkedinData != nil {
		sb.WriteString(fmt.Sprintf("LinkedIn Info: %s - %s\n%s\n\n", bundle.LinkedinData.FullName, bundle.LinkedinData.Headline, bundle.LinkedinData.About))
	}
	
	sb.WriteString("Web Findings:\n")
	for _, s := range bundle.WebSummaries {
		sb.WriteString(fmt.Sprintf("- URL: %s\n  Summary: %s\n  Points: %v\n\n", s.Url, s.Summary, s.KeyPoints))
	}

	prompt := fmt.Sprintf(`Write a comprehensive research report in Markdown format about %s based on the following data:

%s

The report should be professional, structured, and highlight key career achievements, roles, and online presence.
`, bundle.PersonName, sb.String())

	resp, err := llm.Call(context.Background(), prompt)
	if err != nil {
		return nil, err
	}

	return &ResearchReportResult{Report: resp}, nil
}

const MAX_SEARCH_RESULTS = 5
