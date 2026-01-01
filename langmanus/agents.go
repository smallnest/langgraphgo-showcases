package main

import (
	"bytes"
	"context"
	"fmt"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/google/uuid"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

// Agent represents a LangManus agent
type Agent struct {
	Type     AgentType
	Config   *Config
	Tools    *ToolRegistry
	LLM      llms.Model
	LLMSmall llms.Model
	Verbose  bool
}

// NewAgent creates a new agent
func NewAgent(agentType AgentType, config *Config, tools *ToolRegistry) (*Agent, error) {
	// Create main LLM
	llm, err := openai.New(
		openai.WithModel(config.OpenAIModel),
		openai.WithBaseURL(config.OpenAIBaseURL),
		openai.WithToken(config.OpenAIAPIKey),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create LLM: %w", err)
	}

	// Create small LLM for simpler tasks
	llmSmall, err := openai.New(
		openai.WithModel(config.OpenAIModelSmall),
		openai.WithBaseURL(config.OpenAIBaseURL),
		openai.WithToken(config.OpenAIAPIKey),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create small LLM: %w", err)
	}

	return &Agent{
		Type:     agentType,
		Config:   config,
		Tools:    tools,
		LLM:      llm,
		LLMSmall: llmSmall,
		Verbose:  config.Verbose,
	}, nil
}

// Execute runs the agent on the given state
func (a *Agent) Execute(ctx context.Context, state *State) (*State, error) {
	if a.Verbose {
		fmt.Printf("\n=== %s Agent Executing ===\n", strings.ToUpper(string(a.Type)))
	}

	state.UpdateCurrentAgent(a.Type)

	switch a.Type {
	case AgentTypeCoordinator:
		return a.executeCoordinator(ctx, state)
	case AgentTypePlanner:
		return a.executePlanner(ctx, state)
	case AgentTypeSupervisor:
		return a.executeSupervisor(ctx, state)
	case AgentTypeResearcher:
		return a.executeResearcher(ctx, state)
	case AgentTypeCoder:
		return a.executeCoder(ctx, state)
	case AgentTypeBrowser:
		return a.executeBrowser(ctx, state)
	case AgentTypeReporter:
		return a.executeReporter(ctx, state)
	default:
		return nil, fmt.Errorf("unknown agent type: %s", a.Type)
	}
}

func (a *Agent) executeCoordinator(ctx context.Context, state *State) (*State, error) {
	prompt := a.renderPrompt(CoordinatorPrompt, state)
	response, err := a.callLLM(ctx, prompt, false)
	if err != nil {
		return nil, err
	}

	state.AddAIMessage(response, string(AgentTypeCoordinator))

	// Parse the response to determine next agent
	nextAgent := a.parseNextAgent(response)
	if nextAgent != nil {
		state.NextAgent = nextAgent
	}

	return state, nil
}

func (a *Agent) executePlanner(ctx context.Context, state *State) (*State, error) {
	prompt := a.renderPrompt(PlannerPrompt, state)
	response, err := a.callLLM(ctx, prompt, false)
	if err != nil {
		return nil, err
	}

	state.AddAIMessage(response, string(AgentTypePlanner))

	// Parse the plan from the response
	plan := a.parsePlan(response)
	if plan != nil {
		state.Plan = plan
		// Create tasks from plan steps
		for i, step := range plan.Steps {
			assignedTo := a.extractAssignedAgent(step)
			task := Task{
				ID:          uuid.New().String(),
				Description: step,
				Status:      "pending",
				AssignedTo:  assignedTo,
				CreatedAt:   time.Now(),
			}
			state.AddTask(task)
			if a.Verbose {
				fmt.Printf("Task %d: %s -> %s\n", i+1, step, assignedTo)
			}
		}
	}

	// Parse next agent
	nextAgent := a.parseNextAgent(response)
	if nextAgent != nil {
		state.NextAgent = nextAgent
	}

	return state, nil
}

func (a *Agent) executeSupervisor(ctx context.Context, state *State) (*State, error) {
	prompt := a.renderPrompt(SupervisorPrompt, state)
	response, err := a.callLLM(ctx, prompt, false)
	if err != nil {
		return nil, err
	}

	state.AddAIMessage(response, string(AgentTypeSupervisor))

	// Parse task assignment or routing decision
	if task := a.parseTaskAssignment(response); task != "" {
		// Update current task
		for i := range state.Tasks {
			if state.Tasks[i].Status == "pending" && strings.Contains(strings.ToLower(state.Tasks[i].Description), strings.ToLower(task)) {
				state.Tasks[i].Status = "in_progress"
				state.CurrentTask = &state.Tasks[i]
				break
			}
		}
	}

	nextAgent := a.parseNextAgent(response)
	if nextAgent != nil {
		state.NextAgent = nextAgent
	}

	return state, nil
}

func (a *Agent) executeResearcher(ctx context.Context, state *State) (*State, error) {
	prompt := a.renderPrompt(ResearcherPrompt, state)
	response, err := a.callLLM(ctx, prompt, true) // Use main LLM for research
	if err != nil {
		return nil, err
	}

	// Extract search query from response
	searchQuery := a.extractSearchQuery(response, state)

	if a.Verbose {
		fmt.Printf("Search query: %s\n", searchQuery)
	}

	// Perform search
	if searchQuery != "" {
		if a.Tools.Search.APIKey == "" {
			if a.Verbose {
				fmt.Println("⚠️  WARNING: SEARCH_API_KEY not set, skipping web search")
				fmt.Println("    Set SEARCH_API_KEY environment variable to enable search")
			}
			// Create a placeholder result
			state.ResearchResults = append(state.ResearchResults, ResearchResult{
				Query:   searchQuery,
				Sources: []Source{},
				Summary: "Search skipped: API key not configured",
			})
		} else {
			sources, err := a.Tools.Search.Search(ctx, searchQuery, 5)
			if err != nil {
				if a.Verbose {
					fmt.Printf("❌ Search error: %v\n", err)
				}
			} else {
				// Add research result
				result := ResearchResult{
					Query:   searchQuery,
					Sources: sources,
					Summary: a.summarizeSources(sources),
				}
				state.ResearchResults = append(state.ResearchResults, result)

				if a.Verbose {
					fmt.Printf("✓ Research completed: %d sources found\n", len(sources))
					for i, source := range sources {
						fmt.Printf("  %d. %s (%s)\n", i+1, source.Title, source.URL)
					}
				}
			}
		}
	}

	state.AddAIMessage(response, string(AgentTypeResearcher))

	// Mark current task as completed
	if state.CurrentTask != nil {
		for i := range state.Tasks {
			if state.Tasks[i].ID == state.CurrentTask.ID {
				state.Tasks[i].Status = "completed"
				state.Tasks[i].CompletedAt = time.Now()
				state.Tasks[i].Result = response
				break
			}
		}
		state.CurrentTask = nil
	}

	nextAgent := a.parseNextAgent(response)
	if nextAgent != nil {
		state.NextAgent = nextAgent
	}

	return state, nil
}

func (a *Agent) executeCoder(ctx context.Context, state *State) (*State, error) {
	prompt := a.renderPrompt(CoderPrompt, state)
	response, err := a.callLLM(ctx, prompt, true) // Use main LLM for coding
	if err != nil {
		return nil, err
	}

	// Extract code from response
	code, language := a.extractCode(response)

	// Execute code if present and enabled
	if code != "" && a.Tools.Config.EnableCodeExecution {
		var result *CodeExecutionResult
		if language == "python" {
			result, err = a.Tools.Executor.ExecutePython(ctx, code)
		} else if language == "bash" {
			result, err = a.Tools.Executor.ExecuteBash(ctx, code)
		}

		if err != nil {
			if a.Verbose {
				fmt.Printf("Code execution error: %v\n", err)
			}
		}

		if result != nil {
			state.CodeResults = append(state.CodeResults, *result)
		}
	}

	state.AddAIMessage(response, string(AgentTypeCoder))

	// Mark current task as completed
	if state.CurrentTask != nil {
		for i := range state.Tasks {
			if state.Tasks[i].ID == state.CurrentTask.ID {
				state.Tasks[i].Status = "completed"
				state.Tasks[i].CompletedAt = time.Now()
				state.Tasks[i].Result = response
				break
			}
		}
		state.CurrentTask = nil
	}

	nextAgent := a.parseNextAgent(response)
	if nextAgent != nil {
		state.NextAgent = nextAgent
	}

	return state, nil
}

func (a *Agent) executeBrowser(ctx context.Context, state *State) (*State, error) {
	prompt := a.renderPrompt(BrowserPrompt, state)
	response, err := a.callLLM(ctx, prompt, false)
	if err != nil {
		return nil, err
	}

	state.AddAIMessage(response, string(AgentTypeBrowser))

	// Mark current task as completed
	if state.CurrentTask != nil {
		for i := range state.Tasks {
			if state.Tasks[i].ID == state.CurrentTask.ID {
				state.Tasks[i].Status = "completed"
				state.Tasks[i].CompletedAt = time.Now()
				state.Tasks[i].Result = response
				break
			}
		}
		state.CurrentTask = nil
	}

	nextAgent := a.parseNextAgent(response)
	if nextAgent != nil {
		state.NextAgent = nextAgent
	}

	return state, nil
}

func (a *Agent) executeReporter(ctx context.Context, state *State) (*State, error) {
	prompt := a.renderPrompt(ReporterPrompt, state)
	response, err := a.callLLM(ctx, prompt, true) // Use main LLM for final report
	if err != nil {
		return nil, err
	}

	state.AddAIMessage(response, string(AgentTypeReporter))

	// Extract final report
	if report := a.extractFinalReport(response); report != "" {
		state.FinalReport = report
	} else {
		state.FinalReport = response
	}

	state.Status = "completed"

	return state, nil
}

// Helper functions

func (a *Agent) callLLM(ctx context.Context, prompt string, useMainLLM bool) (string, error) {
	model := a.LLMSmall
	modelName := "small"
	if useMainLLM {
		model = a.LLM
		modelName = "main"
	}

	if a.Verbose {
		fmt.Printf("Calling LLM (%s)...\n", modelName)
	}

	response, err := model.GenerateContent(ctx, []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, "You are a helpful AI assistant."),
		llms.TextParts(llms.ChatMessageTypeHuman, prompt),
	})

	if err != nil {
		return "", fmt.Errorf("LLM call failed: %w", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no response from LLM")
	}

	content := response.Choices[0].Content

	if a.Verbose {
		fmt.Printf("LLM Response (first 500 chars):\n%s\n", truncate(content, 500))
		fmt.Println(strings.Repeat("-", 80))
	}

	return content, nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func (a *Agent) renderPrompt(promptTemplate string, state *State) string {
	// Create template with custom functions
	funcMap := template.FuncMap{
		"add": func(a, b int) int { return a + b },
	}

	tmpl, err := template.New("prompt").Funcs(funcMap).Parse(promptTemplate)
	if err != nil {
		return promptTemplate
	}

	var buf bytes.Buffer
	data := map[string]any{
		"Query":           state.Query,
		"Messages":        state.FormatMessages(),
		"Plan":            state.Plan,
		"Tasks":           state.Tasks,
		"CurrentTask":     state.CurrentTask,
		"ResearchResults": state.ResearchResults,
		"CodeResults":     state.CodeResults,
	}

	if err := tmpl.Execute(&buf, data); err != nil {
		return promptTemplate
	}

	return buf.String()
}

func (a *Agent) parseNextAgent(response string) *NextAgent {
	// Look for NEXT_AGENT: pattern
	re := regexp.MustCompile(`(?i)NEXT_AGENT:\s*(\w+)`)
	matches := re.FindStringSubmatch(response)
	if len(matches) < 2 {
		return nil
	}

	agentStr := strings.ToLower(strings.TrimSpace(matches[1]))
	var agentType AgentType

	switch agentStr {
	case "coordinator":
		agentType = AgentTypeCoordinator
	case "planner":
		agentType = AgentTypePlanner
	case "supervisor":
		agentType = AgentTypeSupervisor
	case "researcher":
		agentType = AgentTypeResearcher
	case "coder":
		agentType = AgentTypeCoder
	case "browser":
		agentType = AgentTypeBrowser
	case "reporter":
		agentType = AgentTypeReporter
	default:
		return nil
	}

	// Extract reason
	reasonRe := regexp.MustCompile(`(?i)REASON:\s*(.+?)(?:\n|$)`)
	reasonMatches := reasonRe.FindStringSubmatch(response)
	reason := ""
	if len(reasonMatches) >= 2 {
		reason = strings.TrimSpace(reasonMatches[1])
	}

	return &NextAgent{
		Agent:  agentType,
		Reason: reason,
	}
}

func (a *Agent) parsePlan(response string) *Plan {
	// Extract plan description
	descRe := regexp.MustCompile(`(?i)PLAN_DESCRIPTION:\s*(.+?)(?:\n|STEPS:)`)
	matches := descRe.FindStringSubmatch(response)
	if len(matches) < 2 {
		return nil
	}

	description := strings.TrimSpace(matches[1])

	// Extract steps
	stepsRe := regexp.MustCompile(`(?i)STEPS:\s*\n((?:\d+\..*\n?)+)`)
	stepsMatches := stepsRe.FindStringSubmatch(response)
	if len(stepsMatches) < 2 {
		return nil
	}

	stepsText := stepsMatches[1]
	stepLines := strings.Split(stepsText, "\n")
	var steps []string

	for _, line := range stepLines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Remove numbering
		stepRe := regexp.MustCompile(`^\d+\.\s*(.+)`)
		stepMatches := stepRe.FindStringSubmatch(line)
		if len(stepMatches) >= 2 {
			steps = append(steps, strings.TrimSpace(stepMatches[1]))
		}
	}

	if len(steps) == 0 {
		return nil
	}

	return &Plan{
		Steps:       steps,
		Description: description,
		Strategy:    "multi-agent",
	}
}

