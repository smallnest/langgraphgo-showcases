package main

import (
	"context"

	"github.com/smallnest/langgraphgo/graph"
)

// Request represents the initial input to the research agent.
type Request struct {
	Query    string `json:"query"`
	MaxSteps int    `json:"max_steps,omitempty"` // Example of an additional parameter
}

// State represents the state of the research agent.
type State struct {
	Request         Request  `json:"request"`
	Plan            []string `json:"plan"`
	ResearchResults []string `json:"research_results"`
	Images          []string `json:"images"` // Image URLs from search results
	FinalReport     string   `json:"final_report"`
	PodcastScript   string   `json:"podcast_script"`
	GeneratePodcast bool     `json:"generate_podcast"`
	Step            int      `json:"step"`

	// Intent recognition
	UserIntent      string   `json:"user_intent"`      // 识别的用户意图
	RefinedQuery    string   `json:"refined_query"`    // 优化后的查询
	EntityInfo      string   `json:"entity_info"`      // 实体信息

	// Insight analysis
	InsightResults  []string `json:"insight_results"`  // 洞察分析结果
}

// NewGraph creates and configures the research agent graph.
func NewGraph() (*graph.StateRunnable[*State], error) {
	workflow := graph.NewStateGraph[*State]()

	// Add nodes with typed functions
	workflow.AddNode("query_agent", "Query analysis and intent recognition node", QueryAgentNodeTyped)
	workflow.AddNode("planner", "Research planning node", PlannerNodeTyped)
	workflow.AddNode("researcher", "Research execution node", ResearcherNodeTyped)
	workflow.AddNode("insight_agent", "Deep insight analysis node", InsightAgentNodeTyped)
	workflow.AddNode("reporter", "Report generation node", ReporterNodeTyped)
	workflow.AddNode("podcast", "Podcast script generation node", PodcastNodeTyped)

	// Add edges
	// Start -> QueryAgent
	workflow.SetEntryPoint("query_agent")

	// QueryAgent -> Planner
	workflow.AddEdge("query_agent", "planner")

	// Planner -> Researcher
	workflow.AddEdge("planner", "researcher")

	// Researcher -> InsightAgent
	workflow.AddEdge("researcher", "insight_agent")

	// InsightAgent -> Reporter
	workflow.AddEdge("insight_agent", "reporter")

	// Reporter -> Podcast (Conditional) or END
	workflow.AddConditionalEdge("reporter", func(ctx context.Context, state *State) string {
		if state.GeneratePodcast {
			return "podcast"
		}
		return graph.END
	})

	// Podcast -> End
	workflow.AddEdge("podcast", graph.END)

	return workflow.Compile()
}

// Define the node functions signatures here to avoid compilation errors in this file,
// but the actual implementation will be in nodes.go.
// Since they are in the same package (main), we don't need to declare them here if they are defined in nodes.go.
// But for clarity, I'll just rely on them being in nodes.go.
