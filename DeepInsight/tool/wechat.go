package tool

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// WeChatSearch is a tool for searching WeChat official account articles via Sogou.
type WeChatSearch struct {
	Count     int
	UserAgent string
}

// WeChatOption is a function type for configuring WeChatSearch.
type WeChatOption func(*WeChatSearch)

// WithWeChatCount sets the number of articles to fetch (1-10).
func WithWeChatCount(count int) WeChatOption {
	return func(w *WeChatSearch) {
		if count < 1 {
			count = 1
		}
		if count > 10 {
			count = 10
		}
		w.Count = count
	}
}

// WithWeChatUserAgent sets a custom User-Agent header.
func WithWeChatUserAgent(ua string) WeChatOption {
	return func(w *WeChatSearch) {
		w.UserAgent = ua
	}
}

// NewWeChatSearch creates a new WeChatSearch tool.
func NewWeChatSearch(opts ...WeChatOption) (*WeChatSearch, error) {
	w := &WeChatSearch{
		Count:     10,
		UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/137.0.0.0 Safari/537.36 Edg/137.0.0.0",
	}

	for _, opt := range opts {
		opt(w)
	}

	return w, nil
}

// Name returns the name of the tool.
func (w *WeChatSearch) Name() string {
	return "WeChat_Search"
}

// Description returns the description of the tool.
func (w *WeChatSearch) Description() string {
	return "A search engine for WeChat official account articles via Sogou. " +
		"Useful for finding articles from WeChat official accounts. " +
		"Input should be a search query in Chinese or English."
}

// Article represents a WeChat article search result.
type Article struct {
	Title       string
	SogouURL    string
	RealURL     string
	PublishTime string
	Content     string
}

// Call executes the search and returns formatted article results.
func (w *WeChatSearch) Call(ctx context.Context, input string) (string, error) {
	articles, err := w.SearchArticles(ctx, input, w.Count)
	if err != nil {
		return "", fmt.Errorf("failed to search articles: %w", err)
	}

	if len(articles) == 0 {
		return fmt.Sprintf("No articles found for query: %s", input), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d articles for: %s\n\n", len(articles), input))

	for i, article := range articles {
		sb.WriteString(fmt.Sprintf("%d. Title: %s\n", i+1, article.Title))
		sb.WriteString(fmt.Sprintf("   Publish Time: %s\n", article.PublishTime))
		sb.WriteString(fmt.Sprintf("   URL: %s\n", article.RealURL))
		if article.Content != "" {
			// Truncate content if too long
			content := article.Content
			if len(content) > 500 {
				content = content[:500] + "..."
			}
			sb.WriteString(fmt.Sprintf("   Content: %s\n", content))
		}
		sb.WriteString("\n")
	}

	return sb.String(), nil
}

// SearchArticles searches WeChat articles via Sogou and returns the results.
func (w *WeChatSearch) SearchArticles(ctx context.Context, query string, count int) ([]Article, error) {
	// Step 1: Search articles via Sogou WeChat search
	searchResults, err := w.sogouWeixinSearch(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("sogou search failed: %w", err)
	}

	if len(searchResults) == 0 {
		return []Article{}, nil
	}

	// Limit results
	if count > len(searchResults) {
		count = len(searchResults)
	}
	searchResults = searchResults[:count]

	// Step 2: Get real URLs and content for each article
	articles := make([]Article, 0, count)
	for _, result := range searchResults {
		// Get real URL
		realURL, err := w.getRealURL(ctx, result.SogouURL)
		if err != nil {
			// Continue even if real URL extraction fails
			realURL = result.SogouURL
		}

		// Get article content
		var content string
		if realURL != "" {
			content, _ = w.getArticleContent(ctx, realURL, result.SogouURL)
		}

		articles = append(articles, Article{
			Title:       result.Title,
			SogouURL:    result.SogouURL,
			RealURL:     realURL,
			PublishTime: result.PublishTime,
			Content:     content,
		})
	}

	return articles, nil
}

// SogouSearchResult represents a search result from Sogou WeChat search.
type SogouSearchResult struct {
	Title       string
	SogouURL    string
	PublishTime string
}

// sogouWeixinSearch performs a search on Sogou WeChat search.
func (w *WeChatSearch) sogouWeixinSearch(ctx context.Context, query string) ([]SogouSearchResult, error) {
	baseURL := "https://weixin.sogou.com/weixin"
	params := url.Values{}
	params.Set("type", "2")
	params.Set("s_from", "input")
	params.Set("query", query)
	params.Set("ie", "utf8")
	params.Set("_sug_", "n")
	params.Set("_sug_type_", "")

	reqURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	w.setDefaultHeaders(req, query)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("sogou returned status: %d", resp.StatusCode)
	}

	// Parse HTML document
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Extract results using CSS selectors
	results := []SogouSearchResult{}

	// Find all article title links (id contains "sogou_vr_11002601_title_")
	doc.Find("a[id*='sogou_vr_11002601_title_']").Each(func(i int, s *goquery.Selection) {
		title := strings.TrimSpace(s.Text())
		href, exists := s.Attr("href")
		if !exists || href == "" {
			return
		}

		// Convert relative URL to absolute
		if !strings.HasPrefix(href, "http") {
			href = "https://weixin.sogou.com" + href
		}

		// Find publish time - it's in a sibling div
		publishTime := ""
		// Navigate up to find the containing box, then find the time element
		s.Parent().Parent().Find("span.s2").Each(func(j int, t *goquery.Selection) {
			publishTime = strings.TrimSpace(t.Text())
		})

		results = append(results, SogouSearchResult{
			Title:       title,
			SogouURL:    href,
			PublishTime: publishTime,
		})
	})

	return results, nil
}

