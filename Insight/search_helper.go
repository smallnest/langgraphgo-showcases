package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/smallnest/langgraphgo-showcases/Insight/tool"
)

// SearchHelper manages search operations for the researcher node.
type SearchHelper struct {
	registry *tool.Registry
}

// NewSearchHelper creates a new search helper.
func NewSearchHelper() *SearchHelper {
	return &SearchHelper{
		registry: tool.DefaultRegistry,
	}
}

// SelectTools determines which search tools to use based on the query.
// WeChat Search is always included as a mandatory tool for every search step.
func (sh *SearchHelper) SelectTools(query string) []string {
	queryLower := strings.ToLower(query)

	var selectedTools []string

	// WeChat Search is mandatory for every search step
	if tool, exists := sh.registry.Get("WeChat_Search"); exists {
		selectedTools = append(selectedTools, tool.Name())
	}

	// Detect if query needs Tavily search (general web search with AI)
	// Tavily is good for comprehensive web search and current information
	generalSearchKeywords := []string{"搜索", "search", "最新", "新闻", "news", "资讯", "信息", "查询", "了解", "什么是", "how", "what", "latest"}
	for _, kw := range generalSearchKeywords {
		if strings.Contains(queryLower, kw) {
			if tool, exists := sh.registry.Get("Tavily_Search"); exists {
				selectedTools = append(selectedTools, tool.Name())
			}
			break
		}
	}

	// Detect if query needs code search
	codeKeywords := []string{"代码", "code", "github", "开源", "repository", "api", "函数", "库", "编程", "programming", "项目", "git", "开发"}
	for _, kw := range codeKeywords {
		if strings.Contains(queryLower, kw) {
			if tool, exists := sh.registry.Get("GitHub_Search"); exists {
				selectedTools = append(selectedTools, tool.Name())
			}
			break
		}
	}

	// Detect if query needs academic search
	academicKeywords := []string{"论文", "paper", "研究", "research", "学术", "scholar", "理论", "算法", "algorithm", "实验", "文献", "期刊"}
	for _, kw := range academicKeywords {
		if strings.Contains(queryLower, kw) {
			if tool, exists := sh.registry.Get("Scholar_Search"); exists {
				selectedTools = append(selectedTools, tool.Name())
			}
			break
		}
	}

	// Detect if query needs social search (Zhihu)
	socialKeywords := []string{"讨论", "知乎", "观点", "评价", "经验", "zhihu", "社区", "问答", "如何看待"}
	for _, kw := range socialKeywords {
		if strings.Contains(queryLower, kw) {
			if tool, exists := sh.registry.Get("Zhihu_Search"); exists {
				selectedTools = append(selectedTools, tool.Name())
			}
			break
		}
	}

	// If no specific tools detected beyond WeChat, add default tools
	if len(selectedTools) == 1 {
		// Only WeChat_Search is in the list, add Tavily and Zhihu for comprehensive results
		defaultTools := []string{"Tavily_Search", "Zhihu_Search"}

		for _, toolName := range defaultTools {
			if tool, exists := sh.registry.Get(toolName); exists {
				selectedTools = append(selectedTools, tool.Name())
			}
		}
	}

	return selectedTools
}

// ExecuteSearch performs search using the selected tools.
// Limits total output to 80,000 characters to avoid token limits.
func (sh *SearchHelper) ExecuteSearch(ctx context.Context, query string, maxResultsPerTool int) string {
	toolNames := sh.SelectTools(query)

	if len(toolNames) == 0 {
		return ""
	}

	// Log which tools are being used
	fmt.Printf("[搜索] 使用工具: %v (查询: %s)\n", toolNames, query)

	var allResults []string
	totalLength := 0
	const maxLength = 80000 // Maximum total characters

	for _, toolName := range toolNames {
		searchTool, exists := sh.registry.Get(toolName)
		if !exists {
			continue
		}

		result, err := searchTool.Call(ctx, query)
		if err != nil {
			// Log error but continue with other tools
			fmt.Printf("[搜索] Warning: %s 失败: %v\n", toolName, err)
			continue
		}

		if result != "" {
			fmt.Printf("[搜索] %s 返回 %d 字符\n", toolName, len(result))

			// Check if adding this result would exceed the limit
			resultLength := len(result)
			headerLength := len(fmt.Sprintf("### 搜索来源: %s\n\n", toolName))
			totalWithHeader := totalLength + headerLength + resultLength + 5 // +5 for separator

			if totalWithHeader > maxLength {
				// Would exceed limit, truncate this result
				remaining := maxLength - totalLength - headerLength - 5
				if remaining > 0 {
					if remaining < resultLength {
						result = result[:remaining] + "\n\n... (内容过长，已截断)"
						fmt.Printf("[搜索] %s 结果被截断到 %d 字符以避免超限\n", toolName, remaining)
					}
					allResults = append(allResults, fmt.Sprintf("### 搜索来源: %s\n\n%s", toolName, result))
					totalLength = maxLength
					break // Stop processing more results
				}
				// Not enough space left, skip this result
				fmt.Printf("[搜索] %s 跳过（达到总长度限制）\n", toolName)
				break
			}

			allResults = append(allResults, fmt.Sprintf("### 搜索来源: %s\n\n%s", toolName, result))
			totalLength += headerLength + resultLength
		} else {
			fmt.Printf("[搜索] %s 无结果\n", toolName)
		}
	}

	if len(allResults) == 0 {
		return ""
	}

	joined := strings.Join(allResults, "\n\n---\n\n")
	fmt.Printf("[搜索] 总结果: %d 字符\n", len(joined))
	return joined
}

// GetAvailableTools returns a list of all available search tools.
func (sh *SearchHelper) GetAvailableTools() []string {
	return sh.registry.List()
}

// DescribeTools returns descriptions of all available tools.
func (sh *SearchHelper) DescribeTools() string {
	return sh.registry.DescribeAll()
}
