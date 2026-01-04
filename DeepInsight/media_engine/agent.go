package media_engine

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/smallnest/langgraphgo-showcases/DeepInsight/query_engine"
	"github.com/smallnest/langgraphgo-showcases/DeepInsight/schema"
	"github.com/tmc/langchaingo/llms"
)

// MediaEngineNode searches for visual and contextual information.
func MediaEngineNode(ctx context.Context, state any) (any, error) {
	s := state.(*schema.DeepInsightState)
	fmt.Printf("MediaEngine: 正在搜索视觉和上下文信息 '%s'...\n", s.Query)

	llm, err := query_engine.GetLLM(ctx)
	if err != nil {
		return s, err
	}

	// 1. Generate search queries using SystemPromptFirstSearch
	input := map[string]string{
		"title":   s.Query,
		"content": s.Query,
	}
	inputBytes, _ := json.Marshal(input)
	inputStr := string(inputBytes)

	messages := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, SystemPromptFirstSearch),
		llms.TextParts(llms.ChatMessageTypeHuman, inputStr),
	}

	completion, err := llm.GenerateContent(ctx, messages)
	if err != nil {
		fmt.Printf("MediaEngine: 生成搜索词失败: %v\n", err)
		return s, nil
	}

	var output struct {
		SearchQuery string `json:"search_query"`
		SearchTool  string `json:"search_tool"`
		Reasoning   string `json:"reasoning"`
	}

	content := completion.Choices[0].Content
	// Clean up markdown code blocks if present
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")

	if err := json.Unmarshal([]byte(content), &output); err != nil {
		fmt.Printf("MediaEngine: 解析搜索词失败: %v\nContent: %s\n", err, content)
		// Fallback to original query
		output.SearchQuery = s.Query
	}

	fmt.Printf("MediaEngine: 生成的搜索词: %s (工具: %s)\n", output.SearchQuery, output.SearchTool)

	// 2. Execute search
	// Use Tavily to search for images and visual content
	results, err := query_engine.ExecuteSearch(ctx, output.SearchQuery, "search_images", "", "")
	if err != nil {
		fmt.Printf("MediaEngine: 搜索失败: %v\n", err)
		return s, nil
	}

	var mediaFindings []string
	for _, r := range results {
		// Record visual content information
		mediaFindings = append(mediaFindings, fmt.Sprintf("找到视觉内容: %s\n来源: %s\n上下文: %s", r.Title, r.URL, r.Content))
	}

	s.MediaResults = mediaFindings
	fmt.Printf("MediaEngine: 找到 %d 个视觉和上下文信息项。\n", len(mediaFindings))
	return s, nil
}
