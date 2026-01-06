package main

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

type logKey struct{}

func logf(ctx context.Context, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	// Always print to stdout
	fmt.Print(msg)

	// If log channel exists in context, send it there too
	if ch, ok := ctx.Value(logKey{}).(chan string); ok {
		// Non-blocking send to avoid stalling if channel is full or no one listening
		select {
		case ch <- msg:
		default:
		}
	}
}

// Typed node functions that wrap the untyped implementations
func QueryAgentNodeTyped(ctx context.Context, state *State) (*State, error) {
	result, err := QueryAgentNode(ctx, state)
	if err != nil {
		return nil, err
	}
	return result.(*State), nil
}

func PlannerNodeTyped(ctx context.Context, state *State) (*State, error) {
	result, err := PlannerNode(ctx, state)
	if err != nil {
		return nil, err
	}
	return result.(*State), nil
}

func ResearcherNodeTyped(ctx context.Context, state *State) (*State, error) {
	result, err := ResearcherNode(ctx, state)
	if err != nil {
		return nil, err
	}
	return result.(*State), nil
}

func InsightAgentNodeTyped(ctx context.Context, state *State) (*State, error) {
	result, err := InsightAgentNode(ctx, state)
	if err != nil {
		return nil, err
	}
	return result.(*State), nil
}

func ReporterNodeTyped(ctx context.Context, state *State) (*State, error) {
	result, err := ReporterNode(ctx, state)
	if err != nil {
		return nil, err
	}
	return result.(*State), nil
}

func PodcastNodeTyped(ctx context.Context, state *State) (*State, error) {
	result, err := PodcastNode(ctx, state)
	if err != nil {
		return nil, err
	}
	return result.(*State), nil
}

func StructurePlannerNodeTyped(ctx context.Context, state *State) (*State, error) {
	result, err := StructurePlannerNode(ctx, state)
	if err != nil {
		return nil, err
	}
	return result.(*State), nil
}

func ContentWriterNodeTyped(ctx context.Context, state *State) (*State, error) {
	result, err := ContentWriterNode(ctx, state)
	if err != nil {
		return nil, err
	}
	return result.(*State), nil
}

func ReflectorNodeTyped(ctx context.Context, state *State) (*State, error) {
	result, err := ReflectorNode(ctx, state)
	if err != nil {
		return nil, err
	}
	return result.(*State), nil
}

func ReviserNodeTyped(ctx context.Context, state *State) (*State, error) {
	result, err := ReviserNode(ctx, state)
	if err != nil {
		return nil, err
	}
	return result.(*State), nil
}

// QueryAgentNode analyzes user intent and performs entity disambiguation.
func QueryAgentNode(ctx context.Context, state any) (any, error) {
	s := state.(*State)
	logf(ctx, "--- 查询分析节点：正在分析用户意图和实体消歧 ---\n")
	logf(ctx, "原始查询: %s\n", s.Request.Query)

	llm, err := getLLM()
	if err != nil {
		return nil, err
	}

	// 意图识别和实体消歧的提示词
	prompt := fmt.Sprintf(`你是一位专业的查询分析专家。你的任务是深入分析用户的查询，识别真实意图，并进行实体消歧。

## 用户查询
%s

## 你的任务

### 1. 意图识别
分析用户的真实意图：
- **研究型意图**: 用户想深入了解某个主题的背景、原理、技术细节
- **对比型意图**: 用户想比较多个选项的优缺点
- **操作型意图**: 用户想学习如何使用某项技术或工具
- **新闻型意图**: 用户想了解最新的动态和趋势
- **概述型意图**: 用户想要一个全面的总结介绍

### 2. 实体消歧
识别查询中的关键实体，消除歧义：
- 确认实体的真实含义（例如："Apple"可能是公司或水果）
- 识别实体类型（AI模型、硬件、软件、概念等）
- 提供实体的权威描述

### 3. 查询优化
基于意图识别和实体消歧的结果，优化查询以便获得更好的搜索结果。

## 输出格式

请以 JSON 格式返回结果：
{
    "user_intent": "识别的用户意图（研究型/对比型/操作型/新闻型/概述型）",
    "refined_query": "优化后的查询描述",
    "entity_info": "实体的详细信息，包括名称、类型、关键特征等",
    "reasoning": "分析过程和判断理由"
}

必须使用中文回复。`, s.Request.Query)

	completion, err := llms.GenerateFromSinglePrompt(ctx, llm, prompt)
	if err != nil {
		return nil, err
	}

	// Clean up JSON
	completion = strings.TrimSpace(completion)
	completion = strings.TrimPrefix(completion, "```json")
	completion = strings.TrimPrefix(completion, "```")
	completion = strings.TrimSuffix(completion, "```")
	completion = strings.TrimSpace(completion)

	var output struct {
		UserIntent   string `json:"user_intent"`
		RefinedQuery string `json:"refined_query"`
		EntityInfo   any    `json:"entity_info"`
		Reasoning    string `json:"reasoning"`
	}

	if err := json.Unmarshal([]byte(completion), &output); err != nil {
		logf(ctx, "JSON 解析失败 (%v)，使用默认值\n", err)
		s.UserIntent = "研究型"
		s.RefinedQuery = s.Request.Query
		s.EntityInfo = "未能识别实体信息"
	} else {
		s.UserIntent = output.UserIntent
		s.RefinedQuery = output.RefinedQuery
		// Convert EntityInfo to string
		switch v := output.EntityInfo.(type) {
		case string:
			s.EntityInfo = v
		case map[string]any:
			// If it's an object, convert to JSON string
			if jsonBytes, err := json.Marshal(v); err == nil {
				s.EntityInfo = string(jsonBytes)
			} else {
				s.EntityInfo = fmt.Sprintf("%v", v)
			}
		default:
			s.EntityInfo = fmt.Sprintf("%v", v)
		}
	}

	logf(ctx, "识别意图: %s\n", s.UserIntent)
	logf(ctx, "优化查询: %s\n", s.RefinedQuery)
	logf(ctx, "实体信息: %s\n", s.EntityInfo)

	return s, nil
}

