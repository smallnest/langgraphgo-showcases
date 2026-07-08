package query_engine

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/smallnest/langgraphgo-showcases/DeepInsight/schema"
	"github.com/tmc/langchaingo/llms"
)

// DomainAuthority represents the authority score of a domain
type DomainAuthority struct {
	Domain     string
	BaseScore  float64  // 0-100
	IsOfficial bool     // 是否为官方域名
	Categories []string // academic, tech, government, news, etc.
	UpdatedAt  time.Time
}

// AuthoritativeDomains contains the pre-defined domain authority rankings
var AuthoritativeDomains = []DomainAuthority{
	// AI & Tech Official Domains
	{Domain: "ai.google.dev", BaseScore: 95, IsOfficial: true, Categories: []string{"tech", "ai", "official"}},
	{Domain: "blog.google", BaseScore: 92, IsOfficial: true, Categories: []string{"tech", "ai", "official"}},
	{Domain: "google.com", BaseScore: 90, IsOfficial: true, Categories: []string{"tech", "official"}},
	{Domain: "openai.com", BaseScore: 95, IsOfficial: true, Categories: []string{"tech", "ai", "official"}},
	{Domain: "anthropic.com", BaseScore: 95, IsOfficial: true, Categories: []string{"tech", "ai", "official"}},
	{Domain: "meta.com", BaseScore: 90, IsOfficial: true, Categories: []string{"tech", "ai", "official"}},
	{Domain: "nvidia.com", BaseScore: 92, IsOfficial: true, Categories: []string{"tech", "hardware", "official"}},
	{Domain: "developer.nvidia.com", BaseScore: 94, IsOfficial: true, Categories: []string{"tech", "developer", "official"}},
	{Domain: "docs.anthropic.com", BaseScore: 96, IsOfficial: true, Categories: []string{"tech", "docs", "official"}},
	{Domain: "platform.openai.com", BaseScore: 96, IsOfficial: true, Categories: []string{"tech", "docs", "official"}},

	// Academic Domains
	{Domain: "arxiv.org", BaseScore: 93, IsOfficial: false, Categories: []string{"academic", "research"}},
	{Domain: "scholar.google.com", BaseScore: 90, IsOfficial: true, Categories: []string{"academic", "research"}},
	{Domain: "ieee.org", BaseScore: 88, IsOfficial: true, Categories: []string{"academic", "tech"}},
	{Domain: "acm.org", BaseScore: 88, IsOfficial: true, Categories: []string{"academic", "tech"}},
	{Domain: "springer.com", BaseScore: 85, IsOfficial: true, Categories: []string{"academic"}},
	{Domain: "sciencedirect.com", BaseScore: 85, IsOfficial: true, Categories: []string{"academic"}},
	{Domain: "nature.com", BaseScore: 90, IsOfficial: true, Categories: []string{"academic", "science"}},
	{Domain: "science.org", BaseScore: 90, IsOfficial: true, Categories: []string{"academic", "science"}},

	// Tech Communities & Documentation
	{Domain: "github.com", BaseScore: 82, IsOfficial: false, Categories: []string{"tech", "code", "community"}},
	{Domain: "stackoverflow.com", BaseScore: 80, IsOfficial: false, Categories: []string{"tech", "qna", "community"}},
	{Domain: "dev.to", BaseScore: 75, IsOfficial: false, Categories: []string{"tech", "blog", "community"}},
	{Domain: "medium.com", BaseScore: 70, IsOfficial: false, Categories: []string{"tech", "blog"}},
	{Domain: "hashnode.com", BaseScore: 72, IsOfficial: false, Categories: []string{"tech", "blog", "community"}},

	// Tech News & Media
	{Domain: "techcrunch.com", BaseScore: 78, IsOfficial: false, Categories: []string{"tech", "news"}},
	{Domain: "theverge.com", BaseScore: 75, IsOfficial: false, Categories: []string{"tech", "news"}},
	{Domain: "wired.com", BaseScore: 77, IsOfficial: false, Categories: []string{"tech", "news"}},
	{Domain: "arstechnica.com", BaseScore: 80, IsOfficial: false, Categories: []string{"tech", "news"}},
	{Domain: "informationweek.com", BaseScore: 73, IsOfficial: false, Categories: []string{"tech", "news"}},

	// Government & Standards
	{Domain: "nist.gov", BaseScore: 92, IsOfficial: true, Categories: []string{"government", "standards"}},
	{Domain: "ietf.org", BaseScore: 90, IsOfficial: true, Categories: []string{"standards", "tech"}},
	{Domain: "w3.org", BaseScore: 90, IsOfficial: true, Categories: []string{"standards", "tech"}},

	// Chinese Tech Domains
	{Domain: "csdn.net", BaseScore: 72, IsOfficial: false, Categories: []string{"tech", "blog", "chinese"}},
	{Domain: "juejin.cn", BaseScore: 74, IsOfficial: false, Categories: []string{"tech", "blog", "chinese"}},
	{Domain: "zhihu.com", BaseScore: 70, IsOfficial: false, Categories: []string{"qna", "chinese"}},
	{Domain: "segmentfault.com", BaseScore: 71, IsOfficial: false, Categories: []string{"tech", "qna", "chinese"}},
	{Domain: "bilibili.com", BaseScore: 68, IsOfficial: false, Categories: []string{"video", "chinese"}},

	// Hardware Vendors
	{Domain: "banana-pi.org", BaseScore: 82, IsOfficial: true, Categories: []string{"hardware", "vendor"}},
	{Domain: "wiki.banana-pi.org", BaseScore: 84, IsOfficial: true, Categories: []string{"hardware", "docs", "vendor"}},
	{Domain: "raspberrypi.com", BaseScore: 85, IsOfficial: true, Categories: []string{"hardware", "vendor"}},
	{Domain: "raspberrypi.org", BaseScore: 85, IsOfficial: true, Categories: []string{"hardware", "vendor"}},
	{Domain: "arduino.cc", BaseScore: 83, IsOfficial: true, Categories: []string{"hardware", "vendor"}},
	{Domain: "esp32.com", BaseScore: 82, IsOfficial: true, Categories: []string{"hardware", "vendor"}},

	// Default/Low Authority Domains
	{Domain: "wikipedia.org", BaseScore: 75, IsOfficial: false, Categories: []string{"encyclopedia"}},
	{Domain: "reddit.com", BaseScore: 65, IsOfficial: false, Categories: []string{"community", "forum"}},
}