// getRealURL extracts the real WeChat article URL from a Sogou redirect page.
func (w *WeChatSearch) getRealURL(ctx context.Context, sogouURL string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", sogouURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	w.setDefaultHeaders(req, "")

	// Sogou requires some cookies for redirect
	req.Header.Set("Cookie", "ABTEST=7|1750756616|v1; SUID=0A5BF4788E52A20B00000000685A6D08; IPLOC=CN1100")

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Don't follow redirects automatically
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	bodyStr := string(body)

	// Extract URL parts from JavaScript: url += 'part1'; url += 'part2'; ...
	pattern := regexp.MustCompile(`url\s*\+=\s*'([^']*)'`)
	matches := pattern.FindAllStringSubmatch(bodyStr, -1)

	var urlParts []string
	for _, match := range matches {
		if len(match) > 1 {
			urlParts = append(urlParts, match[1])
		}
	}

	if len(urlParts) == 0 {
		// Try alternative pattern: url = "part1" + "part2"
		pattern = regexp.MustCompile(`url\s*=\s*["']([^"']+)["']`)
		matches = pattern.FindAllStringSubmatch(bodyStr, -1)
		for _, match := range matches {
			if len(match) > 1 {
				urlParts = append(urlParts, match[1])
			}
		}
	}

	fullURL := strings.Join(urlParts, "")
	fullURL = strings.ReplaceAll(fullURL, "@", "")

	if fullURL != "" {
		return "https://mp." + fullURL, nil
	}

	// If JavaScript parsing fails, try to find URL in meta refresh or location
	metaPattern := regexp.MustCompile(`url\s*=\s*["']([^"']+)["']`)
	if match := metaPattern.FindStringSubmatch(bodyStr); len(match) > 1 {
		return match[1], nil
	}

	return "", fmt.Errorf("could not extract real URL")
}

// getArticleContent fetches the content of a WeChat article.
func (w *WeChatSearch) getArticleContent(ctx context.Context, realURL, referer string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", realURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	w.setDefaultHeaders(req, "")
	if referer != "" {
		req.Header.Set("Referer", referer)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("article returned status: %d", resp.StatusCode)
	}

	// Parse HTML and extract content
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Find content div
	content := ""
	doc.Find("#js_content").Each(func(i int, s *goquery.Selection) {
		content = s.Text()
	})

	if content == "" {
		return "", fmt.Errorf("content not found")
	}

	// Clean up whitespace
	lines := strings.Split(content, "\n")
	var cleanedLines []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			cleanedLines = append(cleanedLines, line)
		}
	}

	return strings.Join(cleanedLines, "\n"), nil
}

// setDefaultHeaders sets common HTTP headers for requests.
func (w *WeChatSearch) setDefaultHeaders(req *http.Request, referer string) {
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("User-Agent", w.UserAgent)
	if referer != "" {
		req.Header.Set("Referer", fmt.Sprintf("https://weixin.sogou.com/weixin?query=%s", url.QueryEscape(referer)))
	}
}
