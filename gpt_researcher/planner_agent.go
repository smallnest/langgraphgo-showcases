package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/tmc/langchaingo/llms"
)

// PlannerAgent is responsible for generating research questions
type PlannerAgent struct {
	Model   llms.Model
	Config  *Config
	Verbose bool
}

// NewPlannerAgent creates a new planner agent
func NewPlannerAgent(model llms.Model, config *Config) *PlannerAgent {
	return &PlannerAgent{
		Model:   model,
		Config:  config,
		Verbose: config.Verbose,
	}
}

// GenerateQuestions generates research questions based on the query
func (p *PlannerAgent) GenerateQuestions(ctx context.Context, state *ResearchState) error {
	if p.Verbose {
		fmt.Println("\n🎯 [Planner Agent] Generating research questions...")
	}

	prompt := p.buildPlanningPrompt(state.Query, state.ResearchGoal)

	messages := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, getSystemPromptForPlanner()),
		llms.TextParts(llms.ChatMessageTypeHuman, prompt),
	}

	resp, err := p.Model.GenerateContent(ctx, messages)
	if err != nil {
		return fmt.Errorf("failed to generate questions: %w", err)
	}

	if len(resp.Choices) == 0 {
		return fmt.Errorf("no response from model")
	}

	// Parse questions from response
	questions := p.parseQuestions(resp.Choices[0].Content)

	// Limit to max questions
	if len(questions) > p.Config.MaxQuestions {
		questions = questions[:p.Config.MaxQuestions]
	}

	// Add questions to state
	for _, q := range questions {
		state.AddQuestion(q)
	}

	state.PlanningComplete = true

	if p.Verbose {
		fmt.Printf("✅ [Planner Agent] Generated %d research questions:\n", len(questions))
		for i, q := range questions {
			fmt.Printf("   %d. %s\n", i+1, q)
		}
		fmt.Println()
	}

	return nil
}

// buildPlanningPrompt builds the prompt for generating research questions
func (p *PlannerAgent) buildPlanningPrompt(query string, goal string) string {
	prompt := fmt.Sprintf(`Research Query: %s`, query)

	if goal != "" {
		prompt += fmt.Sprintf(`
Research Goal: %s`, goal)
	}

	prompt += fmt.Sprintf(`

Generate %d comprehensive research questions that will help thoroughly investigate this topic.
Each question should:
1. Focus on a specific aspect of the topic
2. Be answerable through web research
3. Contribute to building an objective understanding
4. Cover different perspectives and dimensions

Format your response as a numbered list of questions.`, p.Config.MaxQuestions)

	return prompt
}

// parseQuestions extracts questions from the model's response
func (p *PlannerAgent) parseQuestions(response string) []string {
	var questions []string

	// Split by lines
	lines := strings.SplitSeq(response, "\n")

	for line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Remove numbering (e.g., "1.", "1)", "Q1:", etc.)
		line = strings.TrimPrefix(line, "- ")
		for i := 1; i <= 20; i++ {
			prefix := fmt.Sprintf("%d.", i)
			if after, ok := strings.CutPrefix(line, prefix); ok {
				line = after
				line = strings.TrimSpace(line)
				break
			}
			prefix = fmt.Sprintf("%d)", i)
			if after, ok := strings.CutPrefix(line, prefix); ok {
				line = after
				line = strings.TrimSpace(line)
				break
			}
			prefix = fmt.Sprintf("Q%d:", i)
			if after, ok := strings.CutPrefix(line, prefix); ok {
				line = after
				line = strings.TrimSpace(line)
				break
			}
		}

		// Check if it looks like a question
		if len(line) > 10 && (strings.HasSuffix(line, "?") || strings.Contains(line, "what") ||
			strings.Contains(line, "how") || strings.Contains(line, "why") ||
			strings.Contains(line, "when") || strings.Contains(line, "where") ||
			strings.Contains(line, "who") || strings.Contains(line, "which")) {
			questions = append(questions, line)
		}
	}

	return questions
}

// getSystemPromptForPlanner returns the system prompt for the planner agent
func getSystemPromptForPlanner() string {
	return `You are a research planning expert. Your role is to generate comprehensive,
well-structured research questions that will guide a thorough investigation of the given topic.

Your questions should:
- Cover multiple aspects and perspectives of the topic
- Be specific and focused (not too broad)
- Be answerable through web research and credible sources
- Build upon each other to create a complete understanding
- Avoid redundancy and overlap
- Progress from foundational to more advanced aspects

Generate questions that collectively will enable the creation of a comprehensive,
objective research report on the topic.`
}