func (a *Agent) extractAssignedAgent(step string) string {
	re := regexp.MustCompile(`(?i)ASSIGN TO:\s*(\w+)`)
	matches := re.FindStringSubmatch(step)
	if len(matches) >= 2 {
		return strings.ToLower(strings.TrimSpace(matches[1]))
	}
	return "unknown"
}

func (a *Agent) parseTaskAssignment(response string) string {
	re := regexp.MustCompile(`(?i)TASK:\s*(.+?)(?:\n|REASON:)`)
	matches := re.FindStringSubmatch(response)
	if len(matches) >= 2 {
		return strings.TrimSpace(matches[1])
	}
	return ""
}

func (a *Agent) extractSearchQuery(response string, state *State) string {
	// Try to extract explicit search query
	re := regexp.MustCompile(`(?i)SEARCH(?:_QUERY)?:\s*(.+?)(?:\n|$)`)
	matches := re.FindStringSubmatch(response)
	if len(matches) >= 2 {
		return strings.TrimSpace(matches[1])
	}

	// Otherwise use the current query or task
	if state.CurrentTask != nil {
		return state.CurrentTask.Description
	}
	return state.Query
}

func (a *Agent) extractCode(response string) (code, language string) {
	// Look for code blocks
	pythonRe := regexp.MustCompile("(?s)```python\\n(.+?)```")
	bashRe := regexp.MustCompile("(?s)```bash\\n(.+?)```")

	if matches := pythonRe.FindStringSubmatch(response); len(matches) >= 2 {
		return strings.TrimSpace(matches[1]), "python"
	}

	if matches := bashRe.FindStringSubmatch(response); len(matches) >= 2 {
		return strings.TrimSpace(matches[1]), "bash"
	}

	return "", ""
}

func (a *Agent) extractFinalReport(response string) string {
	re := regexp.MustCompile(`(?is)FINAL_REPORT:\s*(.+?)(?:STATUS:|$)`)
	matches := re.FindStringSubmatch(response)
	if len(matches) >= 2 {
		return strings.TrimSpace(matches[1])
	}
	return ""
}

func (a *Agent) summarizeSources(sources []Source) string {
	if len(sources) == 0 {
		return "No sources found"
	}

	var summary strings.Builder
	for i, source := range sources {
		if i >= 3 { // Limit to top 3
			break
		}
		summary.WriteString(fmt.Sprintf("%d. %s\n", i+1, source.Title))
		if len(source.Content) > 200 {
			summary.WriteString(source.Content[:200] + "...\n")
		} else {
			summary.WriteString(source.Content + "\n")
		}
	}
	return summary.String()
}
