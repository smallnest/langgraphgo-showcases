package forum_engine

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/smallnest/langgraphgo-showcases/DeepInsight/schema"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

func getLLM() (llms.Model, error) {
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

const (
	SystemPromptModerator = `你是一个深度研究系统的专家讨论会主持人。你的职责是：

1. **研究梳理**：从各专家的发言中自动识别关键研究问题、理论观点、实证发现，按逻辑顺序整理研究脉络
2. **引导讨论**：根据各专家的发言，引导深入讨论关键问题，探究深层次的机制和原因
3. **纠正偏差**：结合不同专家的视角以及观点，如果发现逻辑矛盾或证据不足，请明确指出
4. **整合观点**：综合不同专家的视角，形成更全面的认识，找出共识和分歧
5. **趋势判断**：基于已有信息分析研究发展趋势，提出可能的研究方向
6. **推进分析**：提出新的分析角度或需要关注的问题，引导后续讨论方向

**专家介绍**：
- **INSIGHT Expert**：专注于专家观点和深度洞察分析，提供理论支撑和专家观点
- **MEDIA Expert**：擅长多模态内容分析，关注视觉信息、图表、数据可视化等
- **QUERY Expert**：负责精准信息搜索，提供最新的研究资料和实证数据

**发言要求**：
1. **综合性**：每次发言控制在1000字以内，内容应包括研究梳理、观点整合、问题引导等多个方面
2. **结构清晰**：使用明确的段落结构，包括研究梳理、观点对比、问题提出等部分
3. **深入分析**：不仅仅总结已有信息，还要提出深层次的见解和分析
4. **客观严谨**：基于事实和证据进行分析和判断，避免主观臆测
5. **前瞻性**：提出具有前瞻性的观点和建议，引导讨论向更深入的方向发展`

	SystemPromptQueryExpert = `你是 "Query Expert" (研究查询专家)。你负责精准信息搜索，提供最新的研究资料和实证数据。
你的发言应该严谨、客观，引用具体的研究内容。如果其他专家的观点缺乏实证支持，你应该提出质疑。`

	SystemPromptMediaExpert = `你是 "Media Expert" (多媒体分析专家)。你擅长多模态内容分析，关注视觉信息、图表、数据可视化等。
你的发言应该关注数据图表、可视化内容传达的信息和研究价值。你可以补充 Query Expert 忽略的视觉和多媒体细节。`

	SystemPromptInsightExpert = `你是 "Insight Expert" (洞察分析专家)。你专注于专家观点和深度洞察分析，提供理论支撑和专家观点。
你的发言应该关注理论框架、专家观点和深层次洞察。你可以从理论角度补充其他专家的分析。`
)

// ForumEngineNode simulates a multi-expert discussion.
func ForumEngineNode(ctx context.Context, state any) (any, error) {
	s := state.(*schema.DeepInsightState)
	fmt.Println("ForumEngine: 正在启动专家多轮讨论...")

	llm, err := getLLM()
	if err != nil {
		return nil, err
	}

	// Context for all experts
	researchSummary := "研究报告摘要:\n"
	if len(s.ResearchResults) > 0 {
		// Take the first 1000 chars to avoid token limits if report is huge
		runes := []rune(s.ResearchResults[0])
		if len(runes) > 1000 {
			researchSummary += string(runes[:1000]) + "..."
		} else {
			researchSummary += string(runes)
		}
	}

	mediaSummary := "多媒体发现摘要:\n" + strings.Join(s.MediaResults, "\n")
	insightSummary := "洞察分析摘要:\n" + strings.Join(s.InsightResults, "\n")

	contextInfo := fmt.Sprintf("研究主题: %s\n\n%s\n\n%s\n\n%s", s.Query, researchSummary, mediaSummary, insightSummary)

	// Conversation History
	var history []string

	// Define the conversation flow
	turns := []struct {
		Speaker string
		Prompt  string
	}{
		{"Moderator", SystemPromptModerator},
		{"QueryExpert", SystemPromptQueryExpert},
		{"MediaExpert", SystemPromptMediaExpert},
		{"InsightExpert", SystemPromptInsightExpert},
		{"QueryExpert", SystemPromptQueryExpert},
		{"Moderator", SystemPromptModerator}, // Final summary
	}

	for i, turn := range turns {
		fmt.Printf("  Round %d: %s speaking...\n", i+1, turn.Speaker)

		// Build conversation history string
		historyStr := strings.Join(history, "\n\n")

		var userContent string
		if turn.Speaker == "Moderator" {
			// Use the prompt for the Moderator
			userContent = fmt.Sprintf(`最近的专家发言记录：
%s

请你作为讨论会主持人，基于以上专家的发言进行综合分析，请按以下结构组织你的发言：

**一、研究梳理与逻辑分析**
- 从各专家发言中自动识别关键研究问题、理论观点、实证发现
- 按逻辑顺序整理研究脉络，梳理因果关系
- 指出关键转折点和重要研究节点

**二、观点整合与对比分析**
- 综合QUERY、MEDIA、INSIGHT三个专家的视角和发现
- 指出不同信息源之间的共识与分歧
- 分析每个专家的信息价值和互补性
- 如果发现逻辑矛盾或证据不足，请明确指出并给出理由

**三、深层次分析与趋势判断**
- 基于已有信息分析研究的深层次原因和影响因素
- 预测研究发展趋势，指出可能的研究方向
- 提出需要特别关注的方面和指标

**四、问题引导与讨论方向**
- 提出2-3个值得进一步深入探讨的关键问题
- 为后续研究提出具体的建议和方向
- 引导各专家关注特定的数据维度或分析角度

请发表综合性的主持人发言（控制在1000字以内），内容应包含以上四个部分，并保持逻辑清晰、分析深入、视角独特。`, historyStr)

			// If history is empty (first round), we need to provide the initial context
			if len(history) == 0 {
				userContent = fmt.Sprintf(`初始背景信息：
%s

请你作为讨论会主持人，开启讨论。`, contextInfo)
			}

		} else {
			// For other experts, keep the simple prompt but update the context
			userContent = fmt.Sprintf(`背景信息:
%s

之前的讨论:
%s

轮到你了 (%s)。请发表你的观点 (200字以内):`, contextInfo, historyStr, turn.Speaker)
		}

		messages := []llms.MessageContent{
			llms.TextParts(llms.ChatMessageTypeSystem, turn.Prompt),
			llms.TextParts(llms.ChatMessageTypeHuman, userContent),
		}

		completion, err := llm.GenerateContent(ctx, messages)
		if err != nil {
			fmt.Printf("  Error generating response for %s: %v\n", turn.Speaker, err)
			continue
		}

		response := completion.Choices[0].Content
		// Clean up markdown code blocks
		response = strings.TrimPrefix(response, "```markdown")
		response = strings.TrimPrefix(response, "```md")
		response = strings.TrimPrefix(response, "```")
		response = strings.TrimSuffix(response, "```")
		response = strings.TrimSpace(response)

		entry := fmt.Sprintf("[%s] %s:\n%s", time.Now().Format("15:04:05"), turn.Speaker, response)
		history = append(history, entry)
		fmt.Printf("    -> %s\n", response)
	}

	s.Discussion = history

	fmt.Println("ForumEngine: 讨论完成。")
	return s, nil
}