// EntityArbiterResult represents the result of entity disambiguation
type EntityArbiterResult struct {
	EntityName      string    `json:"entity_name"`      // 确认的实体名称
	EntityType      string    `json:"entity_type"`      // 实体类型: ai_model, hardware, software, concept, etc.
	Confidence      float64   `json:"confidence"`       // 置信度 0-1
	PrimaryDomain   string    `json:"primary_domain"`   // 权威来源域名
	Reasoning       string    `json:"reasoning"`        // 判断理由
	EvidenceSources []string  `json:"evidence_sources"` // 证据来源
	IsVerified      bool      `json:"is_verified"`      // 是否已验证
	Timestamp       time.Time `json:"timestamp"`
}

// SearchContext represents enhanced search result with metadata
type SearchContext struct {
	Title          string  `json:"title"`
	URL            string  `json:"url"`
	Domain         string  `json:"domain"`
	Snippet        string  `json:"snippet"`
	PublishDate    string  `json:"publish_date"`
	Score          float64 `json:"score"`
	AuthorityScore float64 `json:"authority_score"`
}

// GetDomainAuthority calculates the authority score for a given domain
func GetDomainAuthority(domain string) float64 {
	domain = strings.ToLower(domain)

	// Remove www. prefix
	domain = strings.TrimPrefix(domain, "www.")

	// Extract base domain (handle subdomains)
	parts := strings.Split(domain, ".")
	if len(parts) >= 2 {
		// Check for exact match first
		for _, da := range AuthoritativeDomains {
			if domain == da.Domain {
				return calculateTimeDecay(da)
			}
		}

		// Check for base domain match (e.g., ai.google.dev matches google.dev)
		baseDomain := strings.Join(parts[len(parts)-2:], ".")
		for _, da := range AuthoritativeDomains {
			if baseDomain == da.Domain || strings.HasSuffix(domain, "."+da.Domain) {
				return calculateTimeDecay(da)
			}
		}
	}

	// Default authority score for unknown domains
	// Check if it's an academic domain (.edu, .ac.*, etc.)
	if isAcademicDomain(domain) {
		return 85.0
	}

	// Check if it's a government domain (.gov)
	if isGovernmentDomain(domain) {
		return 88.0
	}

	return 50.0 // Default score for unknown domains
}

