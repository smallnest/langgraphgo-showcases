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

// PlannerNode generates a research plan based on the query.
func PlannerNode(ctx context.Context, state any) (any, error) {
	s := state.(*State)
	logf(ctx, "--- 规划节点：正在为查询 '%s' 进行规划 ---\n", s.Request.Query)

	llm, err := getLLM()
	if err != nil {
		return nil, err
	}

	prompt := fmt.Sprintf(`你是一名研究规划师。请为以下查询创建一个分步研究计划：%s。
同时，请判断用户是否希望同时生成播客（Podcast）脚本（例如查询中包含"播客"、"podcast"、"对话"、"脚本"等意图，或者用户明确要求生成播客）。
请以 JSON 格式返回结果，格式如下：
{
    "plan": ["步骤1", "步骤2", ...],
    "generate_podcast": true/false
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

	// Inform LLM about available images
	imageInfo := ""
	if len(s.Images) > 0 {
		imageInfo = fmt.Sprintf("\n\n注意：研究过程中收集到 %d 张相关图片。在报告中适当的位置，你可以使用 [IMAGE_X：图片标题] 占位符来标记应该插入图片的位置（X 为 1 到 %d，图片标题为你为该图片起的标题）。例如：[IMAGE_1：某某图表]。请务必确保引用的图片与周围的文字内容高度相关，如果图片与当前段落无关，请不要强行插入。", len(s.Images), len(s.Images))
	}

	prompt := fmt.Sprintf("你是一名资深报告撰写员。请根据以下研究结果撰写一份全面的最终报告。使用 Markdown 格式，包含清晰的标题、要点，并在适当的地方使用代码块。数学公式请使用 ```math 代码块包裹，或者使用 $$...$$ (块级) 和 $...$ (行内) 包裹。不要透漏撰写人信息。%s必须使用中文撰写报告：\n\n%s\n\n原始查询是：%s", imageInfo, researchData, s.Request.Query)

	completion, err := llms.GenerateFromSinglePrompt(ctx, llm, prompt)
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
