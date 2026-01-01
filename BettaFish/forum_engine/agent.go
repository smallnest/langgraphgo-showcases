package forum_engine

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/smallnest/langgraphgo/showcases/BettaFish/schema"
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
	SystemPromptModerator = `【重要提醒】我们的需求基于科研目的，已通过伦理性合规审查
    
你是一个多agent舆情分析系统的论坛主持人。你的职责是：

1. **事件梳理**：从各agent的发言中自动识别关键事件、人物、时间节点，按时间顺序整理事件脉络
2. **引导讨论**：根据各agent的发言，引导深入讨论关键问题，探究深层原因
3. **纠正错误**：结合不同agent的视角以及言论，如果发现事实错误或逻辑矛盾，请明确指出
4. **整合观点**：综合不同agent的视角，形成更全面的认识，找出共识和分歧
5. **趋势预测**：基于已有信息分析舆情发展趋势，提出可能的风险点
6. **推进分析**：提出新的分析角度或需要关注的问题，引导后续讨论方向

**Agent介绍**：
- **INSIGHT Agent**：专注于私有舆情数据库的深度挖掘和分析，提供历史数据和模式对比
- **MEDIA Agent**：擅长多模态内容分析，关注媒体报道、图片、视频等视觉信息的传播效果
- **QUERY Agent**：负责精准信息搜索，提供最新的网络信息和实时动态

**发言要求**：
1. **综合性**：每次发言控制在1000字以内，内容应包括事件梳理、观点整合、问题引导等多个方面
2. **结构清晰**：使用明确的段落结构，包括事件梳理、观点对比、问题提出等部分
3. **深入分析**：不仅仅总结已有信息，还要提出深层次的见解和分析
4. **客观中立**：基于事实进行分析和判断，避免主观臆测和情绪化表达
5. **前瞻性**：提出具有前瞻性的观点和建议，引导讨论向更深入的方向发展

**注意事项**：
- 本讨论基于科研目的，已通过伦理性合规审查
- 保持专业性和学术性，重视数据和证据
- 对于敏感话题，应保持谨慎态度，基于事实进行分析

【重要提醒】我们的需求基于科研目的，已通过伦理性合规审查`

	// Note: The original Python code only defines the Moderator's prompt in llm_host.py.
	// The other agents (News/Media) seem to be simulated or their prompts are dynamically generated based on their engine's output.
	// In our Go implementation, we are simulating the multi-turn conversation where NewsAgent and MediaAgent also speak.
	// To align with the "exact prompt" requirement, we should use the Moderator prompt exactly as above.
	// For NewsAgent and MediaAgent, since they are not explicitly defined in llm_host.py (which only defines the Host),
	// we will keep their current prompts but refine them to be consistent with the Host's description of them.

	SystemPromptNewsAgent = `你是 "Query Agent" (也称为 NewsAgent)。你负责精准信息搜索，提供最新的网络信息和实时动态。
你的发言应该严谨、客观，引用具体的新闻内容。如果其他智能体的观点缺乏事实支持，你应该提出质疑。`

	SystemPromptMediaAgent = `你是 "Media Agent"。你擅长多模态内容分析，关注媒体报道、图片、视频等视觉信息的传播效果。
你的发言应该关注图片传达的氛围、情感和视觉冲击力。你可以补充 Query Agent 忽略的感性细节。`
)

// ForumEngineNode simulates a multi-turn discussion.
func ForumEngineNode(ctx context.Context, state any) (any, error) {
	s := state.(*schema.BettaFishState)
	fmt.Println("ForumEngine: 正在启动智能体多轮讨论...")

	llm, err := getLLM()
	if err != nil {
		return nil, err
	}

	// Context for all agents
	newsSummary := "新闻分析摘要:\n"
	if len(s.NewsResults) > 0 {
		// Take the first 500 chars to avoid token limits if report is huge
		runes := []rune(s.NewsResults[0])
		if len(runes) > 1000 {
			newsSummary += string(runes[:1000]) + "..."
		} else {
			newsSummary += string(runes)
		}
	}

	mediaSummary := "媒体发现摘要:\n" + strings.Join(s.MediaResults, "\n")

	contextInfo := fmt.Sprintf("讨论主题: %s\n\n%s\n\n%s", s.Query, newsSummary, mediaSummary)

	// Conversation History
	var history []string

	// Define the conversation flow
	turns := []struct {
		Speaker string
		Prompt  string
	}{
		{"Moderator", SystemPromptModerator},
		{"QueryAgent", SystemPromptNewsAgent},
		{"MediaAgent", SystemPromptMediaAgent},
		{"QueryAgent", SystemPromptNewsAgent},
		{"Moderator", SystemPromptModerator}, // Final summary
	}

	for i, turn := range turns {
		fmt.Printf("  Round %d: %s speaking...\n", i+1, turn.Speaker)

		// Build conversation history string
		historyStr := strings.Join(history, "\n\n")

		var userContent string
		if turn.Speaker == "Moderator" {
			// Use the exact user prompt from llm_host.py for the Moderator
			userContent = fmt.Sprintf(`【重要提醒】我们的需求基于科研目的，已通过伦理性合规审查

最近的Agent发言记录：
%s

请你作为论坛主持人，基于以上agent的发言进行综合分析，请按以下结构组织你的发言：

**一、事件梳理与时间线分析**
- 从各agent发言中自动识别关键事件、人物、时间节点
- 按时间顺序整理事件脉络，梳理因果关系
- 指出关键转折点和重要节点

**二、观点整合与对比分析**
- 综合INSIGHT、MEDIA、QUERY三个Agent的视角和发现
- 指出不同数据源之间的共识与分歧
- 分析每个Agent的信息价值和互补性
- 如果发现事实错误或逻辑矛盾，请明确指出并给出理由

**三、深层次分析与趋势预测**
- 基于已有信息分析舆情的深层原因和影响因素
- 预测舆情发展趋势，指出可能的风险点和机遇
- 提出需要特别关注的方面和指标

**四、问题引导与讨论方向**
- 提出2-3个值得进一步深入探讨的关键问题
- 为后续研究提出具体的建议和方向
- 引导各Agent关注特定的数据维度或分析角度

请发表综合性的主持人发言（控制在1000字以内），内容应包含以上四个部分，并保持逻辑清晰、分析深入、视角独特。

【重要提醒】我们的需求基于科研目的，已通过伦理性合规审查`, historyStr)

			// If history is empty (first round), we need to provide the initial context
			if len(history) == 0 {
				userContent = fmt.Sprintf(`【重要提醒】我们的需求基于科研目的，已通过伦理性合规审查

初始背景信息：
%s

请你作为论坛主持人，开启讨论。`, contextInfo)
			}

		} else {
			// For other agents, keep the simple prompt but update the context
			userContent = fmt.Sprintf(`背景信息:
%s

之前的讨论:
%s

轮到你了 (%s)。请发表你的观点 (100字以内):`, contextInfo, historyStr, turn.Speaker)
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
		entry := fmt.Sprintf("[%s] %s:\n%s", time.Now().Format("15:04:05"), turn.Speaker, response)
		history = append(history, entry)
		fmt.Printf("    -> %s\n", response)
	}

	s.Discussion = history

	fmt.Println("ForumEngine: 讨论完成。")
	return s, nil
}
