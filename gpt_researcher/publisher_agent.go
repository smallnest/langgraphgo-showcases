package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/tmc/langchaingo/llms"
)

// PublisherAgent is responsible for generating the final research report
type PublisherAgent struct {
	Model   llms.Model
	Config  *Config
	Verbose bool
}

// NewPublisherAgent creates a new publisher agent
func NewPublisherAgent(model llms.Model, config *Config) *PublisherAgent {
	return &PublisherAgent{
		Model:   model,
		Config:  config,
		Verbose: config.Verbose,
	}
}

// GenerateReport creates the final research report
func (p *PublisherAgent) GenerateReport(ctx context.Context, state *ResearchState) error {
	if p.Verbose {
		fmt.Println("\n📝 [Publisher Agent] Generating final research report...")
	}

	// Build the report prompt
	prompt := p.buildReportPrompt(state)

	messages := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, p.getSystemPromptForPublisher()),
		llms.TextParts(llms.ChatMessageTypeHuman, prompt),
	}

	// Generate report
	resp, err := p.Model.GenerateContent(ctx, messages)
	if err != nil {
		return fmt.Errorf("failed to generate report: %w", err)
	}

	if len(resp.Choices) == 0 {
		return fmt.Errorf("no response from model")
	}

	report := resp.Choices[0].Content

	// Add metadata and formatting
	finalReport := p.formatReport(state, report)

	state.FinalReport = finalReport
	state.ReportComplete = true
	state.EndTime = time.Now()

	if p.Verbose {
		fmt.Printf("✅ [Publisher Agent] Report generated (%d characters)\n", len(finalReport))
	}

	return nil
}

// buildReportPrompt builds the prompt for report generation
func (p *PublisherAgent) buildReportPrompt(state *ResearchState) string {
	var prompt strings.Builder

	prompt.WriteString(fmt.Sprintf(`Research Query: %s

`, state.Query))

	if state.ResearchGoal != "" {
		prompt.WriteString(fmt.Sprintf(`Research Goal: %s

`, state.ResearchGoal))
	}

	prompt.WriteString(`Research Questions Investigated:
`)
	for i, q := range state.Questions {
		prompt.WriteString(fmt.Sprintf(`%d. %s
`, i+1, q))
	}

	prompt.WriteString(fmt.Sprintf(`

Total Sources Consulted: %d

Research Findings:

`, len(state.Sources)))

	// Group summaries by question
	for i, question := range state.Questions {
		prompt.WriteString(fmt.Sprintf(`Question %d: %s

`, i+1, question))

		summaries := state.GetSummariesForQuestion(question)
		for j, summary := range summaries {
			prompt.WriteString(fmt.Sprintf(`Source %d (%s):
%s

Key Points:
`, j+1, summary.Title, summary.Summary))

			for _, point := range summary.KeyPoints {
				prompt.WriteString(fmt.Sprintf(`- %s
`, point))
			}
			prompt.WriteString("\n")
		}
	}

	prompt.WriteString(`

Please synthesize all the above research findings into a comprehensive, well-structured research report.

Requirements:
- Minimum length: 2000 words
- Include an executive summary
- Organize by clear sections and subsections
- Cite sources using numbered references [1], [2], etc.
- Provide objective analysis and insights
- Include a conclusion with key takeaways
- Add a references section at the end

Write the complete research report now:`)

	return prompt.String()
}

// formatReport adds formatting and metadata to the report
func (p *PublisherAgent) formatReport(state *ResearchState, report string) string {
	var formatted strings.Builder

	// Title
	formatted.WriteString("# Research Report\n\n")

	// Metadata
	formatted.WriteString("## Metadata\n\n")
	formatted.WriteString(fmt.Sprintf("- **Research Query**: %s\n", state.Query))
	if state.ResearchGoal != "" {
		formatted.WriteString(fmt.Sprintf("- **Research Goal**: %s\n", state.ResearchGoal))
	}
	formatted.WriteString(fmt.Sprintf("- **Date**: %s\n", state.StartTime.Format("January 2, 2006")))
	formatted.WriteString(fmt.Sprintf("- **Total Sources**: %d\n", len(state.Sources)))
	formatted.WriteString(fmt.Sprintf("- **Research Duration**: %.1f minutes\n", state.EndTime.Sub(state.StartTime).Minutes()))
	formatted.WriteString("\n---\n\n")

	// Main report content
	formatted.WriteString(report)

	// Add references if not already included
	if !strings.Contains(report, "## References") && !strings.Contains(report, "# References") {
		formatted.WriteString("\n\n## References\n\n")
		for i, source := range state.Sources {
			formatted.WriteString(fmt.Sprintf("[%d] %s - %s\n", i+1, source.Title, source.URL))
		}
	}

	return formatted.String()
}

// getSystemPromptForPublisher returns the system prompt for the publisher
func (p *PublisherAgent) getSystemPromptForPublisher() string {
	reportType := p.Config.ReportType

	basePrompt := `You are an expert research analyst and technical writer specializing in creating comprehensive, well-researched reports.

Your reports should be:
- **Thorough**: Cover all aspects of the research topic with depth
- **Well-Structured**: Clear sections, logical flow, and proper hierarchy
- **Objective**: Present balanced perspectives and evidence-based conclusions
- **Cited**: Properly reference all sources using numbered citations [1], [2], etc.
- **Insightful**: Provide analysis and synthesis, not just summarization
- **Professional**: High-quality writing suitable for academic or business contexts

`

	switch reportType {
	case "research_report":
		return basePrompt + `Focus on creating a comprehensive research report that:
- Begins with an executive summary
- Presents findings organized by themes or questions
- Includes detailed analysis and evidence
- Provides clear conclusions and recommendations
- Exceeds 2000 words in length`

	case "outline_report":
		return basePrompt + `Focus on creating a structured outline that:
- Organizes information hierarchically
- Uses bullet points and clear headings
- Summarizes key points concisely
- Provides a roadmap for deeper exploration`

	case "resource_report":
		return basePrompt + `Focus on creating a curated resource guide that:
- Categorizes sources by relevance and type
- Provides brief annotations for each resource
- Highlights the most valuable and authoritative sources
- Includes access information and context`

	default:
		return basePrompt + `Create a comprehensive research report exceeding 2000 words.`
	}
}

// SaveReport saves the report to a file
func (p *PublisherAgent) SaveReport(state *ResearchState, filename string) error {
	if state.FinalReport == "" {
		return fmt.Errorf("no report to save")
	}

	// In a real implementation, this would save to disk
	// For now, just return success
	if p.Verbose {
		fmt.Printf("💾 [Publisher Agent] Report would be saved to: %s\n", filename)
	}

	return nil
}