// PlannerNode generates a research plan based on the query.
func PlannerNode(ctx context.Context, state any) (any, error) {
	s := state.(*State)

	// Parse section titles from ReportStructure
	var sectionTitles []string
	if s.ReportStructure != "" {
		lines := strings.Split(s.ReportStructure, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "## ") && !strings.HasPrefix(line, "### ") {
				title := strings.TrimPrefix(line, "## ")
				sectionTitles = append(sectionTitles, title)
			}
		}
	}

	logf(ctx, "--- 规划节点：正在为 %d 个章节制定研究计划 ---\n", len(sectionTitles))

	llm, err := getLLM()
	if err != nil {
		return nil, err
	}

	// 使用优化后的查询进行规划
	queryToUse := s.RefinedQuery
	if queryToUse == "" {
		queryToUse = s.Request.Query
	}

	// Build section list for prompt
	var sectionsList strings.Builder
	sectionsList.WriteString("## 报告章节\n\n")
	for i, title := range sectionTitles {
		sectionsList.WriteString(fmt.Sprintf("%d. %s\n", i+1, title))
	}

	prompt := fmt.Sprintf(`你是一名专业的研究规划师。你的任务是为报告的每个章节制定针对性的研究计划。

## 原始查询
%s

## 用户意图
%s

## 实体信息
%s

%s

## 你的任务

请为报告的每个章节制定具体的研究任务。每个章节的研究计划应该：

1. **针对性强**: 研究任务应该紧扣该章节的主题和内容需求
2. **可操作性强**: 研究任务应该明确具体，便于执行搜索和信息收集
3. **覆盖全面**: 确保该章节所需的关键信息都被覆盖
4. **搜索友好**: 研究任务应该适合作为搜索关键词

## 输出格式

请以 JSON 格式返回每个章节的研究计划：

{
    "section_plans": [
        "章节1的研究任务描述（适合搜索的具体关键词或问题）",
        "章节2的研究任务描述",
        ...
    ],
    "generate_podcast": true/false
}

## 研究任务示例

- 如果章节是"研究背景"，研究任务可能是："搜索XX技术的发展历史、演进过程和当前状况"
- 如果章节是"技术原理"，研究任务可能是："搜索XX技术的核心原理、关键算法和实现机制"
- 如果章节是"应用案例"，研究任务可能是："搜索XX技术在实际项目中的应用案例和成功故事"

## 注意事项

- 研究任务应该简洁明确，每条 10-30 字
- 避免过于宽泛或过于具体
- 优先考虑可以使用搜索工具找到的信息
- 判断用户是否希望生成播客脚本（查询包含"播客"、"podcast"、"对话"、"脚本"等关键词时设为 true）

必须使用中文回复。`, queryToUse, s.UserIntent, s.EntityInfo, sectionsList.String())

	completion, err := llms.GenerateFromSinglePrompt(ctx, llm, prompt)
	if err != nil {
		return nil, err
	}

	// Clean up JSON
	completion = strings.TrimSpace(completion)
	completion = strings.TrimPrefix(completion, "```json")
	completion = strings.TrimPrefix(completion, "```")
	completion = strings.TrimSuffix(completion, "```")
	completion = strings.TrimSpace(completion)

	var output struct {
		SectionPlans    []string `json:"section_plans"`
		GeneratePodcast bool     `json:"generate_podcast"`
	}

	if err := json.Unmarshal([]byte(completion), &output); err != nil {
		logf(ctx, "JSON 解析失败 (%v)，使用默认研究计划\n", err)
		// Fallback: create generic research plans for each section
		output.SectionPlans = make([]string, len(sectionTitles))
		for i, title := range sectionTitles {
			output.SectionPlans[i] = fmt.Sprintf("搜索关于「%s」的相关信息", title)
		}
		// Check for podcast intent
		queryLower := strings.ToLower(queryToUse)
		output.GeneratePodcast = strings.Contains(queryLower, "播客") || strings.Contains(queryLower, "podcast")
	}

	s.SectionPlans = output.SectionPlans
	s.GeneratePodcast = output.GeneratePodcast

	// Also set Plan for compatibility
	s.Plan = output.SectionPlans

	// Format plans for better readability
	logf(ctx, "为各章节制定的研究计划：\n")
	for i, plan := range output.SectionPlans {
		if i < len(sectionTitles) {
			logf(ctx, "  [%d] %s\n  -> %s\n", i+1, sectionTitles[i], plan)
		} else {
			logf(ctx, "  [%d] %s\n", i+1, plan)
		}
	}
	if s.GeneratePodcast {
		logf(ctx, "检测到播客生成意图。\n")
	}

	return s, nil
}

// StructurePlannerNode determines the optimal report structure based on the query.
func StructurePlannerNode(ctx context.Context, state any) (any, error) {
	s := state.(*State)
	logf(ctx, "--- 结构规划节点：正在确定报告结构 ---\n")

	llm, err := getLLM()
	if err != nil {
		return nil, err
	}

	// 使用优化后的查询
	queryToUse := s.RefinedQuery
	if queryToUse == "" {
		queryToUse = s.Request.Query
	}

	prompt := fmt.Sprintf(`你是一位专业的报告结构规划师。基于用户查询，为报告确定最优的结构。

## 用户查询
%s

## 用户意图
%s

## 实体信息
%s

## 你的任务

请分析查询内容，设计一个最适合该主题的报告结构。结构应该：

1. **逻辑清晰**: 章节之间有明确的逻辑关系
2. **内容全面**: 覆盖主题的所有重要方面
3. **层次分明**: 主要章节和子章节划分合理
4. **详实具体**: 每个章节都有明确的内容指向

## 输出格式

请以 JSON 格式返回报告结构：

{
    "structure": "报告结构的 Markdown 格式大纲（使用 ## 表示主要章节，### 表示子章节）",
    "sections": [
        "章节1的完整标题",
        "章节2的完整标题",
        ...
    ]
}

要求：
- 报告应该包含 3-6 个主要章节
- 每个章节都应该是独立的、有意义的内容单元
- 章节标题应该清晰明了，反映该章节的核心内容
- 必须包含"执行摘要"或"概述"作为第一个章节
- 必须包含"结论与建议"或类似章节作为最后一个章节
- 根据主题特点，决定是否需要"背景介绍"、"技术细节"、"案例分析"、"对比分析"等章节

必须使用中文回复。`, queryToUse, s.UserIntent, s.EntityInfo)

	completion, err := llms.GenerateFromSinglePrompt(ctx, llm, prompt)
	if err != nil {
		return nil, err
	}

	// Clean up JSON
	completion = strings.TrimSpace(completion)
	completion = strings.TrimPrefix(completion, "```json")
	completion = strings.TrimPrefix(completion, "```")
	completion = strings.TrimSuffix(completion, "```")
	completion = strings.TrimSpace(completion)

	var output struct {
		Structure string   `json:"structure"`
		Sections  []string `json:"sections"`
	}

	if err := json.Unmarshal([]byte(completion), &output); err != nil {
		logf(ctx, "JSON 解析失败 (%v)，使用默认结构\n", err)
		// Fallback to default structure
		output.Structure = `## 执行摘要

## 研究背景

## 详细分析

## 结论与建议`
		output.Sections = []string{
			"执行摘要",
			"研究背景",
			"详细分析",
			"结论与建议",
		}
	}

	s.ReportStructure = output.Structure
	s.ReportSections = make([]string, len(output.Sections))
	s.CurrentSection = 0

	logf(ctx, "确定的报告结构（%d 个章节）：\n%s\n", len(output.Sections), output.Structure)
	for i, section := range output.Sections {
		logf(ctx, "  %d. %s\n", i+1, section)
	}

	return s, nil
}

