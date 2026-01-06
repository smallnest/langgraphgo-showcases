package tool

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// ZhihuSearch is a tool for searching Zhihu questions and answers.
type ZhihuSearch struct {
	UserAgent string
	Timeout   time.Duration
}

// ZhihuOption is a function type for configuring ZhihuSearch.
type ZhihuOption func(*ZhihuSearch)

// WithZhihuUserAgent sets a custom User-Agent header.
func WithZhihuUserAgent(ua string) ZhihuOption {
	return func(z *ZhihuSearch) {
		z.UserAgent = ua
	}
}

// WithZhihuTimeout sets the request timeout.
func WithZhihuTimeout(timeout time.Duration) ZhihuOption {
	return func(z *ZhihuSearch) {
		z.Timeout = timeout
	}
}

// NewZhihuSearch creates a new ZhihuSearch tool.
func NewZhihuSearch(opts ...ZhihuOption) (*ZhihuSearch, error) {
	z := &ZhihuSearch{
		UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/137.0.0.0 Safari/537.36 Edg/137.0.0.0",
		Timeout:   30 * time.Second,
	}

	for _, opt := range opts {
		opt(z)
	}

	return z, nil
}

// Name returns the name of the tool.
func (z *ZhihuSearch) Name() string {
	return "Zhihu_Search"
}

// Description returns the description of the tool.
func (z *ZhihuSearch) Description() string {
	return "知乎搜索工具。可以搜索知乎上的问答、文章和专栏内容。" +
		"适用于查找中文社区的讨论、观点和专业见解。" +
		"输入可以是中文或英文的搜索关键词。"
}

// Call executes the search and returns formatted results.
func (z *ZhihuSearch) Call(ctx context.Context, input string) (string, error) {
	results, err := z.Search(ctx, input, 10)
	if err != nil {
		return "", fmt.Errorf("搜索失败: %w", err)
	}
	return FormatResults(results, input), nil
}

// Search implements the SearchTool interface.
func (z *ZhihuSearch) Search(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	// Create a context with timeout if not already set
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, z.Timeout)
		defer cancel()
	}

	// Try web search
	searchResults, err := z.zhihuWebSearch(ctx, query)
	if err != nil {
		// Return empty results on error instead of failing
		return []SearchResult{}, nil
	}

	if len(searchResults) == 0 {
		return []SearchResult{}, nil
	}

	// Limit results
	if limit > len(searchResults) {
		limit = len(searchResults)
	}
	searchResults = searchResults[:limit]

	// Convert to SearchResult format
	results := make([]SearchResult, 0, limit)
	for _, result := range searchResults {
		// Try to fetch content (but don't fail if it doesn't work)
		content, author := "", ""
		if strings.Contains(result.URL, "question") {
			content, author = z.getAnswerContent(ctx, result.URL)
		} else if strings.Contains(result.URL, "article") {
			content, author = z.getArticleContent(ctx, result.URL)
		}

		// Generate summary from content if not already available
		summary := result.Summary
		if summary == "" && content != "" {
			summary = z.generateSummary(content)
		}

		results = append(results, SearchResult{
			Title:       result.Title,
			URL:         result.URL,
			Author:      author,
			PublishTime: result.PublishTime,
			Summary:     summary,
			Content:     content,
			Source:      "知乎",
		})
	}

	return results, nil
}

// ZhihuSearchResult represents a search result from Zhihu.
type ZhihuSearchResult struct {
	Title       string
	URL         string
	Summary     string
	PublishTime string
}

// zhihuWebSearch performs a search on Zhihu web with fallback.
func (z *ZhihuSearch) zhihuWebSearch(ctx context.Context, query string) ([]ZhihuSearchResult, error) {
	// Try HTML search directly (more reliable than API)
	results, err := z.zhihuWebSearchHTML(ctx, query)
	if err == nil && len(results) > 0 {
		return results, nil
	}

	// If HTML search fails, return empty results (don't fail completely)
	return []ZhihuSearchResult{}, nil
}

