package query_engine

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/smallnest/langgraphgo-showcases/DeepInsight/schema"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

// Helper to get LLM
func GetLLM(ctx context.Context) (llms.Model, error) {
	// Ensure OPENAI_API_KEY is set
	if os.Getenv("OPENAI_API_KEY") == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY not set")
	}
	opts := []openai.Option{}
	if base := os.Getenv("OPENAI_API_BASE"); base != "" {
		opts = append(opts, openai.WithBaseURL(base))
	}
	if model := os.Getenv("OPENAI_MODEL"); model != "" {
		opts = append(opts, openai.WithModel(model))
	}
	return openai.New(opts...)
}

// Helper to generate JSON from LLM
func generateJSON(ctx context.Context, llm llms.Model, systemPrompt, userContent string, output any) error {
	messages := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, systemPrompt),
		llms.TextParts(llms.ChatMessageTypeHuman, userContent),
	}

	completion, err := llm.GenerateContent(ctx, messages, llms.WithJSONMode())
	if err != nil {
		return err
	}

	content := completion.Choices[0].Content
	// Clean up markdown code blocks if present
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	// 验证JSON是否有效
	if !json.Valid([]byte(content)) {
		// 输出完整内容以便排查
		fmt.Printf("\n========== LLM返回的无效JSON内容 (长度: %d) ==========\n", len(content))
		fmt.Println(content)
		fmt.Println("========================================\n")
		return fmt.Errorf("LLM返回的内容不是有效的JSON，长度=%d，已输出完整内容", len(content))
	}

	// 尝试解析JSON
	err = json.Unmarshal([]byte(content), output)
	if err != nil {
		// 输出完整内容以便排查
		fmt.Printf("\n========== LLM返回的JSON内容 (长度: %d) ==========\n", len(content))
		fmt.Println(content)
		fmt.Println("========================================\n")
		return fmt.Errorf("JSON解析失败: %w，长度=%d，已输出完整内容", err, len(content))
	}

	return nil
}

// QueryEngineNode implements the main logic.
func QueryEngineNode(ctx context.Context, state any) (any, error) {
	s := state.(*schema.DeepInsightState)
	fmt.Printf("QueryEngine: 正在开始深度调研 '%s'...\n", s.Query)

	llm, err := GetLLM(ctx)
	if err != nil {
		return nil, fmt.Errorf("初始化 LLM 失败: %w", err)
	}

	// 0. 先进行背景文档搜索，为报告结构生成提供上下文
	fmt.Println("QueryEngine: 正在进行背景文档搜索...")
	backgroundDocs, err := fetchBackgroundDocuments(ctx, s.Query)
	if err != nil {
		fmt.Printf("QueryEngine: 背景文档搜索失败: %v，将不使用背景文档\n", err)
		backgroundDocs = "（未能获取背景文档信息）"
	} else {
		fmt.Printf("QueryEngine: 已获取 %d 篇背景文档\n", len(backgroundDocs))
	}

	// 1. Generate Report Structure
	fmt.Println("QueryEngine: 正在生成研究报告结构...")
	var structureWrapper struct {
		Paragraphs []struct {
			Title   string `json:"title"`
			Content string `json:"content"`
		} `json:"paragraphs"`
	}

	// 替换 {BACKGROUND_DOCUMENTS} 占位符
	systemPrompt := SystemPromptReportStructure
	systemPrompt = strings.ReplaceAll(systemPrompt, "{BACKGROUND_DOCUMENTS}", backgroundDocs)

	err = generateJSON(ctx, llm, systemPrompt, s.Query, &structureWrapper)
	if err != nil {
		return nil, fmt.Errorf("生成结构失败: %w", err)
	}

	s.Paragraphs = make([]*schema.Paragraph, len(structureWrapper.Paragraphs))
	for i, item := range structureWrapper.Paragraphs {
		s.Paragraphs[i] = &schema.Paragraph{
			Title:    item.Title,
			Content:  item.Content,
			Research: schema.NewResearchState(),
		}
		fmt.Printf("  - 规划段落: %s\n", item.Title)
	}

	// 2. Process Paragraphs (Parallel or Sequential)
	// For simplicity and to avoid rate limits, we'll do sequential for now, or limited concurrency.
	var wg sync.WaitGroup
	for i := range s.Paragraphs {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			processParagraph(ctx, llm, s.Query, s.Paragraphs[idx])
		}(i)
	}
	wg.Wait()

	// 3. Generate Final Report
	fmt.Println("QueryEngine: 正在生成最终总结报告...")

	// Prepare input for formatting
	var reportData []map[string]string
	for _, p := range s.Paragraphs {
		reportData = append(reportData, map[string]string{
			"title":                  p.Title,
			"paragraph_latest_state": p.Research.LatestSummary,
		})
	}
	reportDataJSON, _ := json.Marshal(reportData)

	messages := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, SystemPromptReportFormatting),
		llms.TextParts(llms.ChatMessageTypeHuman, string(reportDataJSON)),
	}

	completion, err := llm.GenerateContent(ctx, messages)
	if err != nil {
		return nil, err
	}

	// Clean up markdown code blocks
	resultContent := completion.Choices[0].Content
	resultContent = strings.TrimPrefix(resultContent, "```markdown")
	resultContent = strings.TrimPrefix(resultContent, "```md")
	resultContent = strings.TrimPrefix(resultContent, "```")
	resultContent = strings.TrimSuffix(resultContent, "```")
	resultContent = strings.TrimSpace(resultContent)

	// Store the result in ResearchResults as a single large string for now
	s.ResearchResults = []string{resultContent}

	fmt.Println("QueryEngine: 深度调研完成。")
	return s, nil
}

