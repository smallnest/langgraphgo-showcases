package main

import (
	"context"

	"github.com/smallnest/langgraphgo/graph"
)

const (
	END                    = graph.END
	NODE_START             = "start"
	NODE_FETCH_LINKEDIN    = "fetchLinkedIn"
	NODE_GENERATE_QUERY    = "generateSearchQuery"
	NODE_EXECUTE_SEARCH    = "executeSearch"
	NODE_SCRAPE_WEB_PAGE   = "scrapeWebPage"
	NODE_SUMMARIZE_CONTENT = "summarizeContent"
	NODE_AGGREGATE_DATA    = "aggregateData"
	NODE_WRITE_REPORT      = "writeReport"
)

func NewResearchGraph() (*graph.StateRunnable[ResearchState], error) {
	g := graph.NewStateGraph[ResearchState]()

	// Define Schema with Custom Mergers
	fm := graph.NewFieldMerger(ResearchState{})

	// Register mergers for array fields (Append)
	fm.RegisterFieldMerge("ScrapedContents", graph.AppendSliceMerge)
	fm.RegisterFieldMerge("WebSummaries", graph.AppendSliceMerge)
	fm.RegisterFieldMerge("Errors", graph.AppendSliceMerge)

	// Other fields use default merge (Overwrite if non-zero)
	g.SetSchema(fm)

	// Add Nodes
	g.AddNode(NODE_START, "Initialize research", StartNode)
	g.AddNode(NODE_FETCH_LINKEDIN, "Fetch LinkedIn Profile", FetchLinkedInNode)
	g.AddNode(NODE_GENERATE_QUERY, "Generate Search Query", GenerateSearchQueryNode)
	g.AddNode(NODE_EXECUTE_SEARCH, "Execute Google Search", ExecuteSearchNode)
	g.AddNode(NODE_SCRAPE_WEB_PAGE, "Scrape Web Pages", ScrapeWebPageNode)
	g.AddNode(NODE_SUMMARIZE_CONTENT, "Summarize Content", SummarizeContentNode)
	g.AddNode(NODE_AGGREGATE_DATA, "Aggregate Findings", AggregateDataNode)
	g.AddNode(NODE_WRITE_REPORT, "Write Final Report", WriteReportNode)

	// Define Edges

	// Set Entry Point to Start
	g.SetEntryPoint(NODE_START)

	// start -> fetchLinkedIn AND executeSearch (Parallel)
	g.AddEdge(NODE_START, NODE_FETCH_LINKEDIN)
	g.AddEdge(NODE_START, NODE_EXECUTE_SEARCH)

	// fetchLinkedIn -> aggregateData
	g.AddEdge(NODE_FETCH_LINKEDIN, NODE_AGGREGATE_DATA)

	// executeSearch -> (conditional) scrapeWebPage
	g.AddConditionalEdge(NODE_EXECUTE_SEARCH, func(ctx context.Context, state ResearchState) string {
		if len(state.SearchResults) > 0 {
			return NODE_SCRAPE_WEB_PAGE
		}
		// If no results, go straight to aggregate (skip scraping/summary)
		return NODE_AGGREGATE_DATA
	})

	// scrapeWebPage -> (conditional) summarizeContent
	g.AddConditionalEdge(NODE_SCRAPE_WEB_PAGE, func(ctx context.Context, state ResearchState) string {
		if len(state.ScrapedContents) > 0 {
			return NODE_SUMMARIZE_CONTENT
		}
		return NODE_AGGREGATE_DATA
	})

	// summarizeContent -> aggregateData
	g.AddEdge(NODE_SUMMARIZE_CONTENT, NODE_AGGREGATE_DATA)

	// aggregateData -> writeReport
	g.AddEdge(NODE_AGGREGATE_DATA, NODE_WRITE_REPORT)

	// writeReport -> END
	g.AddEdge(NODE_WRITE_REPORT, END)

	// Compile
	return g.Compile()
}
