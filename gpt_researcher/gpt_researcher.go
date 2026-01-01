package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/smallnest/langgraphgo/graph"
	"github.com/tmc/langchaingo/llms/openai"
)

// GPTResearcher is the main research orchestrator
type GPTResearcher struct {
	Config         *Config
	PlannerAgent   *PlannerAgent
	ExecutionAgent *ExecutionAgent
	PublisherAgent *PublisherAgent
	Tools          *ToolRegistry
	Graph          *graph.StateRunnable[*ResearchState]
}

// NewGPTResearcher creates a new GPT Researcher instance
func NewGPTResearcher(config *Config) (*GPTResearcher, error) {
	// Create LLM models
	plannerModel, err := openai.New(
		openai.WithModel(config.Model),
		openai.WithToken(config.OpenAIAPIKey),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create planner model: %w", err)
	}

	executionModel, err := openai.New(
		openai.WithModel(config.Model),
		openai.WithToken(config.OpenAIAPIKey),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create execution model: %w", err)
	}

	publisherModel, err := openai.New(
		openai.WithModel(config.ReportModel),
		openai.WithToken(config.OpenAIAPIKey),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create publisher model: %w", err)
	}

	summaryModel, err := openai.New(
		openai.WithModel(config.SummaryModel),
		openai.WithToken(config.OpenAIAPIKey),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create summary model: %w", err)
	}

	// Create tools

	tools := NewToolRegistry(config, summaryModel)

	// Create agents
	plannerAgent := NewPlannerAgent(plannerModel, config)
	executionAgent := NewExecutionAgent(executionModel, config, tools)

	publisherAgent := NewPublisherAgent(publisherModel, config)

	researcher := &GPTResearcher{
		Config:         config,
		PlannerAgent:   plannerAgent,
		ExecutionAgent: executionAgent,
		PublisherAgent: publisherAgent,
		Tools:          tools,
	}

	// Build the workflow graph
	if err := researcher.buildGraph(); err != nil {
		return nil, fmt.Errorf("failed to build graph: %w", err)
	}

	return researcher, nil
}

// buildGraph constructs the research workflow using langgraphgo
func (r *GPTResearcher) buildGraph() error {
	// Create workflow with typed state
	workflow := graph.NewStateGraph[*ResearchState]()

	// Define nodes
	workflow.AddNode("planner", "Generate research questions", func(ctx context.Context, state *ResearchState) (*ResearchState, error) {
		if err := r.PlannerAgent.GenerateQuestions(ctx, state); err != nil {
			return nil, err
		}
		return state, nil
	})
	workflow.AddNode("executor", "Execute research and gather information", func(ctx context.Context, state *ResearchState) (*ResearchState, error) {
		if err := r.ExecutionAgent.ExecuteAll(ctx, state); err != nil {
			return nil, err
		}
		return state, nil
	})
	workflow.AddNode("publisher", "Generate final research report", func(ctx context.Context, state *ResearchState) (*ResearchState, error) {
		if err := r.PublisherAgent.GenerateReport(ctx, state); err != nil {
			return nil, err
		}
		return state, nil
	})

	// Add edges
	workflow.SetEntryPoint("planner")
	workflow.AddEdge("planner", "executor")
	workflow.AddEdge("executor", "publisher")
	workflow.AddEdge("publisher", graph.END)

	// Compile graph
	compiled, err := workflow.Compile()
	if err != nil {
		return err
	}

	r.Graph = compiled
	return nil
}

// ConductResearch executes the full research workflow
func (r *GPTResearcher) ConductResearch(ctx context.Context, query string) (*ResearchState, error) {
	if r.Config.Verbose {
		fmt.Println("\n" + strings.Repeat("=", 80))
		fmt.Println("GPT RESEARCHER")
		fmt.Println(strings.Repeat("=", 80))
		fmt.Printf("\n📋 Research Query: %s\n", query)
		fmt.Println()
	}

	// Create initial state
	initialState := NewResearchState(query)

	// Execute graph
	result, err := r.Graph.Invoke(ctx, initialState)
	if err != nil {
		return nil, fmt.Errorf("research workflow failed: %w", err)
	}

	if r.Config.Verbose {
		fmt.Println("\n" + strings.Repeat("=", 80))
		fmt.Println("RESEARCH COMPLETE")
		fmt.Println(strings.Repeat("=", 80))
		fmt.Printf("\nStatistics:\n")
		fmt.Printf("- Research Questions: %d\n", len(result.Questions))
		fmt.Printf("- Sources Consulted: %d\n", len(result.Sources))
		fmt.Printf("- Summaries Generated: %d\n", len(result.Summaries))
		fmt.Printf("- Report Length: %d characters\n", len(result.FinalReport))
		fmt.Printf("- Duration: %.1f minutes\n\n", result.EndTime.Sub(result.StartTime).Minutes())
	}

	return result, nil
}

// WriteReport generates and returns the final report
func (r *GPTResearcher) WriteReport(ctx context.Context, state *ResearchState) (string, error) {
	if state.FinalReport != "" {
		return state.FinalReport, nil
	}

	// If report wasn't generated yet, generate it now
	if err := r.PublisherAgent.GenerateReport(ctx, state); err != nil {
		return "", err
	}

	return state.FinalReport, nil
}