// zhihuWebSearchHTML searches using HTML parsing.
func (z *ZhihuSearch) zhihuWebSearchHTML(ctx context.Context, query string) ([]ZhihuSearchResult, error) {
	baseURL := "https://www.zhihu.com/search"
	params := url.Values{}
	params.Set("type", "content")
	params.Set("q", query)

	reqURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return []ZhihuSearchResult{}, nil
	}

	z.setDefaultHeaders(req)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return []ZhihuSearchResult{}, nil
	}
	defer resp.Body.Close()

	// Don't fail on error status, just return empty results
	if resp.StatusCode != http.StatusOK {
		return []ZhihuSearchResult{}, nil
	}

	// Parse HTML document
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return []ZhihuSearchResult{}, nil
	}

	results := []ZhihuSearchResult{}

	// Try multiple selectors for robustness
	selectors := []string{
		"a[href*='/question/']",
		"a[href*='/answer/']",
		".Card-Main .Card-Title a",
	}

	for _, selector := range selectors {
		found := false
		doc.Find(selector).Each(func(i int, s *goquery.Selection) {
			if len(results) >= 10 {
				return
			}

			title := strings.TrimSpace(s.Text())
			href, exists := s.Attr("href")
			if !exists || href == "" || title == "" || len(title) < 5 {
				return
			}

			// Avoid duplicates
			for _, r := range results {
				if r.Title == title {
					return
				}
			}

			// Convert relative URL to absolute
			if !strings.HasPrefix(href, "http") {
				href = "https://www.zhihu.com" + href
			}

			results = append(results, ZhihuSearchResult{
				Title: title,
				URL:   href,
			})
			found = true
		})

		if found && len(results) > 0 {
			break
		}
	}

	return results, nil
}

// getAnswerContent fetches the content of a Zhihu answer.
func (z *ZhihuSearch) getAnswerContent(ctx context.Context, answerURL string) (content, author string) {
	req, err := http.NewRequestWithContext(ctx, "GET", answerURL, nil)
	if err != nil {
		return
	}

	z.setDefaultHeaders(req)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return
	}

	// Extract author name
	doc.Find(".AuthorInfo-name").First().Each(func(i int, s *goquery.Selection) {
		author = strings.TrimSpace(s.Text())
	})

	// Extract answer content
	doc.Find(".RichContent-inner").First().Each(func(i int, s *goquery.Selection) {
		content = strings.TrimSpace(s.Text())
	})

	// Clean up content
	content = z.cleanContent(content)

	return
}

// getArticleContent fetches the content of a Zhihu article.
func (z *ZhihuSearch) getArticleContent(ctx context.Context, articleURL string) (content, author string) {
	req, err := http.NewRequestWithContext(ctx, "GET", articleURL, nil)
	if err != nil {
		return
	}

	z.setDefaultHeaders(req)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return
	}

	// Extract author name
	doc.Find(".AuthorInfo-name, .Post-Author").First().Each(func(i int, s *goquery.Selection) {
		author = strings.TrimSpace(s.Text())
	})

	// Extract article content
	doc.Find(".Post-RichText, .RichContent-inner").First().Each(func(i int, s *goquery.Selection) {
		content = strings.TrimSpace(s.Text())
	})

	// Clean up content
	content = z.cleanContent(content)

	return
}

// generateSummary creates a summary from content.
func (z *ZhihuSearch) generateSummary(content string) string {
	// Remove extra whitespace
	content = strings.TrimSpace(content)
	if len(content) <= 300 {
		return content
	}

	// Try to end at a sentence boundary
	summary := content[:300]
	lastPeriod := strings.LastIndexAny(summary, "。！？.!?")
	if lastPeriod > 200 {
		summary = summary[:lastPeriod+1]
	} else {
		summary += "..."
	}

	return summary
}

// cleanContent removes extra whitespace and formats the content.
func (z *ZhihuSearch) cleanContent(content string) string {
	// Remove extra whitespace
	lines := strings.Split(content, "\n")
	var cleanedLines []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			cleanedLines = append(cleanedLines, line)
		}
	}

	content = strings.Join(cleanedLines, "\n")

	// Remove common unwanted patterns
	content = strings.ReplaceAll(content, "\t", " ")
	for strings.Contains(content, "  ") {
		content = strings.ReplaceAll(content, "  ", " ")
	}

	return content
}

// setDefaultHeaders sets common HTTP headers for requests.
func (z *ZhihuSearch) setDefaultHeaders(req *http.Request) {
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("User-Agent", z.UserAgent)
}
