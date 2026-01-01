package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/tmc/langchaingo/llms"
)

// ExecutionAgent is responsible for gathering information for research questions
type ExecutionAgent struct {
	Model   llms.Model
	Config  *Config
	Tools   *ToolRegistry
	Verbose bool
}

// NewExecutionAgent creates a new execution agent
func NewExecutionAgent(model llms.Model, config *Config, tools *ToolRegistry) *ExecutionAgent {
	return &ExecutionAgent{
		Model:   model,
		Config:  config,
		Tools:   tools,
		Verbose: config.Verbose,
	}
}

// ResearchQuestion executes research for a single question
func (e *ExecutionAgent) ResearchQuestion(ctx context.Context, state *ResearchState, question string) error {
	if e.Verbose {
		fmt.Printf("\n🔍 [Execution Agent] Researching: %s\n", question)
	}

	// Step 1: Search for information
	searchResults, err := e.searchForQuestion(ctx, question)
	if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}

	if e.Verbose {
		fmt.Printf("   Found %d search results\n", len(searchResults))
	}

	// Step 2: Process and summarize each result
	for i, result := range searchResults {
		if i >= e.Config.MaxSourcesToUse {
			break
		}

		summary, err := e.summarizeSource(ctx, result, question)
		if err != nil {
			if e.Verbose {
				fmt.Printf("   ⚠️ Failed to summarize source %s: %v\n", result.URL, err)
			}
			continue
		}

		// Add to state
		state.AddSearchResult(result)
		state.AddSummary(summary)
		state.AddSource(Source{
			URL:      result.URL,
			Title:    result.Title,
			Citation: fmt.Sprintf("[%d] %s", len(state.Sources)+1, result.Title),
		})

		if e.Verbose {
			fmt.Printf("   ✅ Summarized: %s\n", result.Title)
		}
	}

	return nil
}

// ExecuteAll executes research for all questions
func (e *ExecutionAgent) ExecuteAll(ctx context.Context, state *ResearchState) error {
	if e.Verbose {
		fmt.Printf("\n📚 [Execution Agent] Starting research for %d questions...\n", len(state.Questions))
	}

	for i, question := range state.Questions {
		if e.Verbose {
			fmt.Printf("\n--- Question %d/%d ---\n", i+1, len(state.Questions))
		}

		if err := e.ResearchQuestion(ctx, state, question); err != nil {
			if e.Verbose {
				fmt.Printf("⚠️ Error researching question: %v\n", err)
			}
			continue
		}
	}

	state.ExecutionComplete = true

	if e.Verbose {
		fmt.Printf("\n✅ [Execution Agent] Research complete. Collected %d summaries from %d sources\n",
			len(state.Summaries), len(state.Sources))
	}

	return nil
}

// searchForQuestion performs web search for a question
func (e *ExecutionAgent) searchForQuestion(ctx context.Context, question string) ([]SearchResult, error) {
	// Use Tavily search tool
	searchQuery := e.optimizeSearchQuery(question)
	result, err := e.Tools.SearchTool.Call(ctx, searchQuery)
	if err != nil {
		return nil, err
	}

	// Parse search results from tool output
	searchResults := e.parseSearchResults(result, question)

	return searchResults, nil
}

// optimizeSearchQuery optimizes the search query for better results
func (e *ExecutionAgent) optimizeSearchQuery(question string) string {
	// Remove question words and make it more search-friendly
	query := question
	query = strings.ReplaceAll(query, "?", "")
	query = strings.TrimSpace(query)

	// Could add more sophisticated optimization here
	return query
}

// parseSearchResults parses search results from tool output
func (e *ExecutionAgent) parseSearchResults(toolOutput string, question string) []SearchResult {
	var results []SearchResult

	// Parse the formatted output from TavilySearchTool
	lines := strings.Split(toolOutput, "\n")
	var currentResult *SearchResult

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "Result ") {
			if currentResult != nil {
				results = append(results, *currentResult)
			}
			currentResult = &SearchResult{Question: question}
		} else if after, ok := strings.CutPrefix(line, "Title: "); ok {
			if currentResult != nil {
				currentResult.Title = after
			}
		} else if after, ok := strings.CutPrefix(line, "URL: "); ok {
			if currentResult != nil {
				currentResult.URL = after
			}
		} else if after, ok := strings.CutPrefix(line, "Content: "); ok {
			if currentResult != nil {
				currentResult.Content = after
			}
		} else if after, ok := strings.CutPrefix(line, "Relevance Score: "); ok {
			if currentResult != nil {
				scoreStr := after
				if _, err := fmt.Sscanf(scoreStr, "%f", &currentResult.Score); err != nil {
					currentResult.Score = 0.0 // Default score on parse error
				}
			}
		}
	}

	// Add last result
	if currentResult != nil && currentResult.URL != "" {
		results = append(results, *currentResult)
	}

	return results
}

// summarizeSource creates a summary of a source
func (e *ExecutionAgent) summarizeSource(ctx context.Context, result SearchResult, question string) (SourceSummary, error) {
	prompt := fmt.Sprintf(`Based on the following content, provide a concise summary that addresses this research question:

Research Question: %s

Source Title: %s
Source URL: %s

Content:
%s

Please provide:
1. A brief summary of the relevant information
2. 3-5 key points that relate to the research question
3. An assessment of how relevant this source is (0-1 scale)

Format your response as:
Summary: [your summary]
Key Points:
- [point 1]
- [point 2]
- [point 3]
Relevance: [0.0-1.0]`, question, result.Title, result.URL, result.Content)

	messages := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, "You are a research analyst expert at extracting and summarizing relevant information from sources."),
		llms.TextParts(llms.ChatMessageTypeHuman, prompt),
	}

	resp, err := e.Model.GenerateContent(ctx, messages)
	if err != nil {
		return SourceSummary{}, err
	}

	if len(resp.Choices) == 0 {
		return SourceSummary{}, fmt.Errorf("no response from model")
	}

	// Parse the response
	summary := e.parseSummaryResponse(resp.Choices[0].Content, result, question)

	return summary, nil
}

// parseSummaryResponse parses the LLM's summary response
func (e *ExecutionAgent) parseSummaryResponse(response string, result SearchResult, question string) SourceSummary {
	summary := SourceSummary{
		URL:       result.URL,
		Title:     result.Title,
		Question:  question,
		Relevance: 0.7, // default
		KeyPoints: []string{},
	}

	lines := strings.Split(response, "\n")
	inKeyPoints := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if after, ok := strings.CutPrefix(line, "Summary: "); ok {
			summary.Summary = after
			summary.Summary = strings.TrimSpace(summary.Summary)
		} else if strings.Contains(line, "Key Points:") {
			inKeyPoints = true
		} else if after, ok := strings.CutPrefix(line, "Relevance: "); ok {
			relevanceStr := after
			if _, err := fmt.Sscanf(relevanceStr, "%f", &summary.Relevance); err != nil {
				summary.Relevance = 0.0 // Default relevance on parse error
			}
			inKeyPoints = false
		} else if inKeyPoints && strings.HasPrefix(line, "- ") {
			point := strings.TrimPrefix(line, "- ")
			summary.KeyPoints = append(summary.KeyPoints, point)
		} else if inKeyPoints && strings.HasPrefix(line, "* ") {
			point := strings.TrimPrefix(line, "* ")
			summary.KeyPoints = append(summary.KeyPoints, point)
		}
	}

	// If summary wasn't properly extracted, use full response
	if summary.Summary == "" {
		summary.Summary = response
	}

	return summary
}
