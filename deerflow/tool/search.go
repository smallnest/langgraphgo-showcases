package tool

import (
	"context"
	"fmt"
)

// SearchResult represents a unified search result across different sources.
type SearchResult struct {
	Title       string // 标题
	URL         string // 链接
	Author      string // 作者/发布者
	PublishTime string // 发布时间
	Summary     string // 内容摘要
	Content     string // 完整内容（可选）
	Source      string // 来源平台
}

// SearchTool is the unified interface for all search tools.
type SearchTool interface {
	// Name returns the tool name
	Name() string

	// Description returns what this tool does
	Description() string

	// Call executes the search with the given input query
	// Returns formatted results as a string
	Call(ctx context.Context, input string) (string, error)

	// Search performs the actual search and returns structured results
	// This can be used when you need programmatic access to the results
	Search(ctx context.Context, query string, limit int) ([]SearchResult, error)
}

// FormatResults formats search results into a readable string.
// It limits the output size to avoid exceeding token limits.
func FormatResults(results []SearchResult, query string) string {
	if len(results) == 0 {
		return fmt.Sprintf("未找到关于「%s」的相关结果", query)
	}

	var output string
	output += fmt.Sprintf("## 找到 %d 条关于「%s」的结果\n\n", len(results), query)

	for i, result := range results {
		output += fmt.Sprintf("### %d. %s\n", i+1, result.Title)
		if result.Author != "" {
			output += fmt.Sprintf("**作者**: %s  ", result.Author)
		}
		if result.PublishTime != "" {
			output += fmt.Sprintf("**时间**: %s", result.PublishTime)
		}
		output += "\n\n"
		output += fmt.Sprintf("**链接**: %s\n\n", result.URL)

		if result.Summary != "" {
			summary := result.Summary
			// Limit summary to 5000 characters to prevent excessive output
			if len(summary) > 5000 {
				summary = summary[:5000] + "... (摘要过长，已截断)"
			}
			output += fmt.Sprintf("**摘要**: %s\n\n", summary)
		}

		if result.Content != "" {
			content := result.Content
			if len(content) > 2000 {
				content = content[:2000] + "..."
			}
			output += fmt.Sprintf("**内容**: %s\n\n", content)
		}
		output += "---\n\n"
	}

	return output
}

// BaseSearchTool provides common functionality for search tools.
type BaseSearchTool struct {
	ToolName        string
	ToolDescription string
	UserAgent       string
}

// SetUserAgent sets a custom user agent.
func (b *BaseSearchTool) SetUserAgent(ua string) {
	b.UserAgent = ua
}

// GetDefaultUserAgent returns the default user agent.
func (b *BaseSearchTool) GetDefaultUserAgent() string {
	if b.UserAgent == "" {
		return "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/137.0.0.0 Safari/537.36 Edg/137.0.0.0"
	}
	return b.UserAgent
}
