package main

import (
	"context"
	"fmt"

	"github.com/smallnest/langgraphgo/graph"
)

// LangManus represents the main workflow
type LangManus struct {
	Config *Config
	Tools  *ToolRegistry
	Graph  *graph.StateRunnable

	// Agents
	Coordinator *Agent
	Planner     *Agent
	Supervisor  *Agent
	Researcher  *Agent
	Coder       *Agent
	Browser     *Agent
	Reporter    *Agent
}

// NewLangManus creates a new LangManus instance
func NewLangManus(config *Config) (*LangManus, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	tools := NewToolRegistry(config)

	// Create agents
	coordinator, err := NewAgent(AgentTypeCoordinator, config, tools)
	if err != nil {
		return nil, fmt.Errorf("failed to create coordinator: %w", err)
	}

	planner, err := NewAgent(AgentTypePlanner, config, tools)
	if err != nil {
		return nil, fmt.Errorf("failed to create planner: %w", err)
	}

	supervisor, err := NewAgent(AgentTypeSupervisor, config, tools)
	if err != nil {
		return nil, fmt.Errorf("failed to create supervisor: %w", err)
	}

	researcher, err := NewAgent(AgentTypeResearcher, config, tools)
	if err != nil {
		return nil, fmt.Errorf("failed to create researcher: %w", err)
	}

	coder, err := NewAgent(AgentTypeCoder, config, tools)
	if err != nil {
		return nil, fmt.Errorf("failed to create coder: %w", err)
	}

	browser, err := NewAgent(AgentTypeBrowser, config, tools)
	if err != nil {
		return nil, fmt.Errorf("failed to create browser: %w", err)
	}

	reporter, err := NewAgent(AgentTypeReporter, config, tools)
	if err != nil {
		return nil, fmt.Errorf("failed to create reporter: %w", err)
	}

	lm := &LangManus{
		Config:      config,
		Tools:       tools,
		Coordinator: coordinator,
		Planner:     planner,
		Supervisor:  supervisor,
		Researcher:  researcher,
		Coder:       coder,
		Browser:     browser,
		Reporter:    reporter,
	}

	// Build the graph
	if err := lm.buildGraph(); err != nil {
		return nil, fmt.Errorf("failed to build graph: %w", err)
	}

	return lm, nil
}

// buildGraph constructs the LangGraph workflow
func (lm *LangManus) buildGraph() error {
	// Create workflow
	workflow := graph.NewStateGraph()

	// Define schema with reducers
	schema := graph.NewMapSchema()
	schema.RegisterReducer("messages", graph.AppendReducer)
	schema.RegisterReducer("tasks", graph.AppendReducer)
	schema.RegisterReducer("research_results", graph.AppendReducer)
	schema.RegisterReducer("code_results", graph.AppendReducer)
	schema.RegisterReducer("agent_history", graph.AppendReducer)
	workflow.SetSchema(schema)

	// Add nodes for each agent
	workflow.AddNode("coordinator", "Analyze and route initial request", lm.coordinatorNode)
	workflow.AddNode("planner", "Create execution plan", lm.plannerNode)
	workflow.AddNode("supervisor", "Orchestrate task execution", lm.supervisorNode)
	workflow.AddNode("researcher", "Conduct research", lm.researcherNode)
	workflow.AddNode("coder", "Execute code", lm.coderNode)
	workflow.AddNode("browser", "Browse web pages", lm.browserNode)
	workflow.AddNode("reporter", "Generate final report", lm.reporterNode)

	// Set entry point
	workflow.SetEntryPoint("coordinator")

	// Add conditional edges for routing
	workflow.AddConditionalEdge("coordinator", lm.routeFromCoordinator)
	workflow.AddConditionalEdge("planner", lm.routeFromPlanner)
	workflow.AddConditionalEdge("supervisor", lm.routeFromSupervisor)
	workflow.AddConditionalEdge("researcher", lm.routeFromWorker)
	workflow.AddConditionalEdge("coder", lm.routeFromWorker)
	workflow.AddConditionalEdge("browser", lm.routeFromWorker)

	workflow.AddEdge("reporter", graph.END)

	// Compile the graph
	compiled, err := workflow.Compile()
	if err != nil {
		return err
	}

	lm.Graph = compiled
	return nil
}

// Node functions - these convert between any and *State

func (lm *LangManus) coordinatorNode(ctx context.Context, stateInterface any) (any, error) {
	state := lm.interfaceToState(stateInterface)
	updatedState, err := lm.Coordinator.Execute(ctx, state)
	if err != nil {
		return nil, err
	}
	return lm.stateToInterface(updatedState), nil
}

func (lm *LangManus) plannerNode(ctx context.Context, stateInterface any) (any, error) {
	state := lm.interfaceToState(stateInterface)
	updatedState, err := lm.Planner.Execute(ctx, state)
	if err != nil {
		return nil, err
	}
	return lm.stateToInterface(updatedState), nil
}

func (lm *LangManus) supervisorNode(ctx context.Context, stateInterface any) (any, error) {
	state := lm.interfaceToState(stateInterface)
	updatedState, err := lm.Supervisor.Execute(ctx, state)
	if err != nil {
		return nil, err
	}
	return lm.stateToInterface(updatedState), nil
}

func (lm *LangManus) researcherNode(ctx context.Context, stateInterface any) (any, error) {
	state := lm.interfaceToState(stateInterface)
	updatedState, err := lm.Researcher.Execute(ctx, state)
	if err != nil {
		return nil, err
	}
	return lm.stateToInterface(updatedState), nil
}

