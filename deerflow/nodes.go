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
	logf(ctx, "--- 规划节点：正在为查询 '%s' 进行规划 ---\n", s.Request.Query)

	llm, err := getLLM()
	if err != nil {
		return nil, err
	}

	// 使用优化后的查询进行规划
	queryToUse := s.RefinedQuery
	if queryToUse == "" {
		queryToUse = s.Request.Query
	}

	prompt := fmt.Sprintf(`你是一名研究规划师。请为以下查询创建一个分步研究计划：%s

用户意图识别: %s
实体信息: %s

同时，请判断用户是否希望同时生成播客（Podcast）脚本（例如查询中包含"播客"、"podcast"、"对话"、"脚本"等意图，或者用户明确要求生成播客）。
请以 JSON 格式返回结果，格式如下：
{
    "plan": ["步骤1", "步骤2", ...],
    "generate_podcast": true/false
}
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
		Plan            []string `json:"plan"`
		GeneratePodcast bool     `json:"generate_podcast"`
	}

	if err := json.Unmarshal([]byte(completion), &output); err != nil {
		logf(ctx, "JSON 解析失败 (%v)，尝试简单解析\n", err)
		// Fallback: simple parsing
		lines := strings.Split(completion, "\n")
		var plan []string
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if trimmed != "" && !strings.HasPrefix(trimmed, "{") && !strings.HasPrefix(trimmed, "}") {
				plan = append(plan, trimmed)
			}
		}
		s.Plan = plan
		// Default to false if JSON parsing fails, unless we find keywords in query
		queryLower := strings.ToLower(s.Request.Query)
		s.GeneratePodcast = strings.Contains(queryLower, "播客") || strings.Contains(queryLower, "podcast")
	} else {
		s.Plan = output.Plan
		s.GeneratePodcast = output.GeneratePodcast
	}

	// Format plan for better readability
	var formattedPlan strings.Builder
	formattedPlan.WriteString("生成的计划：\n")
	for _, step := range s.Plan {
		formattedPlan.WriteString(fmt.Sprintf("%s\n", step))
	}
	logf(ctx, "%s", formattedPlan.String())
	if s.GeneratePodcast {
		logf(ctx, "检测到播客生成意图。\n")
	}

	return s, nil
}

// ResearcherNode executes the research plan using LLM.
func ResearcherNode(ctx context.Context, state any) (any, error) {
	s := state.(*State)
	logf(ctx, "--- 研究节点：正在执行计划（使用 LLM） ---\n")

	llm, err := getLLM()
	if err != nil {
		return nil, err
	}

	var results []string
	for _, step := range s.Plan {
		logf(ctx, "正在研究步骤：%s\n", step)
		prompt := fmt.Sprintf("你是一名研究员。请为这个研究步骤查找详细信息：%s。提供发现摘要。必须使用中文回复。", step)
		completion, err := llms.GenerateFromSinglePrompt(ctx, llm, prompt)
		if err != nil {
			return nil, err
		}
		results = append(results, fmt.Sprintf("Step: %s\nFindings: %s", step, completion))
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

	var insights []string
	insights = append(insights, completion)
	s.InsightResults = insights

	logf(ctx, "深度洞察分析完成。\n")
	return s, nil
}

// Replace image placeholders with actual image tags
// Regex matches [IMAGE_X：Title] or [IMAGE_X:Title]
var imgRe = regexp.MustCompile(`\[IMAGE_(\d+)[：:]([^\]]+)\]`)

// ReporterNode compiles the final report.
func ReporterNode(ctx context.Context, state any) (any, error) {
	s := state.(*State)
	logf(ctx, "--- 报告节点：正在生成最终报告 ---\n")

	llm, err := getLLM()
	if err != nil {
		return nil, err
	}

	researchData := strings.Join(s.ResearchResults, "\n\n")

	// 包含洞察分析结果
	insightData := ""
	if len(s.InsightResults) > 0 {
		insightData = "\n\n## 深度洞察分析\n\n" + strings.Join(s.InsightResults, "\n\n")
	}

	// Inform LLM about available images
	imageInfo := ""
	if len(s.Images) > 0 {
		imageInfo = fmt.Sprintf("\n\n注意：研究过程中收集到 %d 张相关图片。在报告中适当的位置，你可以使用 [IMAGE_X：图片标题] 占位符来标记应该插入图片的位置（X 为 1 到 %d，图片标题为你为该图片起的标题）。例如：[IMAGE_1：某某图表]。请务必确保引用的图片与周围的文字内容高度相关，如果图片与当前段落无关，请不要强行插入。", len(s.Images), len(s.Images))
	}

	// 构建完整的报告提示词
	var prompt strings.Builder
	if len(s.InsightResults) > 0 {
		prompt.WriteString("你是一名资深报告撰写员。请根据以下研究结果和深度洞察分析，撰写一份全面、详细的最终报告。\n\n")
		prompt.WriteString(fmt.Sprintf("## 用户查询\n%s\n\n", s.Request.Query))
		prompt.WriteString(fmt.Sprintf("## 用户意图\n%s\n\n", s.UserIntent))
		prompt.WriteString(fmt.Sprintf("## 实体信息\n%s\n\n", s.EntityInfo))
		prompt.WriteString("## 研究结果\n")
		prompt.WriteString(researchData)
		prompt.WriteString("\n\n")
		prompt.WriteString(insightData)
		prompt.WriteString("\n\n## 你的任务\n")
		prompt.WriteString("请整合以上信息，撰写一份结构完整、内容详实的专业报告。报告应该包含：\n\n")
		prompt.WriteString("1. **执行摘要**: 简要概述研究主题和核心发现\n")
		prompt.WriteString("2. **详细内容**: 基于研究结果展开详细分析\n")
		prompt.WriteString("3. **深度洞察**: 整合深度洞察分析的关键观点\n")
		prompt.WriteString("4. **结论建议**: 提供清晰的结论和实用建议\n\n")
		prompt.WriteString("使用 Markdown 格式，包含清晰的标题、要点，并在适当的地方使用代码块。数学公式请使用 ```math 代码块包裹，或者使用 $$...$$ (块级) 和 $...$ (行内) 包裹。不要透漏撰写人信息。")
		prompt.WriteString(imageInfo)
		prompt.WriteString("\n\n必须使用中文撰写报告，确保报告内容详实、逻辑清晰、洞察深刻。")
	} else {
		prompt.WriteString("你是一名资深报告撰写员。请根据以下研究结果撰写一份全面的最终报告。使用 Markdown 格式，包含清晰的标题、要点，并在适当的地方使用代码块。数学公式请使用 ```math 代码块包裹，或者使用 $$...$$ (块级) 和 $...$ (行内) 包裹。不要透漏撰写人信息。")
		prompt.WriteString(imageInfo)
		prompt.WriteString("必须使用中文撰写报告：\n\n")
		prompt.WriteString(researchData)
		prompt.WriteString(fmt.Sprintf("\n\n原始查询是：%s", s.Request.Query))
	}

	completion, err := llms.GenerateFromSinglePrompt(ctx, llm, prompt.String())
	if err != nil {
		return nil, err
	}

	// Convert Markdown to HTML
	// Clean up markdown code blocks if present
	completion = strings.TrimPrefix(completion, "```markdown")
	completion = strings.TrimPrefix(completion, "```")
	completion = strings.TrimSuffix(completion, "```")

	completion = imgRe.ReplaceAllStringFunc(completion, func(match string) string {
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
	if len(s.Images) > 0 && !strings.Contains(completion, "<img") {
		completion += "\n\n## 相关图片\n\n"
		for i, imgURL := range s.Images {
			completion += fmt.Sprintf("<img src=\"%s\" alt=\"图片 %d\" style=\"max-width: 90%%; display: block; margin: 10px auto;\" />\n\n", imgURL, i+1)
		}
	}

	extensions := parser.CommonExtensions | parser.AutoHeadingIDs
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse([]byte(completion))

	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	s.FinalReport = string(markdown.Render(doc, renderer))
	logf(ctx, "最终报告已生成（包含 %d 张图片）。\n", len(s.Images))
	return s, nil
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
