package tool

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// ScholarSearch is a tool for searching academic papers via Google Scholar.
type ScholarSearch struct {
	UserAgent string
	Timeout   time.Duration
}

// ScholarOption is a function type for configuring ScholarSearch.
type ScholarOption func(*ScholarSearch)

// WithScholarUserAgent sets a custom User-Agent header.
func WithScholarUserAgent(ua string) ScholarOption {
	return func(s *ScholarSearch) {
		s.UserAgent = ua
	}
}

// WithScholarTimeout sets the request timeout.
func WithScholarTimeout(timeout time.Duration) ScholarOption {
	return func(s *ScholarSearch) {
		s.Timeout = timeout
	}
}

// NewScholarSearch creates a new ScholarSearch tool.
func NewScholarSearch(opts ...ScholarOption) (*ScholarSearch, error) {
	s := &ScholarSearch{
		UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/137.0.0.0 Safari/537.36 Edg/137.0.0.0",
		Timeout:   30 * time.Second,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s, nil
}

// Name returns the name of the tool.
func (s *ScholarSearch) Name() string {
	return "Scholar_Search"
}

// Description returns the description of the tool.
func (s *ScholarSearch) Description() string {
	return "Google Scholar 学术搜索工具。可以搜索学术论文、研究报告、会议论文和学术著作。" +
		"适用于查找学术研究、理论依据和科学文献。" +
		"支持中英文搜索关键词。"
}

// Call executes the search and returns formatted results.
func (s *ScholarSearch) Call(ctx context.Context, input string) (string, error) {
	results, err := s.Search(ctx, input, 10)
	if err != nil {
		return "", fmt.Errorf("搜索失败: %w", err)
	}
	return FormatResults(results, input), nil
}

// Search implements the SearchTool interface.
func (s *ScholarSearch) Search(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	// Create a context with timeout if not already set
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, s.Timeout)
		defer cancel()
	}

	// Search via Google Scholar
	papers, err := s.scholarSearch(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("学术搜索失败: %w", err)
	}

	if len(papers) == 0 {
		return []SearchResult{}, nil
	}

	// Limit results
	if limit > len(papers) {
		limit = len(papers)
	}
	papers = papers[:limit]

	// Convert to SearchResult format
	results := make([]SearchResult, 0, limit)
	for _, paper := range papers {
		results = append(results, SearchResult{
			Title:       paper.Title,
			URL:         paper.URL,
			Author:      paper.Authors,
			PublishTime: paper.Year,
			Summary:     paper.Abstract,
			Content:     paper.Abstract,
			Source:      "Google Scholar",
		})
	}

	return results, nil
}

// ScholarPaper represents a paper from Google Scholar.
type ScholarPaper struct {
	Title     string
	URL       string
	Authors   string
	Year      string
	Abstract  string
	Source    string
	Citations int
}

// scholarSearch performs a search on Google Scholar.
func (s *ScholarSearch) scholarSearch(ctx context.Context, query string) ([]ScholarPaper, error) {
	baseURL := "https://scholar.google.com/scholar"
	params := url.Values{}
	params.Set("q", query)
	params.Set("hl", "zh-CN") // Use Chinese interface
	params.Set("start", "0")
	params.Set("num", "20")

	reqURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	s.setDefaultHeaders(req)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求发送失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Google Scholar may block automated requests
		if resp.StatusCode == 429 {
			return nil, fmt.Errorf("请求过于频繁，被 Google Scholar 限制访问")
		}
		return nil, fmt.Errorf("Google Scholar 返回状态码: %d", resp.StatusCode)
	}

	// Parse HTML document
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("解析 HTML 失败: %w", err)
	}

	papers := []ScholarPaper{}

	// Google Scholar results are in div.gs_r (or .gs_ri for newer versions)
	doc.Find(".gs_r, .gs_ri").Each(func(i int, sel *goquery.Selection) {
		paper := ScholarPaper{}

		// Extract title and URL
		titleLink := sel.Find(".gs_rt a").First()
		if titleLink.Length() > 0 {
			paper.Title = strings.TrimSpace(titleLink.Text())
			if href, exists := titleLink.Attr("href"); exists {
				paper.URL = href
			}
		} else {
			// Try alternative selector
			h3 := sel.Find("h3").First()
			if h3.Length() > 0 {
				paper.Title = strings.TrimSpace(h3.Text())
				// Try to find URL in h3
				if a := h3.Find("a").First(); a.Length() > 0 {
					if href, exists := a.Attr("href"); exists {
						paper.URL = href
					}
				}
			}
		}

		// If still no URL, try to get it from citation link
		if paper.URL == "" {
			if citeLink := sel.Find("a[href*='scholar?cites']").First(); citeLink.Length() > 0 {
				if href, exists := citeLink.Attr("href"); exists {
					// This is a citation link, not the actual paper URL
					// We'll use it as fallback
					if strings.HasPrefix(href, "/") {
						paper.URL = "https://scholar.google.com" + href
					} else {
						paper.URL = href
					}
				}
			}
		}

		// Extract authors and publication info
		sel.Find(".gs_a").Each(func(j int, s *goquery.Selection) {
			info := strings.TrimSpace(s.Text())
			// Format is usually: "Authors - Publication - Year - Source"
			parts := strings.Split(info, " - ")
			if len(parts) > 0 {
				paper.Authors = parts[0]
			}
			if len(parts) > 1 {
				// Extract year from publication info
				yearPattern := regexp.MustCompile(`\b(19|20)\d{2}\b`)
				if match := yearPattern.FindString(info); match != "" {
					paper.Year = match
				}
			}
		})

		// Extract abstract/snippet
		sel.Find(".gs_rs").Each(func(j int, s *goquery.Selection) {
			paper.Abstract = strings.TrimSpace(s.Text())
		})

		// Extract citation count
		sel.Find(".gs_fl a[href*='cites']").Each(func(j int, s *goquery.Selection) {
			text := strings.TrimSpace(s.Text())
			// Format is usually "被引用次数"
			citePattern := regexp.MustCompile(`\d+`)
			if match := citePattern.FindString(text); match != "" {
				fmt.Sscanf(match, "%d", &paper.Citations)
			}
		})

		// Only add if we have at least a title
		if paper.Title != "" && paper.Title != "[引用]" && paper.Title != "[PDF]" {
			papers = append(papers, paper)
		}
	})

	// If no results found, try alternative parsing
	if len(papers) == 0 {
		doc.Find("div[class*='gs_']").Each(func(i int, sel *goquery.Selection) {
			paper := ScholarPaper{}

			// Try to find any links
			if a := sel.Find("a").First(); a.Length() > 0 {
				paper.Title = strings.TrimSpace(a.Text())
				if href, exists := a.Attr("href"); exists {
					paper.URL = href
				}
			}

			if paper.Title != "" {
				// Get surrounding text as abstract
				text := sel.Text()
				if len(text) > 100 {
					abstractLen := 300
					if len(text) < abstractLen {
						abstractLen = len(text)
					}
					paper.Abstract = text[:abstractLen] + "..."
				}

				papers = append(papers, paper)
			}
		})
	}

	return papers, nil
}

// setDefaultHeaders sets common HTTP headers for requests.
func (s *ScholarSearch) setDefaultHeaders(req *http.Request) {
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("User-Agent", s.UserAgent)

	// Add referer to look more like a real browser
	req.Header.Set("Referer", "https://scholar.google.com/")
}
