package query_engine

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/smallnest/langgraphgo/showcases/DeepInsight/schema"
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
		// 输出调试信息
		preview := content
		if len(preview) > 200 {
			preview = preview[:200] + "..."
		}
		return fmt.Errorf("LLM返回的内容不是有效的JSON，前200字符: %s", preview)
	}

	// 尝试解析JSON
	err = json.Unmarshal([]byte(content), output)
	if err != nil {
		// 输出更详细的错误信息
		preview := content
		if len(preview) > 200 {
			preview = preview[:200] + "..."
		}
		return fmt.Errorf("JSON解析失败: %w, 内容前200字符: %s", err, preview)
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

	// 1. Generate Report Structure
	fmt.Println("QueryEngine: 正在生成研究报告结构...")
	var structureWrapper struct {
		Paragraphs []struct {
			Title   string `json:"title"`
			Content string `json:"content"`
		} `json:"paragraphs"`
	}
	err = generateJSON(ctx, llm, SystemPromptReportStructure, s.Query, &structureWrapper)
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
			processParagraph(ctx, llm, s.Paragraphs[idx])
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

	// Store the result in ResearchResults as a single large string for now
	s.ResearchResults = []string{completion.Choices[0].Content}

	fmt.Println("QueryEngine: 深度调研完成。")
	return s, nil
}

func processParagraph(ctx context.Context, llm llms.Model, p *schema.Paragraph) {
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
		"title":   p.Title,
		"content": p.Content,
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