// calculateTimeDecay applies time-based decay to domain authority
// This ensures newer domains aren't unfairly penalized
func calculateTimeDecay(da DomainAuthority) float64 {
	// For now, return base score
	// In production, you might want to adjust based on how recently the domain was updated
	return da.BaseScore
}

// isAcademicDomain checks if a domain is academic
func isAcademicDomain(domain string) bool {
	// Check for common academic TLDs
	academicTLDs := []string{".edu", ".ac.", ".edu."}
	for _, tld := range academicTLDs {
		if strings.Contains(domain, tld) {
			return true
		}
	}
	return false
}

// isGovernmentDomain checks if a domain is government
func isGovernmentDomain(domain string) bool {
	// Check for government TLDs
	govTLDs := []string{".gov", ".gov.", ".go.", ".gov.uk", ".gov.cn"}
	for _, tld := range govTLDs {
		if strings.Contains(domain, tld) {
			return true
		}
	}
	return false
}

// extractDomain extracts the domain from a URL
func extractDomain(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		// Try to extract domain from invalid URL
		if strings.Contains(rawURL, "://") {
			parts := strings.Split(rawURL, "://")
			if len(parts) > 1 {
				hostPart := strings.Split(parts[1], "/")[0]
				return hostPart
			}
		}
		return ""
	}

	host := u.Hostname()
	if host == "" {
		return ""
	}

	return host
}

// EnhanceSearchResults adds authority scores and domain information
func EnhanceSearchResults(results []schema.SearchResult) []SearchContext {
	enhanced := make([]SearchContext, 0, len(results))

	for _, r := range results {
		domain := extractDomain(r.URL)
		authorityScore := GetDomainAuthority(domain)

		// Check if result has publish date for temporal analysis
		publishDate := r.PublishedDate

		enhanced = append(enhanced, SearchContext{
			Title:          r.Title,
			URL:            r.URL,
			Domain:         domain,
			Snippet:        r.Content,
			PublishDate:    publishDate,
			Score:          r.Score,
			AuthorityScore: authorityScore,
		})
	}

	return enhanced
}