// ResearcherNode executes the research plan using LLM and search tools.
func ResearcherNode(ctx context.Context, state any) (any, error) {
	s := state.(*State)

	// Use SectionPlans if available, otherwise fall back to Plan
	plans := s.SectionPlans
	if len(plans) == 0 {
		plans = s.Plan
	}

	// Parse section titles
	var sectionTitles []string
	if s.ReportStructure != "" {
		lines := strings.Split(s.ReportStructure, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "## ") && !strings.HasPrefix(line, "### ") {
				title := strings.TrimPrefix(line, "## ")
				sectionTitles = append(sectionTitles, title)
			}
		}
	}

	logf(ctx, "--- 研究节点：正在执行研究计划（%d 个任务） ---\n", len(plans))

	llm, err := getLLM()
	if err != nil {
		return nil, err
	}

	// Create search helper
	searchHelper := NewSearchHelper()

	var results []string
	for i, step := range plans {
		// Get section title if available
		sectionTitle := ""
		if i < len(sectionTitles) {
			sectionTitle = sectionTitles[i]
		} else {
			sectionTitle = fmt.Sprintf("章节 %d", i+1)
		}

		logf(ctx, "正在研究: [%s] %s\n", sectionTitle, step)

		// First, try to search using available tools
		searchResults := searchHelper.ExecuteSearch(ctx, step, 5)

		// Determine if we need to summarize the search results first
		// If search results are very long (>10000 chars), use a two-step approach:
		// Step 1: Summarize the search results
		// Step 2: Generate the chapter research materials based on the summary
		var researchMaterials string
		if len(searchResults) > 10000 && len(searchResults) > 0 {
			logf(ctx, "搜索结果过长 (%d 字符)，先进行摘要...\n", len(searchResults))

			// Step 1: Summarize the search results
			summarizePrompt := fmt.Sprintf(`你是一名专业的研究助理。请将以下搜索结果进行精炼的摘要。

## 研究任务
%s

## 搜索结果（原文）
%s

## 你的任务
请将上述搜索结果摘要为约 2000 字的内容，要求：
1. **保留核心信息**: 保留所有重要的数据、事实、观点和结论
2. **逻辑清晰**: 按照主题分类组织内容
3. **详实具体**: 不要省略重要的细节和数据
4. **结构分明**: 使用适当的标题和分段

请直接输出摘要内容，不要包含"摘要如下"等引导语。
必须使用中文回复。`, step, searchResults)

			summary, err := llms.GenerateFromSinglePrompt(ctx, llm, summarizePrompt)
			if err != nil {
				logf(ctx, "摘要失败，直接使用原始搜索结果: %v\n", err)
				researchMaterials = searchResults
			} else {
				logf(ctx, "摘要完成 (%d 字符 -> %d 字符)\n", len(searchResults), len(summary))
				researchMaterials = summary
			}
		} else {
			researchMaterials = searchResults
		}

		// Step 2: Generate chapter research materials
		var prompt string
		if len(researchMaterials) > 0 {
			prompt = fmt.Sprintf(`你是一名专业的研究员。请基于以下研究材料，为 "%s" 章节提供详细的研究内容。

## 章节
%s

## 研究任务
%s

## 研究材料
%s

## 你的任务
请综合以上研究材料，为该章节提供：
1. **关键信息摘要**: 用3-5句话总结核心发现
2. **重要见解**: 列出3-5个关键要点
3. **具体细节**: 补充重要的背景信息和细节
4. **数据支持**: 提供具体的数据、案例或实例

请以结构化的方式输出，每一点都要详细展开（每点至少50字）。
必须使用中文回复。`, sectionTitle, sectionTitle, step, researchMaterials)
		} else {
			// Fallback to pure LLM generation if no search results
			prompt = fmt.Sprintf(`你是一名专业的研究员。请为 "%s" 章节提供详细的研究材料。

## 章节
%s

## 研究任务
%s

请提供：
1. **关键信息摘要**: 用3-5句话总结核心发现
2. **重要见解**: 列出3-5个关键要点
3. **具体细节**: 补充重要的背景信息和细节
4. **数据支持**: 提供具体的数据、案例或实例

请以结构化的方式输出，每一点都要详细展开（每点至少50字）。
必须使用中文回复。`, sectionTitle, sectionTitle, step)
		}

		completion, err := llms.GenerateFromSinglePrompt(ctx, llm, prompt)
		if err != nil {
			return nil, err
		}
		results = append(results, fmt.Sprintf("## 章节：%s\n\n研究任务：%s\n\n%s", sectionTitle, step, completion))
	}

	s.ResearchResults = results
	s.Images = nil // No images collected
	return s, nil
}

