package main

import (
	"context"
	"strings"

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
	Plan            []string `json:"plan"`           // Deprecated: kept for compatibility
	SectionPlans    []string `json:"section_plans"`  // Research plan for each section
	ResearchResults []string `json:"research_results"`
	Images          []string `json:"images"` // Image URLs from search results
	FinalReport     string   `json:"final_report"`
	PodcastScript   string   `json:"podcast_script"`
	GeneratePodcast bool     `json:"generate_podcast"`
	Step            int      `json:"step"`

	// Intent recognition
	UserIntent   string `json:"user_intent"`   // 识别的用户意图
	RefinedQuery string `json:"refined_query"` // 优化后的查询
	EntityInfo   string `json:"entity_info"`   // 实体信息

	// Report structure
	ReportStructure string   `json:"report_structure"` // 报告结构大纲
	ReportSections  []string `json:"report_sections"`  // 报告各章节内容
	CurrentSection   int      `json:"current_section"` // 当前正在写的章节索引

	// Reflection and revision
	ReflectionResults []string `json:"reflection_results"` // 每个章节的反思结果
	RevisionCounts    []int    `json:"revision_counts"`    // 每个章节的补充次数
	MaxRevisions      int      `json:"max_revisions"`      // 最大补充轮数（默认10）
}

// NewGraph creates and configures the research agent graph.
func NewGraph() (*graph.StateRunnable[*State], error) {
	workflow := graph.NewStateGraph[*State]()

	// Add nodes with typed functions
	workflow.AddNode("query_agent", "Query analysis and intent recognition node", QueryAgentNodeTyped)
	workflow.AddNode("structure_planner", "Report structure planning node", StructurePlannerNodeTyped)
	workflow.AddNode("planner", "Research planning node", PlannerNodeTyped)
	workflow.AddNode("researcher", "Research execution node", ResearcherNodeTyped)
	workflow.AddNode("content_writer", "Content writing node", ContentWriterNodeTyped)
	workflow.AddNode("reflector", "Content reflection node", ReflectorNodeTyped)
	workflow.AddNode("reviser", "Content revision node", ReviserNodeTyped)
	workflow.AddNode("reporter", "Final report compilation node", ReporterNodeTyped)
	workflow.AddNode("podcast", "Podcast script generation node", PodcastNodeTyped)

	// Add edges
	// Start -> QueryAgent
	workflow.SetEntryPoint("query_agent")

	// QueryAgent -> StructurePlanner (determine report structure first)
	workflow.AddEdge("query_agent", "structure_planner")

	// StructurePlanner -> Planner (create research plan based on structure)
	workflow.AddEdge("structure_planner", "planner")

	// Planner -> Researcher (execute research)
	workflow.AddEdge("planner", "researcher")

	// Researcher -> ContentWriter
	workflow.AddEdge("researcher", "content_writer")

	// ContentWriter -> Reflector (check if content meets requirements)
	workflow.AddEdge("content_writer", "reflector")

	// Reflector -> Check if revision needed or move to next section
	workflow.AddConditionalEdge("reflector", func(ctx context.Context, state *State) string {
		reflection := state.ReflectionResults[state.CurrentSection]
		if reflection == "" || strings.Contains(reflection, "无需补充") || strings.Contains(reflection, "内容完整") {
			// Content is good, move to next section
			if state.CurrentSection < len(state.ReportSections)-1 {
				state.CurrentSection++
				return "content_writer"
			}
			// All sections done, go to reporter
			return "reporter"
		}
		// Need revision
		return "reviser"
	})

	// Reviser -> Check if more revisions needed or reflection is complete
	workflow.AddConditionalEdge("reviser", func(ctx context.Context, state *State) string {
		// Increment revision count
		if state.RevisionCounts == nil {
			state.RevisionCounts = make([]int, len(state.ReportSections))
		}
		state.RevisionCounts[state.CurrentSection]++

		// Check if max revisions reached
		maxRev := state.MaxRevisions
		if maxRev == 0 {
			maxRev = 5 // Default max 5 revisions
		}

		if state.RevisionCounts[state.CurrentSection] >= maxRev {
			// Max revisions reached, move to next section
			if state.CurrentSection < len(state.ReportSections)-1 {
				state.CurrentSection++
				return "content_writer"
			}
			return "reporter"
		}

		// Reflect again to check if the revision is sufficient
		return "reflector"
	})

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
