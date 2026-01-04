package insight_engine

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/smallnest/langgraphgo/showcases/DeepInsight/query_engine"
	"github.com/smallnest/langgraphgo/showcases/DeepInsight/schema"
	"github.com/tmc/langchaingo/llms"
)

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// InsightEngineNode performs expert analysis and deep insight generation.
func InsightEngineNode(ctx context.Context, state any) (any, error) {
	s := state.(*schema.DeepInsightState)
	fmt.Printf("InsightEngine: 正在进行深度洞察分析 '%s'...\n", s.Query)

	llm, err := query_engine.GetLLM(ctx)
	if err != nil {
		return s, err
	}

	var insights []string

	// 1. Generate Insight Analysis Structure
	fmt.Println("InsightEngine: 正在规划洞察分析结构...")

	var structureWrapper struct {
		Paragraphs []struct {
			Title   string `json:"title"`
			Content string `json:"content"`
		} `json:"paragraphs"`
	}

	messages := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, SystemPromptReportStructure),
		llms.TextParts(llms.ChatMessageTypeHuman, s.Query),
	}
	completion, err := llm.GenerateContent(ctx, messages, llms.WithJSONMode())
	if err != nil {
		fmt.Printf("InsightEngine: 生成结构失败: %v\n", err)
		// Fallback
		insights = append(insights, "洞察分析结构生成失败。")
		s.InsightResults = insights
		return s, nil
	}

	content := completion.Choices[0].Content
	// Clean up markdown
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	// Check if JSON is valid
	if !json.Valid([]byte(content)) {
		fmt.Printf("InsightEngine: 结构JSON无效: %s\n", content[:min(200, len(content))])
		insights = append(insights, "洞察分析结构解析失败。")
		s.InsightResults = insights
		return s, nil
	}

	if err := json.Unmarshal([]byte(content), &structureWrapper); err != nil {
		fmt.Printf("InsightEngine: 解析结构失败: %v\n", err)
		insights = append(insights, "洞察分析结构解析失败。")
		s.InsightResults = insights
		return s, nil
	}

	// 2. Process each paragraph for deep insight
	for _, p := range structureWrapper.Paragraphs {
		fmt.Printf("InsightEngine: 分析洞察段落 '%s'...\n", p.Title)

		// Generate Search Query
		input := map[string]string{
			"title":   p.Title,
			"content": p.Content,
		}
		inputBytes, _ := json.Marshal(input)

		messages = []llms.MessageContent{
			llms.TextParts(llms.ChatMessageTypeSystem, SystemPromptFirstSearch),
			llms.TextParts(llms.ChatMessageTypeHuman, string(inputBytes)),
		}

		completion, err := llm.GenerateContent(ctx, messages, llms.WithJSONMode())
		var searchQuery, searchTool string
		if err != nil {
			fmt.Printf("InsightEngine: 生成搜索词失败: %v\n", err)
			searchQuery = p.Title // Fallback
			searchTool = "basic_search_news"
		} else {
			var searchOutput struct {
				SearchQuery string `json:"search_query"`
				SearchTool  string `json:"search_tool"`
				Reasoning   string `json:"reasoning"`
				StartDate   string `json:"start_date"`
				EndDate     string `json:"end_date"`
			}
			content := completion.Choices[0].Content
			content = strings.TrimPrefix(content, "```json")
			content = strings.TrimPrefix(content, "```")
			content = strings.TrimSuffix(content, "```")
			content = strings.TrimSpace(content)
			// Check if JSON is valid
			if !json.Valid([]byte(content)) {
				fmt.Printf("InsightEngine: 搜索查询JSON无效: %s\n", content[:min(200, len(content))])
			}
			if err := json.Unmarshal([]byte(content), &searchOutput); err != nil {
				fmt.Printf("InsightEngine: JSON解析失败: %v\n", err)
				searchQuery = s.Query // Fallback to original query
				searchTool = "basic_search_news"
			} else {
				searchQuery = searchOutput.SearchQuery
				searchTool = searchOutput.SearchTool
			}
			fmt.Printf("  搜索词: %s (工具: %s)\n", searchQuery, searchTool)
		}

		// Execute Search for expert insights
		results, err := query_engine.ExecuteSearch(ctx, searchQuery, searchTool, "", "")
		if err != nil {
			fmt.Printf("InsightEngine: 搜索失败: %v\n", err)
			continue
		}

		// Format results for analysis
		var resultStrs []string
		for _, r := range results {
			resultStrs = append(resultStrs, fmt.Sprintf("Title: %s\nURL: %s\nContent: %s", r.Title, r.URL, r.Content))
		}

		// Generate Insight Summary
		summaryInput := map[string]any{
			"title":          p.Title,
			"content":        p.Content,
			"search_query":   searchQuery,
			"search_results": resultStrs,
		}
		summaryInputBytes, _ := json.Marshal(summaryInput)

		messages = []llms.MessageContent{
			llms.TextParts(llms.ChatMessageTypeSystem, SystemPromptFirstSummary),
			llms.TextParts(llms.ChatMessageTypeHuman, string(summaryInputBytes)),
		}

		completion, err = llm.GenerateContent(ctx, messages, llms.WithJSONMode())
		if err != nil {
			fmt.Printf("InsightEngine: 生成洞察总结失败: %v\n", err)
			continue
		}

		var summaryOutput struct {
			ParagraphLatestState string `json:"paragraph_latest_state"`
		}
		content = completion.Choices[0].Content
		content = strings.TrimPrefix(content, "```json")
		content = strings.TrimPrefix(content, "```")
		content = strings.TrimSuffix(content, "```")
		content = strings.TrimSpace(content)
		// Check if JSON is valid
		if !json.Valid([]byte(content)) {
			fmt.Printf("InsightEngine: 总结JSON无效: %s\n", content[:min(200, len(content))])
			continue
		}
		if err := json.Unmarshal([]byte(content), &summaryOutput); err != nil {
			fmt.Printf("InsightEngine: JSON解析失败: %v\n", err)
			continue
		}

		insights = append(insights, fmt.Sprintf("### %s\n%s", p.Title, summaryOutput.ParagraphLatestState))
	}

	if len(insights) == 0 {
		// Fallback if structure generation failed
		insights = append(insights, "未能生成深度洞察分析。")
	}

	s.InsightResults = insights
	fmt.Println("InsightEngine: 深度洞察分析完成。")
	return s, nil
}