// EntityArbiter performs entity disambiguation using LLM
func EntityArbiter(ctx context.Context, llm llms.Model, query string, searchContexts []SearchContext) (*EntityArbiterResult, error) {
	fmt.Printf("EntityArbiter: 正在进行实体消歧，查询='%s'，搜索结果数=%d\n", query, len(searchContexts))

	// Prepare search context summary
	var searchSummary strings.Builder
	searchSummary.WriteString(fmt.Sprintf("## 用户查询\n%s\n\n", query))
	searchSummary.WriteString(fmt.Sprintf("## 搜索结果分析（共%d条）\n\n", len(searchContexts)))

	// Group results by domain authority
	highAuthority := make([]SearchContext, 0)
	mediumAuthority := make([]SearchContext, 0)
	lowAuthority := make([]SearchContext, 0)

	for _, sc := range searchContexts {
		if sc.AuthorityScore >= 85 {
			highAuthority = append(highAuthority, sc)
		} else if sc.AuthorityScore >= 70 {
			mediumAuthority = append(mediumAuthority, sc)
		} else {
			lowAuthority = append(lowAuthority, sc)
		}
	}

	// Sort results by authority score within each group
	// (simplified - you might want to implement proper sorting)

	// Output high authority results first
	searchSummary.WriteString("### 高权威来源 (85+分)\n")
	for i, sc := range highAuthority {
		searchSummary.WriteString(fmt.Sprintf("**[%d]** %s\n", i+1, sc.Title))
		searchSummary.WriteString(fmt.Sprintf("- 来源: %s (权威度: %.0f)\n", sc.Domain, sc.AuthorityScore))
		searchSummary.WriteString(fmt.Sprintf("- 链接: %s\n", sc.URL))
		if sc.PublishDate != "" {
			searchSummary.WriteString(fmt.Sprintf("- 发布时间: %s\n", sc.PublishDate))
		}
		searchSummary.WriteString(fmt.Sprintf("- 摘要: %s\n\n", sc.Snippet))
	}

	searchSummary.WriteString("\n### 中等权威来源 (70-84分)\n")
	for i, sc := range mediumAuthority {
		searchSummary.WriteString(fmt.Sprintf("**[%d]** %s\n", i+1, sc.Title))
		searchSummary.WriteString(fmt.Sprintf("- 来源: %s (权威度: %.0f)\n", sc.Domain, sc.AuthorityScore))
		searchSummary.WriteString(fmt.Sprintf("- 摘要: %s\n\n", sc.Snippet))
	}

	// Build LLM prompt
	systemPrompt := SystemPromptEntityArbiter + searchSummary.String()

	// Generate entity disambiguation result
	var result EntityArbiterResult
	err := generateJSON(ctx, llm, systemPrompt, query, &result)
	if err != nil {
		return nil, fmt.Errorf("实体消歧失败: %w", err)
	}

	result.Timestamp = time.Now()

	// Calculate confidence based on authority scores
	if result.PrimaryDomain != "" {
		authority := GetDomainAuthority(result.PrimaryDomain)
		// Boost confidence if high authority domain is used
		if authority >= 85 {
			result.Confidence = min(1.0, result.Confidence*1.2)
			result.IsVerified = true
		}
	}

	fmt.Printf("EntityArbiter: 消歧完成 -> 实体='%s', 类型='%s', 置信度=%.2f\n",
		result.EntityName, result.EntityType, result.Confidence)

	return &result, nil
}

// TemporalAnalysis performs time-sensitive analysis to identify latest information
func TemporalAnalysis(searchContexts []SearchContext) (*SearchContext, []SearchContext) {
	// Find the most recent result
	var mostRecent *SearchContext
	var recentResults []SearchContext
	var latestDate time.Time

	for i, sc := range searchContexts {
		if sc.PublishDate != "" {
			// Try to parse the date
			parsedDate, err := parsePublishDate(sc.PublishDate)
			if err == nil {
				if parsedDate.After(latestDate) {
					latestDate = parsedDate
					mostRecent = &searchContexts[i]
				}

				// Collect results from last 6 months
				sixMonthsAgo := time.Now().AddDate(0, -6, 0)
				if parsedDate.After(sixMonthsAgo) {
					recentResults = append(recentResults, sc)
				}
			}
		}
	}

	return mostRecent, recentResults
}

