package query_engine

import (
	"testing"
)

// TestGetDomainAuthority tests the domain authority scoring function
func TestGetDomainAuthority(t *testing.T) {
	tests := []struct {
		name     string
		domain   string
		minScore float64
		maxScore float64
	}{
		{"AI Google Official", "ai.google.dev", 90, 100},
		{"Google Official", "blog.google", 90, 100},
		{"OpenAI Official", "openai.com", 90, 100},
		{"Anthropic Official", "anthropic.com", 90, 100},
		{"NVIDIA Official", "nvidia.com", 90, 100},
		{"NVIDIA Developer", "developer.nvidia.com", 90, 100},
		{"Arxiv", "arxiv.org", 85, 95},
		{"GitHub", "github.com", 75, 90},
		{"CSDN", "csdn.net", 65, 80},
		{"Reddit", "reddit.com", 55, 75},
		{"Unknown Domain", "unknown-domain.com", 40, 60},
		{"Harvard (academic)", "harvard.edu", 80, 95},
		{"NIST (government)", "nist.gov", 85, 95},
		{"Banana Pi Wiki", "wiki.banana-pi.org", 75, 90},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := GetDomainAuthority(tt.domain)
			if score < tt.minScore || score > tt.maxScore {
				t.Errorf("GetDomainAuthority(%q) = %v, want [%v, %v]", tt.domain, score, tt.minScore, tt.maxScore)
			}
			t.Logf("Domain: %s, Score: %.0f", tt.domain, score)
		})
	}
}

// TestExtractDomain tests the domain extraction function
func TestExtractDomain(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{
			name:     "Simple URL",
			url:      "https://example.com/path",
			expected: "example.com",
		},
		{
			name:     "URL with subdomain",
			url:      "https://ai.google.dev/docs",
			expected: "ai.google.dev",
		},
		{
			name:     "URL with www",
			url:      "https://www.example.com/path",
			expected: "www.example.com",
		},
		{
			name:     "URL with port",
			url:      "https://example.com:8080/path",
			expected: "example.com",
		},
		{
			name:     "Invalid URL",
			url:      "not-a-url",
			expected: "",
		},
		{
			name:     "Complex URL",
			url:      "https://developer.nvidia.com/gpus/geforce-rtx-4090/specs",
			expected: "developer.nvidia.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractDomain(tt.url)
			if result != tt.expected {
				t.Errorf("extractDomain(%q) = %q, want %q", tt.url, result, tt.expected)
			}
		})
	}
}

// TestContextualDisambiguation tests the context-based disambiguation
func TestContextualDisambiguation(t *testing.T) {
	// Create mock search contexts
	searchContexts := []SearchContext{
		{
			Title:          "Google AI Model",
			Domain:         "ai.google.dev",
			Snippet:        "New AI model for image generation and multimodal tasks",
			AuthorityScore: 95,
		},
		{
			Title:          "Hardware Board",
			Domain:         "wiki.banana-pi.org",
			Snippet:        "Single board computer with GPIO pins and CPU specifications",
			AuthorityScore: 80,
		},
	}

	tests := []struct {
		name            string
		query           string
		expectedTopType string
		minTopScore     float64
	}{
		{
			name:            "AI Model Query",
			query:           "Nano Banana Pro AI image generation model",
			expectedTopType: "ai_model",
			minTopScore:     0.3,
		},
		{
			name:            "Hardware Query",
			query:           "Banana Pi board GPIO CPU hardware specifications",
			expectedTopType: "hardware",
			minTopScore:     0.3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scores := ContextualDisambiguation(tt.query, searchContexts)

			// Find the highest scoring type
			var topType string
			var topScore float64
			for entityType, score := range scores {
				if score > topScore {
					topScore = score
					topType = entityType
				}
			}

			t.Logf("Query: %s", tt.query)
			for entityType, score := range scores {
				t.Logf("  %s: %.3f", entityType, score)
			}

			if topType != tt.expectedTopType {
				t.Errorf("Expected top type %q, got %q", tt.expectedTopType, topType)
			}
			if topScore < tt.minTopScore {
				t.Errorf("Expected top score >= %.3f, got %.3f", tt.minTopScore, topScore)
			}
		})
	}
}

// TestEnhanceSearchResults tests the search results enhancement
func TestEnhanceSearchResults(t *testing.T) {
	// Import schema package
	// Note: This test would need schema.SearchResult to work properly
	// For now, we'll skip the actual implementation test

	// Mock results (would be []schema.SearchResult in real usage)
	mockResults := []struct {
		Title         string
		URL           string
		Content       string
		Score         float64
		PublishedDate string
	}{
		{
			Title:         "Google AI Documentation",
			URL:           "https://ai.google.dev/docs",
			Content:       "Official documentation for AI models",
			Score:         0.9,
			PublishedDate: "2025-11-20",
		},
		{
			Title:         "Blog Post",
			URL:           "https://medium.com/article",
			Content:       "Some blog content",
			Score:         0.7,
			PublishedDate: "2025-10-15",
		},
	}

	t.Log("TestEnhanceSearchResults would verify:")
	t.Log("  1. Domain extraction works correctly")
	t.Log("  2. Authority scores are assigned properly")
	t.Log("  3. Results are properly enhanced with metadata")

	for _, mr := range mockResults {
		domain := extractDomain(mr.URL)
		authority := GetDomainAuthority(domain)
		t.Logf("  URL: %s -> Domain: %s, Authority: %.0f", mr.URL, domain, authority)
	}
}

// BenchmarkGetDomainAuthority benchmarks the domain authority lookup
func BenchmarkGetDomainAuthority(b *testing.B) {
	domains := []string{
		"ai.google.dev",
		"github.com",
		"reddit.com",
		"unknown-domain.com",
	}

	for i := 0; i < b.N; i++ {
		for _, domain := range domains {
			GetDomainAuthority(domain)
		}
	}
}
