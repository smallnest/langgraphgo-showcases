package media_engine

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/smallnest/langgraphgo/showcases/BettaFish/query_engine"
	"github.com/smallnest/langgraphgo/showcases/BettaFish/schema"
	"github.com/tmc/langchaingo/llms"
)

// MediaEngineNode searches for media content.
func MediaEngineNode(ctx context.Context, state any) (any, error) {
	s := state.(*schema.BettaFishState)
	fmt.Printf("MediaEngine: 正在搜索媒体内容 '%s'...\n", s.Query)

	llm, err := query_engine.GetLLM(ctx)
	if err != nil {
		return s, err
	}

	// 1. Generate search queries using SystemPromptFirstSearch
	// The prompt expects a paragraph title and content. We'll use the query for both.
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
	// Use Tavily to search for images
	results, err := query_engine.ExecuteSearch(ctx, output.SearchQuery, "search_images_for_news", "", "")
	if err != nil {
		fmt.Printf("MediaEngine: 搜索失败: %v\n", err)
		return s, nil
	}

	var mediaFindings []string
	for _, r := range results {
		// In a real system, we would download the image and pass it to a Vision LLM.
		// Here we just record the metadata.
		mediaFindings = append(mediaFindings, fmt.Sprintf("找到图片: %s\n来源: %s\n上下文: %s", r.Title, r.URL, r.Content))
	}

	s.MediaResults = mediaFindings
	fmt.Printf("MediaEngine: 找到 %d 个媒体项。\n", len(mediaFindings))
	return s, nil
}