// InsightAgentNode performs deep insight analysis based on research results.
func InsightAgentNode(ctx context.Context, state any) (any, error) {
	s := state.(*State)
	logf(ctx, "--- 洞察分析节点：正在进行深度洞察分析 ---\n")

	llm, err := getLLM()
	if err != nil {
		return nil, err
	}

	// 准备研究数据摘要
	researchData := strings.Join(s.ResearchResults, "\n\n")

	// 根据用户意图生成不同类型的洞察分析
	var insightPrompt string
	switch s.UserIntent {
	case "研究型":
		insightPrompt = fmt.Sprintf(`你是一位深度洞察分析师。基于以下研究结果，进行深层次的分析和洞察。

## 原始查询
%s

## 用户意图
研究型 - 用户希望深入了解该主题的背景、原理和技术细节

## 实体信息
%s

## 研究结果
%s

## 你的任务
请进行以下深度洞察分析：

1. **核心要点提取**: 总结研究中的关键发现和核心观点
2. **深层原理分析**: 分析背后的原理、机制和因果关系
3. **趋势与模式识别**: 识别发展趋势和模式
4. **关键洞察**: 提炼出有价值的深度洞察
5. **实践意义**: 分析研究结果的实际应用价值

请以结构化的 Markdown 格式输出分析结果，每个部分都要详细展开。
必须使用中文回复。`, s.Request.Query, s.EntityInfo, researchData)
	case "对比型":
		insightPrompt = fmt.Sprintf(`你是一位对比分析专家。基于以下研究结果，进行全面的对比分析。

## 原始查询
%s

## 用户意图
对比型 - 用户希望比较不同选项的优缺点

## 研究结果
%s

## 你的任务
请进行以下对比分析：

1. **选项识别**: 识别需要对比的主要选项
2. **多维度对比**: 从多个维度进行详细对比（性能、成本、易用性等）
3. **优缺点分析**: 分析每个选项的优势和劣势
4. **适用场景**: 分析各选项的最佳使用场景
5. **推荐建议**: 基于分析给出选择建议

请以结构化的 Markdown 格式输出分析结果。
必须使用中文回复。`, s.Request.Query, researchData)
	case "操作型":
		insightPrompt = fmt.Sprintf(`你是一位实践指南专家。基于以下研究结果，提炼实用的操作指南。

## 原始查询
%s

## 用户意图
操作型 - 用户希望学习如何使用某项技术或工具

## 研究结果
%s

## 你的任务
请提炼以下操作指南：

1. **快速入门**: 提供快速上手的步骤
2. **核心概念**: 解释关键概念和术语
3. **最佳实践**: 总结行业最佳实践
4. **常见问题**: 列出常见问题和解决方案
5. **进阶技巧**: 提供进阶使用技巧

请以结构化的 Markdown 格式输出指南。
必须使用中文回复。`, s.Request.Query, researchData)
	case "新闻型":
		insightPrompt = fmt.Sprintf(`你是一位趋势分析专家。基于以下研究结果，分析最新动态和趋势。

## 原始查询
%s

## 用户意图
新闻型 - 用户希望了解最新的动态和趋势

## 研究结果
%s

## 你的任务
请进行以下趋势分析：

1. **最新动态**: 总结最新的重要事件和发展
2. **影响分析**: 分析这些动态的影响和意义
3. **趋势预测**: 基于当前信息预测未来趋势
4. **关键观点**: 提炼专家和业界的观点
5. **关注重点**: 指出需要重点关注的方向

请以结构化的 Markdown 格式输出分析结果。
必须使用中文回复。`, s.Request.Query, researchData)
	default: // 概述型或其他
		insightPrompt = fmt.Sprintf(`你是一位综合分析专家。基于以下研究结果，进行全面的分析和总结。

## 原始查询
%s

## 用户意图
%s

## 实体信息
%s

## 研究结果
%s

## 你的任务
请进行全面的分析和总结：

1. **核心要点**: 总结研究中的关键信息
2. **深度分析**: 进行深入的分析和解读
3. **洞察发现**: 提炼有价值的洞察
4. **实用建议**: 提供实践建议
5. **总结评价**: 给出综合评价

请以结构化的 Markdown 格式输出分析结果，每个部分都要详细展开。
必须使用中文回复。`, s.Request.Query, s.UserIntent, s.EntityInfo, researchData)
	}

	completion, err := llms.GenerateFromSinglePrompt(ctx, llm, insightPrompt)
	if err != nil {
		return nil, err
	}

	// 清理输出
	completion = strings.TrimSpace(completion)
	completion = strings.TrimPrefix(completion, "```markdown")
	completion = strings.TrimPrefix(completion, "```")
	completion = strings.TrimSuffix(completion, "```")

	// Note: This node is no longer used in the main workflow
	// The field has been removed from State struct
	logf(ctx, "深度洞察分析完成。\n")
	return s, nil
}

