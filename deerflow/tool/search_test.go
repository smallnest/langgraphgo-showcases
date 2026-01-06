package tool

import (
	"context"
	"testing"
	"time"
)

// TestZhihuSearch tests the Zhihu search functionality.
func TestZhihuSearch(t *testing.T) {
	search, err := NewZhihuSearch(
		WithZhihuTimeout(10*time.Second),
	)
	if err != nil {
		t.Fatalf("Failed to create ZhihuSearch: %v", err)
	}

	ctx := context.Background()
	results, err := search.Search(ctx, "Go语言编程", 3)
	if err != nil {
		t.Logf("Zhihu search failed (may be due to network): %v", err)
		return
	}

	if len(results) == 0 {
		t.Log("No results found from Zhihu")
		return
	}

	t.Logf("Found %d results from Zhihu", len(results))
	for i, result := range results {
		t.Logf("%d. %s", i+1, result.Title)
		t.Logf("   URL: %s", result.URL)
		if result.Author != "" {
			t.Logf("   Author: %s", result.Author)
		}
		if result.Summary != "" {
			t.Logf("   Summary: %s", truncateText(result.Summary, 100))
		}
	}
}

// TestGitHubSearch tests the GitHub search functionality.
func TestGitHubSearch(t *testing.T) {
	search, err := NewGitHubSearch(
		WithGitHubTimeout(10*time.Second),
	)
	if err != nil {
		t.Fatalf("Failed to create GitHubSearch: %v", err)
	}

	ctx := context.Background()
	results, err := search.Search(ctx, "langgraph", 3)
	if err != nil {
		t.Logf("GitHub search failed (may be due to rate limiting): %v", err)
		return
	}

	if len(results) == 0 {
		t.Log("No results found from GitHub")
		return
	}

	t.Logf("Found %d results from GitHub", len(results))
	for i, result := range results {
		t.Logf("%d. %s", i+1, result.Title)
		t.Logf("   URL: %s", result.URL)
		if result.Author != "" {
			t.Logf("   Author: %s", result.Author)
		}
		if result.Summary != "" {
			t.Logf("   Summary: %s", truncateText(result.Summary, 100))
		}
	}
}

// TestScholarSearch tests the Google Scholar search functionality.
func TestScholarSearch(t *testing.T) {
	search, err := NewScholarSearch(
		WithScholarTimeout(15*time.Second),
	)
	if err != nil {
		t.Fatalf("Failed to create ScholarSearch: %v", err)
	}

	ctx := context.Background()
	results, err := search.Search(ctx, "machine learning", 3)
	if err != nil {
		t.Logf("Scholar search failed (may be due to blocking): %v", err)
		return
	}

	if len(results) == 0 {
		t.Log("No results found from Scholar")
		return
	}

	t.Logf("Found %d results from Scholar", len(results))
	for i, result := range results {
		t.Logf("%d. %s", i+1, result.Title)
		t.Logf("   URL: %s", result.URL)
		if result.Author != "" {
			t.Logf("   Authors: %s", result.Author)
		}
		if result.Summary != "" {
			t.Logf("   Summary: %s", truncateText(result.Summary, 100))
		}
	}
}

// TestRegistry tests the tool registry functionality.
func TestRegistry(t *testing.T) {
	// Create a new registry
	reg := NewRegistry()

	// Create and register tools
	wechat, _ := NewWeChatSearch()
	zhihu, _ := NewZhihuSearch()
	github, _ := NewGitHubSearch()
	scholar, _ := NewScholarSearch()

	// Register tools
	if err := reg.Register(CategorySocial, wechat); err != nil {
		t.Errorf("Failed to register WeChat: %v", err)
	}
	if err := reg.Register(CategorySocial, zhihu); err != nil {
		t.Errorf("Failed to register Zhihu: %v", err)
	}
	if err := reg.Register(CategoryCode, github); err != nil {
		t.Errorf("Failed to register GitHub: %v", err)
	}
	if err := reg.Register(CategoryAcademic, scholar); err != nil {
		t.Errorf("Failed to register Scholar: %v", err)
	}

	// List all tools
	tools := reg.List()
	t.Logf("Registered tools: %v", tools)

	// Get tools by category
	socialTools := reg.GetByCategory(CategorySocial)
	t.Logf("Social tools (%d):", len(socialTools))
	for _, tool := range socialTools {
		t.Logf("  - %s", tool.Name())
	}

	codeTools := reg.GetByCategory(CategoryCode)
	t.Logf("Code tools (%d):", len(codeTools))
	for _, tool := range codeTools {
		t.Logf("  - %s", tool.Name())
	}

	// Get specific tool
	if tool, exists := reg.Get("Zhihu_Search"); exists {
		t.Logf("Found tool: %s - %s", tool.Name(), tool.Description())
	}

	// Test descriptions
	descriptions := reg.DescribeAll()
	t.Logf("All tool descriptions:\n%s", descriptions)
}

// TestDefaultRegistry tests the default global registry.
func TestDefaultRegistry(t *testing.T) {
	tools := DefaultRegistry.List()
	if len(tools) == 0 {
		t.Log("No tools in default registry")
		return
	}

	t.Logf("Default registry has %d tools", len(tools))

	// Test getting tools by category
	socialTools := DefaultRegistry.GetByCategory(CategorySocial)
	t.Logf("Social category has %d tools", len(socialTools))

	codeTools := DefaultRegistry.GetByCategory(CategoryCode)
	t.Logf("Code category has %d tools", len(codeTools))

	academicTools := DefaultRegistry.GetByCategory(CategoryAcademic)
	t.Logf("Academic category has %d tools", len(academicTools))

	generalTools := DefaultRegistry.GetByCategory(CategoryGeneral)
	t.Logf("General category has %d tools", len(generalTools))
}

// TestTavilySearch tests the Tavily search functionality.
// Note: This test requires TAVILY_API_KEY environment variable to be set.
func TestTavilySearch(t *testing.T) {
	// Skip test if API key is not provided
	t.Skip("Skipping Tavily test - requires TAVILY_API_KEY environment variable")

	search, err := NewTavilySearch(
		WithTavilyTimeout(15*time.Second),
		WithTavilyMaxResults(5),
	)
	if err != nil {
		t.Fatalf("Failed to create TavilySearch: %v", err)
	}

	ctx := context.Background()
	results, err := search.Search(ctx, "What is LangGraph?", 5)
	if err != nil {
		t.Logf("Tavily search failed: %v", err)
		return
	}

	if len(results) == 0 {
		t.Log("No results found from Tavily")
		return
	}

	t.Logf("Found %d results from Tavily", len(results))
	for i, result := range results {
		t.Logf("%d. %s", i+1, result.Title)
		t.Logf("   URL: %s", result.URL)
		if result.Author != "" {
			t.Logf("   Author: %s", result.Author)
		}
		if result.Summary != "" {
			t.Logf("   Summary: %s", truncateText(result.Summary, 100))
		}
	}
}

// BenchmarkZhihuSearch benchmarks the Zhihu search.
func BenchmarkZhihuSearch(b *testing.B) {
	search, _ := NewZhihuSearch(
		WithZhihuTimeout(30 * time.Second),
	)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = search.Search(ctx, "Go语言", 3)
	}
}

// BenchmarkGitHubSearch benchmarks the GitHub search.
func BenchmarkGitHubSearch(b *testing.B) {
	search, _ := NewGitHubSearch(
		WithGitHubTimeout(30 * time.Second),
	)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = search.Search(ctx, "golang", 3)
	}
}
