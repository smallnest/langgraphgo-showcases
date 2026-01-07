package tool

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewWeChatSearch tests the constructor with default values and options.
func TestNewWeChatSearch(t *testing.T) {
	t.Run("default values", func(t *testing.T) {
		w, err := NewWeChatSearch()
		require.NoError(t, err)
		assert.Equal(t, 10, w.Count)
		assert.Contains(t, w.UserAgent, "Mozilla")
		assert.Contains(t, w.UserAgent, "Chrome")
	})

	t.Run("with options", func(t *testing.T) {
		w, err := NewWeChatSearch(
			WithWeChatCount(5),
			WithWeChatUserAgent("custom-ua/1.0"),
		)
		require.NoError(t, err)
		assert.Equal(t, 5, w.Count)
		assert.Equal(t, "custom-ua/1.0", w.UserAgent)
	})
}

// TestWeChatSearch_Name tests the Name method.
func TestWeChatSearch_Name(t *testing.T) {
	w, err := NewWeChatSearch()
	require.NoError(t, err)
	assert.Equal(t, "WeChat_Search", w.Name())
}

// TestWeChatSearch_Description tests the Description method.
func TestWeChatSearch_Description(t *testing.T) {
	w, err := NewWeChatSearch()
	require.NoError(t, err)
	desc := w.Description()
	assert.Contains(t, desc, "WeChat")
	assert.Contains(t, desc, "Sogou")
	assert.NotEmpty(t, desc)
}

// TestWithWeChatCount tests the count option with boundary conditions.
func TestWithWeChatCount(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected int
	}{
		{"zero becomes one", 0, 1},
		{"negative becomes one", -5, 1},
		{"one stays one", 1, 1},
		{"five stays five", 5, 5},
		{"ten stays ten", 10, 10},
		{"eleven becomes ten", 11, 10},
		{"large number becomes ten", 100, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, err := NewWeChatSearch(WithWeChatCount(tt.input))
			require.NoError(t, err)
			assert.Equal(t, tt.expected, w.Count)
		})
	}
}

// TestWithWeChatUserAgent tests the user agent option.
func TestWithWeChatUserAgent(t *testing.T) {
	w, err := NewWeChatSearch(WithWeChatUserAgent("test-ua/2.0"))
	require.NoError(t, err)
	assert.Equal(t, "test-ua/2.0", w.UserAgent)
}

// TestRealSearch_NanoBananaPro performs a real search for "nano banana pro" on Sogou WeChat.
// This is an integration test that makes actual HTTP requests to weixin.sogou.com.
func TestRealSearch_NanoBananaPro(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping real search test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	w, err := NewWeChatSearch(WithWeChatCount(5))
	require.NoError(t, err)

	// Perform the actual search
	result, err := w.Call(ctx, "nano banana pro")
	require.NoError(t, err, "Search should not return an error")

	// Verify we got some results
	assert.NotEmpty(t, result, "Result should not be empty")

	// Check for expected content in the result
	assert.Contains(t, result, "nano", "Result should contain 'nano'")
	assert.Contains(t, result, "banana", "Result should contain 'banana'")

	// Parse the result to verify structure
	lines := strings.Split(result, "\n")
	foundArticleCount := false
	foundTitle := false
	foundURL := false

	for _, line := range lines {
		if strings.Contains(line, "Found") && strings.Contains(line, "articles") {
			foundArticleCount = true
		}
		if strings.Contains(line, "Title:") {
			foundTitle = true
		}
		if strings.Contains(line, "URL:") {
			foundURL = true
			// Verify it's a valid URL format
			assert.True(t,
				strings.Contains(line, "http://") || strings.Contains(line, "https://"),
				"URL should contain http:// or https://")
		}
	}

	assert.True(t, foundArticleCount, "Should find article count in result")
	assert.True(t, foundTitle, "Should find at least one title")
	assert.True(t, foundURL, "Should find at least one URL")

	t.Logf("Search result:\n%s", result)
}