func processParagraph(ctx context.Context, llm llms.Model, originalQuery string, p *schema.Paragraph) {
	fmt.Printf("  正在处理段落: %s\n", p.Title)

	// --- Initial Search ---
	var firstSearchOutput struct {
		SearchQuery string `json:"search_query"`
		SearchTool  string `json:"search_tool"`
		Reasoning   string `json:"reasoning"`
		StartDate   string `json:"start_date"`
		EndDate     string `json:"end_date"`
	}

	inputJSON, _ := json.Marshal(map[string]string{
		"original_query": originalQuery,
		"title":          p.Title,
		"content":        p.Content,
	})

	err := generateJSON(ctx, llm, SystemPromptFirstSearch, string(inputJSON), &firstSearchOutput)
	if err != nil {
		fmt.Printf("    生成首次搜索失败 '%s': %v\n", p.Title, err)
		return
	}

	fmt.Printf("    搜索: %s (工具: %s)\n", firstSearchOutput.SearchQuery, firstSearchOutput.SearchTool)

	results, err := ExecuteSearch(ctx, firstSearchOutput.SearchQuery, firstSearchOutput.SearchTool, firstSearchOutput.StartDate, firstSearchOutput.EndDate)
	if err != nil {
		fmt.Printf("    搜索失败: %v\n", err)
		// Continue with empty results?
	}
	p.Research.AddSearchResults(firstSearchOutput.SearchQuery, results)

	// --- Initial Summary ---
	var firstSummaryOutput struct {
		ParagraphLatestState string `json:"paragraph_latest_state"`
	}

	resultsStr := formatResults(results)
	summaryInputJSON, _ := json.Marshal(map[string]any{
		"title":          p.Title,
		"content":        p.Content,
		"search_query":   firstSearchOutput.SearchQuery,
		"search_results": []string{resultsStr}, // Prompt expects array of strings
	})

	err = generateJSON(ctx, llm, SystemPromptFirstSummary, string(summaryInputJSON), &firstSummaryOutput)
	if err != nil {
		fmt.Printf("    ❌ 段落 '%s' 生成总结失败: %v\n", p.Title, err)
		fmt.Printf("    💡 建议：检查LLM返回的内容是否符合JSON格式要求\n")
		return
	}
	p.Research.LatestSummary = firstSummaryOutput.ParagraphLatestState

	// --- Reflection Loop (Max 1 for now to save time/tokens) ---
	maxReflections := 1
	for i := range maxReflections {
		fmt.Printf("    正在反思 (%d/%d)...\n", i+1, maxReflections)

		var reflectionOutput struct {
			SearchQuery string `json:"search_query"`
			SearchTool  string `json:"search_tool"`
			Reasoning   string `json:"reasoning"`
			StartDate   string `json:"start_date"`
			EndDate     string `json:"end_date"`
		}

		reflectInputJSON, _ := json.Marshal(map[string]string{
			"original_query":         originalQuery,
			"title":                  p.Title,
			"content":                p.Content,
			"paragraph_latest_state": p.Research.LatestSummary,
		})

		err = generateJSON(ctx, llm, SystemPromptReflection, string(reflectInputJSON), &reflectionOutput)
		if err != nil {
			fmt.Printf("    ❌ 段落 '%s' 反思查询生成失败 (轮次 %d/%d): %v\n", p.Title, i+1, maxReflections, err)
			break
		}

		fmt.Printf("    反思搜索: %s\n", reflectionOutput.SearchQuery)

		newResults, err := ExecuteSearch(ctx, reflectionOutput.SearchQuery, reflectionOutput.SearchTool, reflectionOutput.StartDate, reflectionOutput.EndDate)
		if err != nil {
			fmt.Printf("    反思搜索失败: %v\n", err)
			continue
		}
		p.Research.AddSearchResults(reflectionOutput.SearchQuery, newResults)

		// Update Summary
		var reflectionSummaryOutput struct {
			UpdatedParagraphLatestState string `json:"updated_paragraph_latest_state"`
		}

		newResultsStr := formatResults(newResults)
		reflectSummaryInputJSON, _ := json.Marshal(map[string]any{
			"title":                  p.Title,
			"content":                p.Content,
			"search_query":           reflectionOutput.SearchQuery,
			"search_results":         []string{newResultsStr},
			"paragraph_latest_state": p.Research.LatestSummary,
		})

		err = generateJSON(ctx, llm, SystemPromptReflectionSummary, string(reflectSummaryInputJSON), &reflectionSummaryOutput)
		if err != nil {
			fmt.Printf("    ❌ 段落 '%s' 反思总结生成失败 (轮次 %d/%d): %v\n", p.Title, i+1, maxReflections, err)
			break
		}
		p.Research.LatestSummary = reflectionSummaryOutput.UpdatedParagraphLatestState
	}

	p.Research.MarkCompleted()
	fmt.Printf("  段落 '%s' 完成。\n", p.Title)
}