// ContentWriterNode writes content for a specific section of the report.
func ContentWriterNode(ctx context.Context, state any) (any, error) {
	s := state.(*State)

	// Get the current section index
	idx := s.CurrentSection

	// Get section titles
	var sectionTitles []string
	if s.ReportStructure != "" {
		// Parse section titles from structure
		lines := strings.Split(s.ReportStructure, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "## ") && !strings.HasPrefix(line, "### ") {
				title := strings.TrimPrefix(line, "## ")
				sectionTitles = append(sectionTitles, title)
			}
		}
	}

	// Get current section title
	currentSectionTitle := ""
	if idx < len(sectionTitles) {
		currentSectionTitle = sectionTitles[idx]
	} else if idx < len(s.ReportSections) {
		// Try to get title from ReportSections (should have been populated)
		// Since ReportSections is empty initially, we need a fallback
		currentSectionTitle = fmt.Sprintf("章节 %d", idx+1)
	}

	logf(ctx, "--- 内容写作节点：正在撰写「%s」(%d/%d) ---\n", currentSectionTitle, idx+1, len(s.ReportSections))

	llm, err := getLLM()
	if err != nil {
		return nil, err
	}

	// Prepare research data for the current section
	// Only use the research results that correspond to this section
	var researchData string
	if idx < len(s.ResearchResults) {
		researchData = s.ResearchResults[idx]
	} else {
		// Fallback: join all research results if index out of bounds
		researchData = strings.Join(s.ResearchResults, "\n\n")
	}

	// If research data is still too long, summarize it first
	var finalResearchMaterial string
	if len(researchData) > 15000 && len(researchData) > 0 {
		logf(ctx, "研究材料过长 (%d 字符)，先进行摘要...\n", len(researchData))

		summarizePrompt := fmt.Sprintf(`你是一名专业的研究助理。请将以下研究材料进行精炼的摘要。

## 目标章节
%s

## 研究材料（原文）
%s

## 你的任务
请将上述研究材料摘要为约 2000 字的内容，要求：
1. **保留核心信息**: 保留所有重要的数据、事实、观点和结论
2. **逻辑清晰**: 按照主题分类组织内容
3. **详实具体**: 不要省略重要的细节和数据
4. **结构分明**: 使用适当的标题和分段

请直接输出摘要内容，不要包含"摘要如下"等引导语。
必须使用中文回复。`, currentSectionTitle, researchData)

		summary, err := llms.GenerateFromSinglePrompt(ctx, llm, summarizePrompt)
		if err != nil {
			logf(ctx, "摘要失败，直接使用原始研究材料: %v\n", err)
			finalResearchMaterial = researchData
		} else {
			logf(ctx, "摘要完成 (%d 字符 -> %d 字符)\n", len(researchData), len(summary))
			finalResearchMaterial = summary
		}
	} else {
		finalResearchMaterial = researchData
	}

	// Prepare context from previously written sections
	var previousSections strings.Builder
	if idx > 0 {
		previousSections.WriteString("## 已完成的章节内容（供参考）\n\n")
		for i := 0; i < idx; i++ {
			if i < len(s.ReportSections) && s.ReportSections[i] != "" {
				title := ""
				if i < len(sectionTitles) {
					title = sectionTitles[i]
				} else {
					title = fmt.Sprintf("章节 %d", i+1)
				}
				previousSections.WriteString(fmt.Sprintf("### %s\n\n%s\n\n", title, s.ReportSections[i]))
			}
		}
	}

	// Get section titles for context
	var allSections strings.Builder
	allSections.WriteString("## 报告章节规划\n\n")
	for i, title := range sectionTitles {
		prefix := "  "
		if i == idx {
			prefix = "-> "
		}
		allSections.WriteString(fmt.Sprintf("%s%d. %s\n", prefix, i+1, title))
	}

	// Build prompt for writing the current section
	queryToUse := s.RefinedQuery
	if queryToUse == "" {
		queryToUse = s.Request.Query
	}

	prompt := fmt.Sprintf(`你是一位专业的技术报告撰写专家。请为报告撰写指定的章节内容。

## 原始查询
%s

## 用户意图
%s

## 实体信息
%s

%s

## 当前章节
**序号**: %d / %d
**标题**: %s

## 你的任务

请为该章节撰写详细、专业的内容。

### 内容要求
1. **内容详实**: 至少 500 字，内容充实具体
2. **逻辑清晰**: 条理分明，层次清楚
3. **专业准确**: 使用专业术语，表述准确
4. **论据充分**: 基于研究结果，有理有据
5. **可读性强**: 语言流畅，易于理解

### 格式要求
- 使用 Markdown 格式
- 适当使用子标题（###）组织内容
- 使用列表（- 或 1.）列出要点
- 关键概念使用**加粗**突出
- 需要时使用代码块展示技术内容

### 注意事项
- 这是整个报告的一部分，不需要完整的开头和结尾
- 内容应该紧扣章节标题，不要偏离主题
- 充分利用研究结果中的信息和数据
- 如果是技术章节，应该包含具体的技术细节和实现方法
- 如果是分析章节，应该包含深入的分析和见解
- 保持客观、专业的语气

必须使用中文撰写。`,
		queryToUse,
		s.UserIntent,
		s.EntityInfo,
		allSections.String(),
		idx+1,
		len(s.ReportSections),
		currentSectionTitle)

	// Add previous sections as context if available
	if previousSections.Len() > 0 {
		prompt += fmt.Sprintf("\n\n%s\n--- 研究数据 ---\n\n%s\n\n", previousSections.String(), finalResearchMaterial)
	} else {
		prompt += fmt.Sprintf("\n\n--- 研究数据 ---\n\n%s\n\n", finalResearchMaterial)
	}

	prompt += "请直接输出该章节的详细内容，不要包含章节标题（标题会自动添加）。"

	completion, err := llms.GenerateFromSinglePrompt(ctx, llm, prompt)
	if err != nil {
		return nil, err
	}

	// Clean up the output
	completion = strings.TrimSpace(completion)
	completion = strings.TrimPrefix(completion, "```markdown")
	completion = strings.TrimPrefix(completion, "```")
	completion = strings.TrimSuffix(completion, "```")

	// Store the written content
	s.ReportSections[idx] = completion

	logf(ctx, "章节「%s」撰写完成（%d 字符）\n", currentSectionTitle, len(completion))

	return s, nil
}

