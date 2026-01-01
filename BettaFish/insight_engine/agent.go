package insight_engine

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/smallnest/langgraphgo/showcases/BettaFish/mind_spider"
	"github.com/smallnest/langgraphgo/showcases/BettaFish/query_engine"
	"github.com/smallnest/langgraphgo/showcases/BettaFish/schema"
	"github.com/smallnest/langgraphgo/showcases/BettaFish/sentiment_model"
	"github.com/tmc/langchaingo/llms"
)

// InsightEngineNode simulates a private database mining agent.
func InsightEngineNode(ctx context.Context, state any) (any, error) {
	s := state.(*schema.BettaFishState)
	fmt.Printf("InsightEngine: 正在挖掘内部洞察 '%s'...\n", s.Query)

	llm, err := query_engine.GetLLM(ctx)
	if err != nil {
		return s, err
	}

	var insights []string

	// 1. Mock Private DB Check (Keep existing logic)
	if _, err := os.Stat("internal_data.csv"); err == nil {
		insights = append(insights, "发现本地数据文件 'internal_data.csv'，但本演示未实现分析逻辑。")
	} else {
		insights = append(insights, "未配置内部数据库连接或未找到本地数据文件。")
	}

	// 2. Public Opinion Analysis using Prompts
	fmt.Println("InsightEngine: 正在规划舆情分析结构...")

	// 2.1 Generate Structure
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
	} else {
		content := completion.Choices[0].Content
		// Clean up markdown
		content = strings.TrimPrefix(content, "```json")
		content = strings.TrimPrefix(content, "```")
		content = strings.TrimSuffix(content, "```")

		if err := json.Unmarshal([]byte(content), &structureWrapper); err != nil {
			fmt.Printf("InsightEngine: 解析结构失败: %v\n", err)
		}
	}

	// 2.2 Process each paragraph
	for _, p := range structureWrapper.Paragraphs {
		fmt.Printf("InsightEngine: 分析段落 '%s'...\n", p.Title)

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

		completion, err = llm.GenerateContent(ctx, messages, llms.WithJSONMode())
		var searchQuery string
		if err != nil {
			fmt.Printf("InsightEngine: 生成搜索词失败: %v\n", err)
			searchQuery = p.Title // Fallback
		} else {
			var searchOutput struct {
				SearchQuery string `json:"search_query"`
				SearchTool  string `json:"search_tool"`
				Reasoning   string `json:"reasoning"`
			}
			content := completion.Choices[0].Content
			content = strings.TrimPrefix(content, "```json")
			content = strings.TrimPrefix(content, "```")
			content = strings.TrimSuffix(content, "```")
			if err := json.Unmarshal([]byte(content), &searchOutput); err != nil {
				fmt.Printf("InsightEngine: JSON解析失败: %v\n", err)
				searchQuery = s.Query // Fallback to original query
			} else {
				searchQuery = searchOutput.SearchQuery
			}
			fmt.Printf("  搜索词: %s (工具: %s)\n", searchQuery, searchOutput.SearchTool)
		}

		// Execute Search (using MindSpider)
		posts, err := mind_spider.CrawlSocialMedia(ctx, searchQuery)
		if err != nil {
			fmt.Printf("InsightEngine: 爬取失败: %v\n", err)
			continue
		}

		// Analyze Sentiment for top posts
		var analyzedPosts []string
		for i, post := range posts {
			if i >= 3 {
				break
			}
			sentiment, err := sentiment_model.AnalyzeSentiment(ctx, post)
			if err == nil {
				analyzedPosts = append(analyzedPosts, fmt.Sprintf("Post: %s\nSentiment: %s", post, sentiment))
			} else {
				analyzedPosts = append(analyzedPosts, fmt.Sprintf("Post: %s", post))
			}
		}

		// Generate Summary
		summaryInput := map[string]any{
			"title":          p.Title,
			"content":        p.Content,
			"search_query":   searchQuery,
			"search_results": analyzedPosts,
		}
		summaryInputBytes, _ := json.Marshal(summaryInput)

		messages = []llms.MessageContent{
			llms.TextParts(llms.ChatMessageTypeSystem, SystemPromptFirstSummary),
			llms.TextParts(llms.ChatMessageTypeHuman, string(summaryInputBytes)),
		}

		completion, err = llm.GenerateContent(ctx, messages, llms.WithJSONMode())
		if err == nil {
			var summaryOutput struct {
				ParagraphLatestState string `json:"paragraph_latest_state"`
			}
			content := completion.Choices[0].Content
			content = strings.TrimPrefix(content, "```json")
			content = strings.TrimPrefix(content, "```")
			content = strings.TrimSuffix(content, "```")
			if err := json.Unmarshal([]byte(content), &summaryOutput); err != nil {
				fmt.Printf("InsightEngine: JSON解析失败: %v\n", err)
			} else {
				insights = append(insights, fmt.Sprintf("### %s\n%s", p.Title, summaryOutput.ParagraphLatestState))
			}
		} else {
			fmt.Printf("InsightEngine: 生成总结失败: %v\n", err)
		}
	}

	if len(insights) == 0 {
		// Fallback if structure generation failed
		insights = append(insights, "未能生成深入分析。")
	}

	s.InsightResults = insights
	return s, nil
}
