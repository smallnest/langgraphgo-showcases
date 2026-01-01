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
}

// NewGraph creates and configures the research agent graph.
func NewGraph() (*graph.StateRunnable[*State], error) {
	workflow := graph.NewStateGraph[*State]()

	// Add nodes with typed functions
	workflow.AddNode("planner", "Research planning node", PlannerNodeTyped)
	workflow.AddNode("researcher", "Research execution node", ResearcherNodeTyped)
	workflow.AddNode("reporter", "Report generation node", ReporterNodeTyped)
	workflow.AddNode("podcast", "Podcast script generation node", PodcastNodeTyped)

	// Add edges
	// Start -> Planner
	workflow.SetEntryPoint("planner")

	// Planner -> Researcher
	workflow.AddEdge("planner", "researcher")

	// Researcher -> Reporter
	workflow.AddEdge("researcher", "reporter")

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