// ReflectorNode reflects on the current section content to check if it meets requirements
func ReflectorNode(ctx context.Context, state any) (any, error) {
	s := state.(*State)
	idx := s.CurrentSection

	// Get section title
	var sectionTitles []string
	if s.ReportStructure != "" {
		lines := strings.Split(s.ReportStructure, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "## ") && !strings.HasPrefix(line, "### ") {
				title := strings.TrimPrefix(line, "## ")
				sectionTitles = append(sectionTitles, title)
			}
		}
	}

	currentSectionTitle := ""
	if idx < len(sectionTitles) {
		currentSectionTitle = sectionTitles[idx]
	} else {
		currentSectionTitle = fmt.Sprintf("章节 %d", idx+1)
	}

	// Initialize reflection results array if needed
	if s.ReflectionResults == nil {
		s.ReflectionResults = make([]string, len(s.ReportSections))
	}

	// Get current section content
	currentContent := s.ReportSections[idx]
	if currentContent == "" {
		logf(ctx, "--- 反思节点：章节「%s」内容为空，跳过反思 ---\n", currentSectionTitle)
		s.ReflectionResults[idx] = "无需补充：内容为空"
		return s, nil
	}

	logf(ctx, "--- 反思节点：正在检查章节「%s」内容质量 ---\n", currentSectionTitle)

	llm, err := getLLM()
	if err != nil {
		return nil, err
	}

	// Get section plan for context
	sectionPlan := ""
	if idx < len(s.SectionPlans) {
		sectionPlan = s.SectionPlans[idx]
	}

	// Build reflection prompt
	prompt := fmt.Sprintf(`你是一名专业的质量审查专家。请对以下章节内容进行审查和评估。

## 章节
标题：%s

## 研究任务
%s

## 章节内容
%s

## 审查标准

请从以下维度评估该章节内容：

1. **内容完整性** (150字以上)
   - 是否覆盖了章节主题的所有重要方面？
   - 是否遗漏了关键信息或要点？

2. **内容详实度** (150字以上)
   - 内容是否充实具体？
   - 是否有足够的细节、数据和案例支撑？

3. **逻辑清晰度** (100字以上)
   - 内容结构是否清晰？
   - 论述逻辑是否连贯？

4. **专业准确性** (100字以上)
   - 专业术语使用是否准确？
   - 技术描述是否准确无误？

5. **可读性** (100字以上)
   - 语言表达是否流畅？
   - 是否易于理解？

## 输出格式

请以 JSON 格式返回评估结果：

{
    "overall_assessment": "整体评价（优/良/中/差）",
    "needs_revision": true/false,
    "missing_points": [
        "缺失点1的详细描述",
        "缺失点2的详细描述",
        ...
    ],
    "revision_suggestions": [
        "改进建议1",
        "改进建议2",
        ...
    ],
    "strengths": [
        "优点1",
        "优点2"
    ],
    "detailed_feedback": "详细的反馈意见（300字以上）"
}

## 判断标准

- 如果内容质量达到"良"或"优"的水平，且没有明显缺失，则 needs_revision 设为 false
- 如果存在明显的内容缺失、质量问题或不足，则 needs_revision 设为 true

必须使用中文回复。`, currentSectionTitle, sectionPlan, currentContent)

	completion, err := llms.GenerateFromSinglePrompt(ctx, llm, prompt)
	if err != nil {
		logf(ctx, "反思失败: %v，标记为无需补充\n", err)
		s.ReflectionResults[idx] = "无需补充：反思过程出错"
		return s, nil
	}

	// Clean up JSON
	completion = strings.TrimSpace(completion)
	completion = strings.TrimPrefix(completion, "```json")
	completion = strings.TrimPrefix(completion, "```")
	completion = strings.TrimSuffix(completion, "```")
	completion = strings.TrimSpace(completion)

	// Parse reflection result
	var reflection struct {
		OverallAssessment   string   `json:"overall_assessment"`
		NeedsRevision       bool     `json:"needs_revision"`
		MissingPoints       []string `json:"missing_points"`
		RevisionSuggestions []string `json:"revision_suggestions"`
		Strengths           []string `json:"strengths"`
		DetailedFeedback    string   `json:"detailed_feedback"`
	}

	if err := json.Unmarshal([]byte(completion), &reflection); err != nil {
		logf(ctx, "反思结果解析失败: %v，使用默认判断\n", err)
		// Default: if content is too short, mark for revision
		if len(currentContent) < 500 {
			reflection.NeedsRevision = true
			reflection.MissingPoints = []string{"内容过短，需要扩展到至少500字"}
			reflection.DetailedFeedback = "内容长度不足，需要补充更多细节"
		} else {
			reflection.NeedsRevision = false
			reflection.DetailedFeedback = "内容长度达标"
		}
	}

	// Store reflection result
	reflectionSummary := fmt.Sprintf("整体评价: %s | 需要补充: %v", reflection.OverallAssessment, reflection.NeedsRevision)
	if reflection.DetailedFeedback != "" {
		reflectionSummary += fmt.Sprintf("\n反馈: %s", reflection.DetailedFeedback)
	}
	s.ReflectionResults[idx] = reflectionSummary

	if reflection.NeedsRevision {
		logf(ctx, "反思结果：需要补充（%s）\n", reflection.OverallAssessment)
		if len(reflection.MissingPoints) > 0 {
			for _, point := range reflection.MissingPoints {
				logf(ctx, "  - %s\n", point)
			}
		}
		if len(reflection.RevisionSuggestions) > 0 {
			logf(ctx, "改进建议：\n")
			for _, suggestion := range reflection.RevisionSuggestions {
				logf(ctx, "  - %s\n", suggestion)
			}
		}
	} else {
		logf(ctx, "反思结果：内容合格（%s）\n", reflection.OverallAssessment)
		if len(reflection.Strengths) > 0 {
			logf(ctx, "优点：\n")
			for _, strength := range reflection.Strengths {
				logf(ctx, "  - %s\n", strength)
			}
		}
	}

	return s, nil
}

// ReviserNode revises the current section content based on reflection feedback
func ReviserNode(ctx context.Context, state any) (any, error) {
	s := state.(*State)
	idx := s.CurrentSection

	// Get section title
	var sectionTitles []string
	if s.ReportStructure != "" {
		lines := strings.Split(s.ReportStructure, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "## ") && !strings.HasPrefix(line, "### ") {
				title := strings.TrimPrefix(line, "## ")
				sectionTitles = append(sectionTitles, title)
			}
		}
	}

	currentSectionTitle := ""
	if idx < len(sectionTitles) {
		currentSectionTitle = sectionTitles[idx]
	} else {
		currentSectionTitle = fmt.Sprintf("章节 %d", idx+1)
	}

	// Get revision count
	revisionCount := 0
	if s.RevisionCounts != nil && idx < len(s.RevisionCounts) {
		revisionCount = s.RevisionCounts[idx]
	}

	logf(ctx, "--- 补充节点：正在补充章节「%s」内容（第 %d 轮）---\n", currentSectionTitle, revisionCount+1)

	llm, err := getLLM()
	if err != nil {
		return nil, err
	}

	// Get current content
	currentContent := s.ReportSections[idx]

	// Get reflection feedback
	reflection := s.ReflectionResults[idx]

	// Get section plan
	sectionPlan := ""
	if idx < len(s.SectionPlans) {
		sectionPlan = s.SectionPlans[idx]
	}

	// Get research data for this section
	var researchData string
	if idx < len(s.ResearchResults) {
		researchData = s.ResearchResults[idx]
	}

	// Build revision prompt
	prompt := fmt.Sprintf(`你是一名专业的内容编辑。请根据反馈意见对章节内容进行补充和完善。

## 章节
标题：%s

## 研究任务
%s

## 当前内容
%s

## 反思反馈
%s

## 研究数据
%s

## 你的任务

请根据反思反馈中的建议，对当前内容进行补充和完善：

1. **补充缺失内容**: 针对 feedback 中指出的缺失点，补充相关内容
2. **扩展细节**: 在现有内容基础上，增加更多细节、数据和案例
3. **改进表达**: 优化语言表达，提高可读性
4. **保持连贯**: 确保补充内容与原内容自然衔接

## 输出要求

- 直接输出完整的章节内容（包含原内容和补充内容）
- 不要包含章节标题（标题会自动添加）
- 保持原有的结构和格式
- 补充内容应该自然融入，不要显得突兀

## 补充轮次
这是第 %d 轮补充，请确保每次补充都能实质性提升内容质量。

必须使用中文回复。`, currentSectionTitle, sectionPlan, currentContent, reflection, researchData, revisionCount+1)

	completion, err := llms.GenerateFromSinglePrompt(ctx, llm, prompt)
	if err != nil {
		return nil, err
	}

	// Clean up the output
	completion = strings.TrimSpace(completion)
	completion = strings.TrimPrefix(completion, "```markdown")
	completion = strings.TrimPrefix(completion, "```")
	completion = strings.TrimSuffix(completion, "```")

	// Update the section content
	oldLength := len(currentContent)
	s.ReportSections[idx] = completion
	newLength := len(completion)

	logf(ctx, "章节「%s」补充完成（%d -> %d 字符，增加 %d 字符）\n",
		currentSectionTitle, oldLength, newLength, newLength-oldLength)

	return s, nil
}