func (lm *LangManus) coderNode(ctx context.Context, stateInterface any) (any, error) {
	state := lm.interfaceToState(stateInterface)
	updatedState, err := lm.Coder.Execute(ctx, state)
	if err != nil {
		return nil, err
	}
	return lm.stateToInterface(updatedState), nil
}

func (lm *LangManus) browserNode(ctx context.Context, stateInterface any) (any, error) {
	state := lm.interfaceToState(stateInterface)
	updatedState, err := lm.Browser.Execute(ctx, state)
	if err != nil {
		return nil, err
	}
	return lm.stateToInterface(updatedState), nil
}

func (lm *LangManus) reporterNode(ctx context.Context, stateInterface any) (any, error) {
	state := lm.interfaceToState(stateInterface)
	updatedState, err := lm.Reporter.Execute(ctx, state)
	if err != nil {
		return nil, err
	}
	return lm.stateToInterface(updatedState), nil
}

// Routing functions

func (lm *LangManus) routeFromCoordinator(ctx context.Context, stateInterface any) string {
	state := lm.interfaceToState(stateInterface)
	if state.NextAgent != nil {
		return string(state.NextAgent.Agent)
	}
	return "planner" // Default to planner
}

func (lm *LangManus) routeFromPlanner(ctx context.Context, stateInterface any) string {
	return "supervisor" // Always go to supervisor after planning
}

func (lm *LangManus) routeFromSupervisor(ctx context.Context, stateInterface any) string {
	state := lm.interfaceToState(stateInterface)

	if state.NextAgent != nil {
		nextAgent := string(state.NextAgent.Agent)

		// Check if all tasks are completed
		if nextAgent == "reporter" {
			return "reporter"
		}

		// Route to the assigned worker
		if nextAgent == "researcher" || nextAgent == "coder" || nextAgent == "browser" {
			return nextAgent
		}

		// If we need to replan
		if nextAgent == "planner" {
			return "planner"
		}
	}

	// Check if there are pending tasks
	if len(state.GetPendingTasks()) > 0 {
		// Find the next pending task and route to its assigned agent
		for _, task := range state.Tasks {
			if task.Status == "pending" {
				switch task.AssignedTo {
				case "researcher":
					return "researcher"
				case "coder":
					return "coder"
				case "browser":
					return "browser"
				}
			}
		}
	}

	// If all tasks are done, go to reporter
	return "reporter"
}

func (lm *LangManus) routeFromWorker(ctx context.Context, stateInterface any) string {
	// Workers always return to supervisor
	return "supervisor"
}

// State conversion helpers

func (lm *LangManus) interfaceToState(stateInterface any) *State {
	stateMap, ok := stateInterface.(map[string]any)
	if !ok {
		return NewState("")
	}

	state := NewState("")

	// Convert map fields to State fields
	if query, ok := stateMap["query"].(string); ok {
		state.Query = query
	}

	if messages, ok := stateMap["messages"].([]Message); ok {
		state.Messages = messages
	}

	if plan, ok := stateMap["plan"].(*Plan); ok {
		state.Plan = plan
	}

	if tasks, ok := stateMap["tasks"].([]Task); ok {
		state.Tasks = tasks
	}

	if currentTask, ok := stateMap["current_task"].(*Task); ok {
		state.CurrentTask = currentTask
	}

	if currentAgent, ok := stateMap["current_agent"].(AgentType); ok {
		state.CurrentAgent = currentAgent
	}

	if nextAgent, ok := stateMap["next_agent"].(*NextAgent); ok {
		state.NextAgent = nextAgent
	}

	if agentHistory, ok := stateMap["agent_history"].([]AgentType); ok {
		state.AgentHistory = agentHistory
	}

	if researchResults, ok := stateMap["research_results"].([]ResearchResult); ok {
		state.ResearchResults = researchResults
	}

	if codeResults, ok := stateMap["code_results"].([]CodeExecutionResult); ok {
		state.CodeResults = codeResults
	}

	if finalReport, ok := stateMap["final_report"].(string); ok {
		state.FinalReport = finalReport
	}

	if status, ok := stateMap["status"].(string); ok {
		state.Status = status
	}

	if metadata, ok := stateMap["metadata"].(map[string]any); ok {
		state.Metadata = metadata
	}

	return state
}

func (lm *LangManus) stateToInterface(state *State) any {
	return map[string]any{
		"query":            state.Query,
		"messages":         state.Messages,
		"plan":             state.Plan,
		"tasks":            state.Tasks,
		"current_task":     state.CurrentTask,
		"current_agent":    state.CurrentAgent,
		"next_agent":       state.NextAgent,
		"agent_history":    state.AgentHistory,
		"research_results": state.ResearchResults,
		"code_results":     state.CodeResults,
		"final_report":     state.FinalReport,
		"status":           state.Status,
		"metadata":         state.Metadata,
	}
}

// Run executes the workflow
func (lm *LangManus) Run(ctx context.Context, query string) (*State, error) {
	if lm.Config.Verbose {
		fmt.Printf("\n=== LangManus Starting ===\n")
		fmt.Printf("Query: %s\n\n", query)
	}

	// Create initial state
	state := NewState(query)
	state.AddHumanMessage(query)

	// Convert state to interface
	initialState := lm.stateToInterface(state)

	// Run the graph
	finalStateInterface, err := lm.Graph.Invoke(ctx, initialState)
	if err != nil {
		return nil, fmt.Errorf("workflow execution failed: %w", err)
	}

	// Convert back to State
	finalState := lm.interfaceToState(finalStateInterface)

	if lm.Config.Verbose {
		fmt.Printf("\n=== LangManus Complete ===\n")
		fmt.Println(finalState.Summary())
	}

	return finalState, nil
}