func formatResults(results []schema.SearchResult) string {
	var sb strings.Builder
	for i, r := range results {
		sb.WriteString(fmt.Sprintf("[%d] Title: %s\nURL: %s\nDate: %s\nContent: %s\n\n", i+1, r.Title, r.URL, r.PublishedDate, r.Content))
	}
	return sb.String()
}

// fetchBackgroundDocuments performs an initial search using wechat_search to gather context
// Returns a formatted string of document summaries for background context
func fetchBackgroundDocuments(ctx context.Context, query string) (string, error) {
	// Use wechat_search to get relevant documents
	results, err := ExecuteSearch(ctx, query, "wechat_search", "", "")
	if err != nil {
		return "", fmt.Errorf("背景文档搜索失败: %w", err)
	}

	if len(results) == 0 {
		return "（未找到相关背景文档）", nil
	}

	// Limit to top 5 documents
	maxDocs := 5
	if len(results) < maxDocs {
		maxDocs = len(results)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## 查询: %s\n\n", query))
	sb.WriteString(fmt.Sprintf("找到 %d 篇相关文档（显示前 %d 篇）:\n\n", len(results), maxDocs))

	for i := 0; i < maxDocs; i++ {
		r := results[i]
		sb.WriteString(fmt.Sprintf("### 文档 %d: %s\n", i+1, r.Title))
		sb.WriteString(fmt.Sprintf("**来源**: %s\n", r.URL))
		if r.PublishedDate != "" {
			sb.WriteString(fmt.Sprintf("**发布时间**: %s\n", r.PublishedDate))
		}

		// Get first 1000 characters of content
		content := r.Content
		if len(content) > 1000 {
			content = content[:1000] + "..."
		}
		sb.WriteString(fmt.Sprintf("**内容摘要**: %s\n\n", content))
	}

	return sb.String(), nil
}