// Replace image placeholders with actual image tags
// Regex matches [IMAGE_X：Title] or [IMAGE_X:Title]
var imgRe = regexp.MustCompile(`\[IMAGE_(\d+)[：:]([^\]]+)\]`)

// ReporterNode compiles the final report from pre-written sections.
func ReporterNode(ctx context.Context, state any) (any, error) {
	s := state.(*State)
	logf(ctx, "--- 报告节点：正在编译最终报告 ---\n")

	// Parse section titles from ReportStructure
	var sectionTitles []string
	if s.ReportStructure != "" {
		lines := strings.Split(s.ReportStructure, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "## ") && !strings.HasPrefix(line, "### ") {
				title := strings.TrimPrefix(line, "## ")
				sectionTitles = append(sectionTitles, title)
			}
		}
	}

	// Build the complete markdown report by combining all sections
	var reportBuilder strings.Builder

	// Add title
	queryToUse := s.RefinedQuery
	if queryToUse == "" {
		queryToUse = s.Request.Query
	}
	reportBuilder.WriteString(fmt.Sprintf("# %s\n\n", queryToUse))
	reportBuilder.WriteString(fmt.Sprintf("*生成时间: 2024年*\n\n"))

	// // Add metadata section
	// reportBuilder.WriteString("---\n\n")
	// reportBuilder.WriteString(fmt.Sprintf("**用户查询**: %s\n\n", s.Request.Query))
	// if s.UserIntent != "" {
	// 	reportBuilder.WriteString(fmt.Sprintf("**用户意图**: %s\n\n", s.UserIntent))
	// }
	// reportBuilder.WriteString("---\n\n")

	// Combine all sections
	for i, sectionContent := range s.ReportSections {
		if sectionContent == "" {
			logf(ctx, "警告: 章节 %d 内容为空，跳过\n", i+1)
			continue
		}

		// Get section title
		title := ""
		if i < len(sectionTitles) {
			title = sectionTitles[i]
		} else {
			title = fmt.Sprintf("章节 %d", i+1)
		}

		// Clean and normalize the section content
		cleanedContent := cleanMarkdownContent(sectionContent)

		// Add section heading and content
		reportBuilder.WriteString(fmt.Sprintf("## %s\n\n", title))
		reportBuilder.WriteString(cleanedContent)
		reportBuilder.WriteString("\n\n")
	}

	// Get the complete markdown
	markdownContent := reportBuilder.String()

	// Handle images if present
	markdownContent = imgRe.ReplaceAllStringFunc(markdownContent, func(match string) string {
		parts := imgRe.FindStringSubmatch(match)
		if len(parts) < 3 {
			return match
		}
		idxStr := parts[1]
		title := strings.TrimSpace(parts[2])

		idx, err := strconv.Atoi(idxStr)
		if err != nil || idx < 1 || idx > len(s.Images) {
			return match
		}

		imgURL := s.Images[idx-1]
		return fmt.Sprintf("\n\n<img src=\"%s\" alt=\"%s\" style=\"max-width: 90%%; display: block; margin: 10px auto;\" />\n\n", imgURL, title)
	})

	// If LLM didn't use placeholders, append images at the end
	if len(s.Images) > 0 && !strings.Contains(markdownContent, "<img") {
		markdownContent += "\n\n## 相关图片\n\n"
		for i, imgURL := range s.Images {
			markdownContent += fmt.Sprintf("<img src=\"%s\" alt=\"图片 %d\" style=\"max-width: 90%%; display: block; margin: 10px auto;\" />\n\n", imgURL, i+1)
		}
	}

	// Final cleanup of the entire markdown document
	markdownContent = cleanMarkdownDocument(markdownContent)

	// Convert Markdown to HTML
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse([]byte(markdownContent))

	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	s.FinalReport = string(markdown.Render(doc, renderer))

	// Calculate total character count
	totalChars := 0
	for _, content := range s.ReportSections {
		totalChars += len(content)
	}

	logf(ctx, "最终报告已编译完成（%d 个章节，共 %d 字符，包含 %d 张图片）。\n",
		len(s.ReportSections), totalChars, len(s.Images))

	return s, nil
}

// cleanMarkdownContent cleans and normalizes markdown content from LLM output
func cleanMarkdownContent(content string) string {
	lines := strings.Split(content, "\n")
	var cleaned []string

	// Remove leading/trailing empty lines
	start := 0
	for start < len(lines) && strings.TrimSpace(lines[start]) == "" {
		start++
	}
	end := len(lines)
	for end > start && strings.TrimSpace(lines[end-1]) == "" {
		end--
	}

	if start >= end {
		return ""
	}

	// Track if we're in a code block
	inCodeBlock := false
	codeBlockFence := ""

	for i := start; i < end; i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		// Handle code blocks
		if strings.HasPrefix(trimmed, "```") {
			if !inCodeBlock {
				// Starting a code block
				inCodeBlock = true
				codeBlockFence = trimmed
				cleaned = append(cleaned, line)
			} else if strings.HasPrefix(trimmed, codeBlockFence) {
				// Ending a code block
				inCodeBlock = false
				cleaned = append(cleaned, line)
			} else {
				// Different fence, treat as content
				cleaned = append(cleaned, line)
			}
			continue
		}

		// If inside code block, keep as-is
		if inCodeBlock {
			cleaned = append(cleaned, line)
			continue
		}

		// Remove markdown code block markers if they wrap the entire content
		if trimmed == "```markdown" || trimmed == "```" || trimmed == "```md" {
			// Skip these standalone markers
			continue
		}

		// Clean up the line but preserve structure
		// Remove excessive empty lines (will be handled later)
		cleaned = append(cleaned, line)
	}

	// Join lines and normalize spacing
	result := strings.Join(cleaned, "\n")

	// Remove multiple consecutive empty lines (keep max 2)
	result = regexp.MustCompile(`\n{3,}`).ReplaceAllString(result, "\n\n")

	// Fix common formatting issues
	// Ensure proper spacing around headings
	result = regexp.MustCompile(`([^\n])\n(#{1,6})`).ReplaceAllString(result, "$1\n\n$2")
	result = regexp.MustCompile(`(#{1,6}[^\n]+)\n([^#\s])`).ReplaceAllString(result, "$1\n\n$2")

	// Ensure proper spacing around lists
	result = regexp.MustCompile(`([^\n])\n([-\*]\s)`).ReplaceAllString(result, "$1\n\n$2")
	result = regexp.MustCompile(`([^\n])\n(\d+\.\s)`).ReplaceAllString(result, "$1\n\n$2")

	// Fix bold formatting issues
	result = regexp.MustCompile(`\*\*([^*]+)\*\*`).ReplaceAllString(result, "**$1**")

	// Fix inline code issues - use double quotes to escape backticks
	result = regexp.MustCompile("`([^`]+)`").ReplaceAllString(result, "`$1`")

	return result
}