// parsePublishDate attempts to parse various date formats
func parsePublishDate(dateStr string) (time.Time, error) {
	// Common date formats to try
	formats := []string{
		time.RFC3339,
		"2006-01-02",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"January 2, 2006",
		"Jan 2, 2006",
		"2006年1月2日",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("无法解析日期: %s", dateStr)
}

// ContextualDisambiguation uses context clues to disambiguate entities
func ContextualDisambiguation(query string, searchContexts []SearchContext) map[string]float64 {
	scores := make(map[string]float64)

	queryLower := strings.ToLower(query)

	// Check for AI/ML context indicators
	aiKeywords := []string{"ai", "model", "gpt", "llm", "prompt", "generate", "image generation", "multimodal"}
	hardwareKeywords := []string{"cpu", "gpu", "board", "pin", "gpio", "hardware", "spec", "raspberry", "arduino"}
	softwareKeywords := []string{"api", "sdk", "library", "framework", "code", "programming", "tutorial"}

	aiScore := 0.0
	hardwareScore := 0.0
	softwareScore := 0.0

	for _, keyword := range aiKeywords {
		if strings.Contains(queryLower, keyword) {
			aiScore += 0.15
		}
	}

	for _, keyword := range hardwareKeywords {
		if strings.Contains(queryLower, keyword) {
			hardwareScore += 0.15
		}
	}

	for _, keyword := range softwareKeywords {
		if strings.Contains(queryLower, keyword) {
			softwareScore += 0.15
		}
	}

	// Also check search results for context
	for _, sc := range searchContexts {
		snippetLower := strings.ToLower(sc.Snippet)

		for _, keyword := range aiKeywords {
			if strings.Contains(snippetLower, keyword) {
				aiScore += 0.05 * sc.AuthorityScore / 100
			}
		}

		for _, keyword := range hardwareKeywords {
			if strings.Contains(snippetLower, keyword) {
				hardwareScore += 0.05 * sc.AuthorityScore / 100
			}
		}

		for _, keyword := range softwareKeywords {
			if strings.Contains(snippetLower, keyword) {
				softwareScore += 0.05 * sc.AuthorityScore / 100
			}
		}
	}

	scores["ai_model"] = min(1.0, aiScore)
	scores["hardware"] = min(1.0, hardwareScore)
	scores["software"] = min(1.0, softwareScore)

	return scores
}

// PerformEntityDisambiguation is the main entry point for entity disambiguation
func PerformEntityDisambiguation(ctx context.Context, llm llms.Model, query string) (*EntityArbiterResult, []SearchContext, error) {
	// Step 1: Perform initial search to gather context
	fmt.Println("EntityDisambiguation: 步骤1 - 执行初始搜索以收集上下文...")
	results, err := ExecuteSearch(ctx, query, "deep_search", "", "")
	if err != nil {
		return nil, nil, fmt.Errorf("初始搜索失败: %w", err)
	}

	// Step 2: Enhance search results with authority scores
	fmt.Println("EntityDisambiguation: 步骤2 - 增强搜索结果（添加权威度评分）...")
	searchContexts := EnhanceSearchResults(results)

	// Display authority analysis
	fmt.Println("\n=== 权威域分析 ===")
	for _, sc := range searchContexts {
		if sc.AuthorityScore >= 70 {
			fmt.Printf("  [%s] %.0f分 - %s\n", sc.Domain, sc.AuthorityScore, sc.Title)
		}
	}

	// Step 3: Perform temporal analysis
	fmt.Println("\nEntityDisambiguation: 步骤3 - 时间敏感性分析...")
	mostRecent, recentResults := TemporalAnalysis(searchContexts)
	if mostRecent != nil {
		fmt.Printf("  最新信息: %s (%s)\n", mostRecent.Title, mostRecent.PublishDate)
	}
	fmt.Printf("  近期信息数: %d篇（6个月内）\n", len(recentResults))

	// Step 4: Perform contextual disambiguation
	fmt.Println("\nEntityDisambiguation: 步骤4 - 上下文消歧分析...")
	contextScores := ContextualDisambiguation(query, searchContexts)
	fmt.Println("  上下文评分:")
	for entityType, score := range contextScores {
		if score > 0.1 {
			fmt.Printf("    %s: %.2f\n", entityType, score)
		}
	}

	// Step 5: Run entity arbiter
	fmt.Println("\nEntityDisambiguation: 步骤5 - 运行实体仲裁器...")
	result, err := EntityArbiter(ctx, llm, query, searchContexts)
	if err != nil {
		return nil, nil, err
	}

	// Apply contextual scores to boost confidence
	if contextScore, exists := contextScores[result.EntityType]; exists {
		result.Confidence = (result.Confidence + contextScore) / 2
	}

	return result, searchContexts, nil
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
