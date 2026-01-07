package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"
)

// GitHubSearch is a tool for searching GitHub repositories, code, and issues.
type GitHubSearch struct {
	UserAgent string
	Timeout   time.Duration
	Token     string // GitHub Personal Access Token (optional but recommended)
}

// GitHubOption is a function type for configuring GitHubSearch.
type GitHubOption func(*GitHubSearch)

// WithGitHubUserAgent sets a custom User-Agent header.
func WithGitHubUserAgent(ua string) GitHubOption {
	return func(g *GitHubSearch) {
		g.UserAgent = ua
	}
}

// WithGitHubTimeout sets the request timeout.
func WithGitHubTimeout(timeout time.Duration) GitHubOption {
	return func(g *GitHubSearch) {
		g.Timeout = timeout
	}
}

// WithGitHubToken sets the GitHub Personal Access Token.
func WithGitHubToken(token string) GitHubOption {
	return func(g *GitHubSearch) {
		g.Token = token
	}
}

// NewGitHubSearch creates a new GitHubSearch tool.
func NewGitHubSearch(opts ...GitHubOption) (*GitHubSearch, error) {
	g := &GitHubSearch{
		UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/137.0.0.0 Safari/537.36 Edg/137.0.0.0",
		Timeout:   30 * time.Second,
	}

	for _, opt := range opts {
		opt(g)
	}

	return g, nil
}

// Name returns the name of the tool.
func (g *GitHubSearch) Name() string {
	return "GitHub_Search"
}

// Description returns the description of the tool.
func (g *GitHubSearch) Description() string {
	return "GitHub 搜索工具。可以搜索 GitHub 上的代码仓库、代码片段、问题(Issues)和用户。" +
		"适用于查找开源项目、技术实现和代码示例。" +
		"支持中文和英文搜索关键词。"
}

// Call executes the search and returns formatted results.
func (g *GitHubSearch) Call(ctx context.Context, input string) (string, error) {
	results, err := g.Search(ctx, input, 10)
	if err != nil {
		return "", fmt.Errorf("搜索失败: %w", err)
	}
	return FormatResults(results, input), nil
}

// Search implements the SearchTool interface.
func (g *GitHubSearch) Search(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	// Create a context with timeout if not already set
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, g.Timeout)
		defer cancel()
	}

	// GitHub API has a rate limit of 30 requests per minute for authenticated requests
	// and 10 for unauthenticated. We'll search repositories and issues.

	// Step 1: Search repositories
	repos, err := g.searchRepositories(ctx, query, limit/2+1)
	if err != nil {
		return nil, fmt.Errorf("搜索仓库失败: %w", err)
	}

	// Step 2: Search issues (if we haven't hit the limit yet)
	var issues []SearchResult
	if len(repos) < limit {
		issues, err = g.searchIssues(ctx, query, limit-len(repos))
		if err != nil {
			// Continue even if issue search fails
			issues = []SearchResult{}
		}
	}

	// Combine results
	results := append(repos, issues...)

	// Limit total results
	if limit < len(results) {
		results = results[:limit]
	}

	return results, nil
}

// searchRepositories searches for GitHub repositories.
func (g *GitHubSearch) searchRepositories(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	// Properly URL encode the query
	apiURL := fmt.Sprintf("https://api.github.com/search/repositories?q=%s&sort=stars&order=desc&per_page=%d",
		url.QueryEscape(query), limit)

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	g.setDefaultHeaders(req)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求发送失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Read body for more info
		body := make([]byte, 200)
		resp.Body.Read(body)
		return nil, fmt.Errorf("GitHub API 返回状态码: %d, body: %s", resp.StatusCode, string(body))
	}

	var response struct {
		Items []struct {
			Name        string `json:"name"`
			FullName    string `json:"full_name"`
			Description string `json:"description"`
			HTMLURL     string `json:"html_url"`
			Language    string `json:"language"`
			Stars       int    `json:"stargazers_count"`
			Owner       struct {
				Login string `json:"login"`
			} `json:"owner"`
			UpdatedAt string `json:"updated_at"`
		} `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	results := make([]SearchResult, 0, len(response.Items))
	for _, item := range response.Items {
		// Build summary
		summary := item.Description
		if summary == "" {
			summary = "No description provided"
		}

		summary += fmt.Sprintf(" | Stars: %d", item.Stars)
		if item.Language != "" {
			summary += fmt.Sprintf(" | Language: %s", item.Language)
		}

		results = append(results, SearchResult{
			Title:       item.FullName,
			URL:         item.HTMLURL,
			Author:      item.Owner.Login,
			PublishTime: formatTime(item.UpdatedAt),
			Summary:     summary,
			Source:      "GitHub",
		})
	}

	return results, nil
}

// searchIssues searches for GitHub issues and pull requests.
func (g *GitHubSearch) searchIssues(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	apiURL := fmt.Sprintf("https://api.github.com/search/issues?q=%s&sort=updated&order=desc&per_page=%d",
		url.QueryEscape(query), limit)

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	g.setDefaultHeaders(req)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求发送失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Read body for more info
		body := make([]byte, 200)
		resp.Body.Read(body)
		return nil, fmt.Errorf("GitHub API 返回状态码: %d, body: %s", resp.StatusCode, string(body))
	}

	var response struct {
		Items []struct {
			Title    string `json:"title"`
			HTMLURL  string `json:"html_url"`
			Number   int    `json:"number"`
			User     struct {
				Login string `json:"login"`
			} `json:"user"`
			CreatedAt string `json:"created_at"`
			UpdatedAt string `json:"updated_at"`
			State     string `json:"state"`
			PullRequest struct {
				HTMLURL string `json:"html_url"`
			} `json:"pull_request"`
			Repository struct {
				FullName string `json:"full_name"`
			} `json:"repository"`
			Body string `json:"body"`
		} `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	results := make([]SearchResult, 0, len(response.Items))
	for _, item := range response.Items {
		// Determine if it's a PR or issue
		itemType := "Issue"
		if item.PullRequest.HTMLURL != "" {
			itemType = "Pull Request"
		}

		// Build summary
		summary := fmt.Sprintf("%s in %s | %s", itemType, item.Repository.FullName, item.State)
		if item.Body != "" {
			summary += fmt.Sprintf("\n%s", truncateText(item.Body, 200))
		}

		results = append(results, SearchResult{
			Title:       item.Title,
			URL:         item.HTMLURL,
			Author:      item.User.Login,
			PublishTime: formatTime(item.UpdatedAt),
			Summary:     summary,
			Content:     item.Body,
			Source:      "GitHub",
		})
	}

	return results, nil
}

// formatTime formats a time string in a more readable way.
func formatTime(timeStr string) string {
	// GitHub API returns time in RFC3339 format
	// We can parse and reformat it, but for now just return as-is
	t, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		return timeStr
	}
	return t.Format("2006-01-02")
}

// truncateText truncates text to a maximum length.
func truncateText(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen] + "..."
}

// setDefaultHeaders sets common HTTP headers for requests.
func (g *GitHubSearch) setDefaultHeaders(req *http.Request) {
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", g.UserAgent)

	// Get token - first from struct, then from environment variable
	token := g.Token
	if token == "" {
		token = os.Getenv("GITHUB_TOKEN")
	}
	if token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("token %s", token))
	}
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
}