// cleanMarkdownDocument performs final cleanup on the entire markdown document
func cleanMarkdownDocument(doc string) string {
	// Remove document-level markdown code fences
	doc = strings.TrimSpace(doc)
	doc = strings.TrimPrefix(doc, "```markdown")
	doc = strings.TrimPrefix(doc, "```md")
	doc = strings.TrimPrefix(doc, "```")
	doc = strings.TrimSuffix(doc, "```")
	doc = strings.TrimSpace(doc)

	// Ensure document starts with a single # (not ## or ###)
	lines := strings.Split(doc, "\n")
	if len(lines) > 0 {
		firstLine := strings.TrimSpace(lines[0])
		if strings.HasPrefix(firstLine, "#") && !strings.HasPrefix(firstLine, "# ") {
			// Fix headings like ## or ### at the start to #
			for i := 0; i < len(firstLine); i++ {
				if firstLine[i] != '#' {
					lines[0] = "# " + strings.TrimSpace(firstLine[i:])
					break
				}
			}
		}
	}

	// Normalize line endings
	doc = strings.Join(lines, "\n")

	// Remove excessive empty lines at document level
	doc = regexp.MustCompile(`\n{4,}`).ReplaceAllString(doc, "\n\n\n")

	return doc
}

// PodcastNode generates a podcast script based on the research results.
func PodcastNode(ctx context.Context, state any) (any, error) {
	s := state.(*State)
	logf(ctx, "--- 播客节点：正在生成播客脚本 ---\n")

	llm, err := getLLM()
	if err != nil {
		return nil, err
	}

	researchData := strings.Join(s.ResearchResults, "\n\n")
	prompt := fmt.Sprintf(`你是一名专业的播客制作人。请根据以下研究结果，创作一段引人入胜的播客对话脚本。
对话应该由两名主持人（Host 1 和 Host 2）进行，风格轻松幽默，通俗易懂。
请深入讨论研究结果中的关键点，并加入一些生动的例子或类比。

请以 JSON 格式返回结果，格式如下：
{
    "title": "播客标题",
    "lines": [
        {"speaker": "Host 1", "content": "对话内容..."},
        {"speaker": "Host 2", "content": "对话内容..."}
    ]
}

研究结果：
%s

原始查询：%s
必须使用中文创作。`, researchData, s.Request.Query)

	completion, err := llms.GenerateFromSinglePrompt(ctx, llm, prompt)
	if err != nil {
		return nil, err
	}

	// Clean up JSON
	completion = strings.TrimSpace(completion)
	completion = strings.TrimPrefix(completion, "```json")
	completion = strings.TrimPrefix(completion, "```")
	completion = strings.TrimSuffix(completion, "```")
	completion = strings.TrimSpace(completion)

	var script struct {
		Title string `json:"title"`
		Lines []struct {
			Speaker string `json:"speaker"`
			Content string `json:"content"`
		} `json:"lines"`
	}

	if err := json.Unmarshal([]byte(completion), &script); err != nil {
		logf(ctx, "播客脚本 JSON 解析失败 (%v)，使用原始文本\n", err)
		s.PodcastScript = fmt.Sprintf("<pre>%s</pre>", completion)
		return s, nil
	}

	// Serialize script back to JSON for export
	jsonBytes, _ := json.Marshal(script)
	jsonString := string(jsonBytes)
	jsonString = strings.ReplaceAll(jsonString, "</div>", "<\\/div>") // Escape for HTML embedding

	// Render HTML
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(`
<div class="podcast-container" style="max-width: 800px; margin: 0 auto; font-family: 'Inter', sans-serif;">
    <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px;">
        <h2 style="margin: 0;">%s</h2>
        <button onclick="window.exportPodcastJson()" style="background-color: #28a745; color: white; border: none; padding: 8px 16px; border-radius: 4px; cursor: pointer; display: flex; align-items: center; gap: 5px;">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path><polyline points="7 10 12 15 17 10"></polyline><line x1="12" y1="15" x2="12" y2="3"></line></svg>
            导出 JSON 脚本
        </button>
    </div>
    <div id="podcastJsonData" style="display:none">%s</div>
`, script.Title, jsonString))

	for _, line := range script.Lines {
		speakerClass := "host-1"
		bgColor := "#e6f7ff"
		borderColor := "#1890ff"
		textColor := "#0050b3"

		if strings.Contains(strings.ToLower(line.Speaker), "2") {
			speakerClass = "host-2"
			bgColor = "#fff0f6"
			borderColor = "#eb2f96"
			textColor = "#9e1068"
		}

		sb.WriteString(fmt.Sprintf(`
    <div class="podcast-message %s" style="margin-bottom: 20px; padding: 20px; border-radius: 8px; border-left: 5px solid %s; background-color: %s; box-shadow: 0 2px 5px rgba(0,0,0,0.05);">
        <div class="speaker-name" style="font-weight: 700; margin-bottom: 8px; color: %s; text-transform: uppercase; letter-spacing: 0.5px;">%s</div>
        <div class="message-content" style="line-height: 1.6; color: #333; font-size: 16px;">%s</div>
    </div>
`, speakerClass, borderColor, bgColor, textColor, line.Speaker, line.Content))
	}

	sb.WriteString("</div>")

	s.PodcastScript = sb.String()
	logf(ctx, "播客脚本已生成。\n")
	return s, nil
}

func getLLM() (llms.Model, error) {
	// Use DeepSeek as per user preference
	// Ensure OPENAI_API_KEY and OPENAI_API_BASE are set in the environment
	return openai.New()
}