// TestRealSearch_SearchArticles tests the SearchArticles method with real HTTP requests.
func TestRealSearch_SearchArticles(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping real search test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	w, err := NewWeChatSearch(WithWeChatCount(3))
	require.NoError(t, err)

	articles, err := w.SearchArticles(ctx, "nano banana pro", 3)
	require.NoError(t, err, "SearchArticles should not return an error")

	// Verify we got some articles
	assert.NotEmpty(t, articles, "Should return at least one article")

	// Verify article structure
	for i, article := range articles {
		assert.NotEmpty(t, article.Title, fmt.Sprintf("Article %d should have a title", i))
		assert.NotEmpty(t, article.SogouURL, fmt.Sprintf("Article %d should have a Sogou URL", i))
		assert.True(t,
			strings.HasPrefix(article.SogouURL, "http://") || strings.HasPrefix(article.SogouURL, "https://"),
			fmt.Sprintf("Article %d SogouURL should be a valid URL", i))

		t.Logf("Article %d: %s", i, article.Title)
		t.Logf("  Sogou URL: %s", article.SogouURL)
		if article.RealURL != "" {
			t.Logf("  Real URL: %s", article.RealURL)
		}
		if article.PublishTime != "" {
			t.Logf("  Publish Time: %s", article.PublishTime)
		}
	}
}

// TestRealSearch_SogouWeixinSearch tests the sogouWeixinSearch method directly.
func TestRealSearch_SogouWeixinSearch(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping real search test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	w, err := NewWeChatSearch()
	require.NoError(t, err)

	results, err := w.sogouWeixinSearch(ctx, "nano banana pro")
	require.NoError(t, err, "sogouWeixinSearch should not return an error")

	// Verify we got some results
	assert.NotEmpty(t, results, "Should return at least one search result")

	// Verify result structure
	for i, result := range results {
		assert.NotEmpty(t, result.Title, fmt.Sprintf("Result %d should have a title", i))
		assert.NotEmpty(t, result.SogouURL, fmt.Sprintf("Result %d should have a URL", i))
		assert.Contains(t, result.SogouURL, "sogou", "URL should contain 'sogou'")

		t.Logf("Result %d: %s", i, result.Title)
		t.Logf("  URL: %s", result.SogouURL)
		if result.PublishTime != "" {
			t.Logf("  Publish Time: %s", result.PublishTime)
		}
	}
}

// TestSetDefaultHeaders tests the header setting logic.
func TestSetDefaultHeaders(t *testing.T) {
	tests := []struct {
		name     string
		referer  string
		expected map[string]string
	}{
		{
			name:    "without referer",
			referer: "",
			expected: map[string]string{
				"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
				"Accept-Language": "zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6",
				"Cache-Control":   "no-cache",
				"Connection":      "keep-alive",
				"Pragma":          "no-cache",
			},
		},
		{
			name:    "with referer",
			referer: "test query",
			expected: map[string]string{
				"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
				"Accept-Language": "zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6",
				"Cache-Control":   "no-cache",
				"Connection":      "keep-alive",
				"Pragma":          "no-cache",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, err := NewWeChatSearch()
			require.NoError(t, err)

			req, err := http.NewRequest("GET", "https://example.com", nil)
			require.NoError(t, err)

			w.setDefaultHeaders(req, tt.referer)

			for key, expectedValue := range tt.expected {
				actualValue := req.Header.Get(key)
				assert.Equal(t, expectedValue, actualValue, "Header %s should match", key)
			}

			assert.Equal(t, w.UserAgent, req.Header.Get("User-Agent"))

			if tt.referer != "" {
				referer := req.Header.Get("Referer")
				assert.Contains(t, referer, "weixin.sogou.com")
				assert.Contains(t, referer, "test+query")
			} else {
				assert.Empty(t, req.Header.Get("Referer"))
			}
		})
	}
}

// TestRealSearch_EmptyQuery tests search with a query that returns no results.
func TestRealSearch_EmptyQuery(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping real search test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	w, err := NewWeChatSearch()
	require.NoError(t, err)

	// Use a very unlikely query string
	result, err := w.Call(ctx, "xyzabc123def456ghi789jkl")
	require.NoError(t, err)

	// Empty results should still return a formatted message
	assert.NotEmpty(t, result)
	assert.Contains(t, result, "No articles found")
}

// TestRealSearch_ContextCancellation tests that the search respects context cancellation.
func TestRealSearch_ContextCancellation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping real search test in short mode")
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	w, err := NewWeChatSearch()
	require.NoError(t, err)

	_, err = w.Call(ctx, "nano banana pro")
	assert.Error(t, err, "Should return error when context is cancelled")
}

// TestRealSearch_MultipleQueries tests multiple different search queries.
func TestRealSearch_MultipleQueries(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping real search test in short mode")
	}

	queries := []string{"nano banana pro", "Raspberry Pi", "Arduino"}

	for _, query := range queries {
		t.Run(query, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			w, err := NewWeChatSearch(WithWeChatCount(2))
			require.NoError(t, err)

			result, err := w.Call(ctx, query)
			require.NoError(t, err)
			assert.NotEmpty(t, result)
		})
	}
}
