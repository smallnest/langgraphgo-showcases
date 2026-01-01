package main

import (
	"context"
	"fmt"
)

func StartNode(ctx context.Context, state ResearchState) (ResearchState, error) {
	if state.PersonName == "" {
		return ResearchState{}, fmt.Errorf("personName is required")
	}
	if state.LinkedinUrl == "" {
		return ResearchState{}, fmt.Errorf("linkedinUrl is required")
	}
	return ResearchState{
		Status: "Initializing research...",
	}, nil
}

func FetchLinkedInNode(ctx context.Context, state ResearchState) (ResearchState, error) {
	if state.LinkedinUrl == "" {
		return ResearchState{
			Status: "LinkedIn URL missing",
			Errors: []string{"Cannot fetch LinkedIn profile: linkedinUrl is missing"},
		}, nil
	}

	profile, err := fetchLinkedInProfile(state.LinkedinUrl)
	if err != nil {
		return ResearchState{
			Status: "LinkedIn profile unavailable",
			Errors: []string{fmt.Sprintf("LinkedIn fetch error: %v", err)},
		}, nil
	}

	return ResearchState{
		LinkedinData: profile,
		Status:       "Fetching LinkedIn profile...",
	}, nil
}

func GenerateSearchQueryNode(ctx context.Context, state ResearchState) (ResearchState, error) {
	if state.PersonName == "" || state.LinkedinUrl == "" {
		return ResearchState{
			Status: "Search query unavailable",
			Errors: []string{"Cannot generate search query: missing personName or linkedinUrl"},
		}, nil
	}

	query, err := generateSearchQuery(state.PersonName, state.LinkedinUrl)
	if err != nil {
		return ResearchState{
			Status: "Search query unavailable",
			Errors: []string{fmt.Sprintf("Search query error: %v", err)},
		}, nil
	}

	return ResearchState{
		SearchQuery: query,
		Status:      "Generating search query...",
	}, nil
}

func ExecuteSearchNode(ctx context.Context, state ResearchState) (ResearchState, error) {
	if state.PersonName == "" || state.LinkedinUrl == "" {
		return ResearchState{
			Status: "Web search unavailable",
			Errors: []string{"Cannot execute search: missing personName or linkedinUrl"},
		}, nil
	}

	options := PersonSearchOptions{
		MaxResults: MAX_SEARCH_RESULTS,
	}
	if state.LinkedinData != nil && state.LinkedinData.Headline != "" {
		options.Context = state.LinkedinData.Headline
	}

	results, err := searchGoogleForPerson(state.PersonName, state.LinkedinUrl, options)
	if err != nil {
		return ResearchState{
			Status: "Web search unavailable",
			Errors: []string{fmt.Sprintf("Web search error: %v", err)},
		}, nil
	}

	return ResearchState{
		SearchResults: results,
		Status:        "Searching the web...",
	}, nil
}

// ScrapeWebPageNode - Modified to handle batch scraping
func ScrapeWebPageNode(ctx context.Context, state ResearchState) (ResearchState, error) {
	var urlsToScrape []string
	for _, res := range state.SearchResults {
		urlsToScrape = append(urlsToScrape, res.Url)
	}

	if len(urlsToScrape) == 0 {
		return ResearchState{
			Status: "Scraping skipped (no URLs)",
		}, nil
	}

	scraped, err := scrapeUrls(urlsToScrape, nil)
	if err != nil {
		return ResearchState{
			Status: "Scraping failed",
			Errors: []string{fmt.Sprintf("Scrape error: %v", err)},
		}, nil
	}

	return ResearchState{
		ScrapedContents: scraped,
		Status:          fmt.Sprintf("Scraped %d pages", len(scraped)),
	}, nil
}

// SummarizeContentNode - Modified to handle batch summarization
func SummarizeContentNode(ctx context.Context, state ResearchState) (ResearchState, error) {
	var summaries []WebSummary
	
	// Process ALL scraped contents in state, assuming they are new or we want to process all.
	// Since ScrapedContents is "Appended", we might re-process if we loop.
	// But this graph is DAG (mostly), so it's fine.
	
	for _, content := range state.ScrapedContents {
		summary, err := summarizeWebContent(content.Url, content.Content, state.PersonName)
		if err != nil {
			continue 
		}
		if summary != nil {
			summaries = append(summaries, *summary)
		}
	}

	return ResearchState{
		WebSummaries: summaries,
		Status:       fmt.Sprintf("Summarized %d pages", len(summaries)),
	}, nil
}

func AggregateDataNode(ctx context.Context, state ResearchState) (ResearchState, error) {
	// Dedupe logic
	seen := make(map[string]bool)
	var deduped []WebSummary
	for _, s := range state.WebSummaries {
		if !seen[s.Url] {
			seen[s.Url] = true
			deduped = append(deduped, s)
		}
	}

	if state.LinkedinData == nil && len(deduped) == 0 {
		return ResearchState{
			Status: "Insufficient research data",
			Errors: []string{"Need LinkedIn data or at least one web summary before aggregation"},
		}, nil
	}

	// Important: We need to replace the WebSummaries with deduped list.
	// But FieldMerger logic for WebSummaries is APPEND.
	// If we return deduped list here, it will be appended to existing list!
	// This is tricky.
	// The original JS uses reducer (state, update) => update for WebSummaries in AggregateData?
	// JS: webSummaries: Annotation({ reducer: (state, update) => ... }) which appends.
	// BUT AggregateDataNode returns { webSummaries: deduped }.
	// If the reducer appends, then returning deduped list will duplicate everything again!
	
	// Wait, JS reducer: (state, update) => (Array.isArray(update) ? [...state, ...update] : [...state, update])
	// So JS ALWAYS appends.
	// AggregateData logic:
	// const deduped = dedupeSummaries(summaries);
    // return { webSummaries: deduped };
    // This looks like it would double the list in JS too if it just returns it.
    // Unless the JS code assumes that "aggregateData" is the *only* one writing to it at this stage or it overwrites?
    // Actually, maybe I misread the JS.
    // Let's check the JS code again.
    
    // JS:
    // webSummaries: Annotation<WebSummary[], WebSummary | WebSummary[]>({
    //   reducer: (state, update) => (Array.isArray(update) ? [...state, ...update] : [...state, update]),
    //   default: () => [],
    // }),
    
    // Yes, it appends.
    
    // To implement "Replace" behavior for a specific node in `langgraphgo` using `FieldMerger` is hard if the field is globally "Append".
    // I can use a different field, e.g. `FinalSummaries`?
    // Or I can just rely on `WriteReportNode` to dedupe internally before writing report, and skip `AggregateDataNode`'s modification of state.
    
    // Let's modify AggregateDataNode to NOT return WebSummaries, but just check status.
    // And WriteReportNode will read WebSummaries and dedupe them locally.
    
	return ResearchState{
		Status: "Aggregation complete",
	}, nil
}

func WriteReportNode(ctx context.Context, state ResearchState) (ResearchState, error) {
	if state.PersonName == "" || state.LinkedinUrl == "" {
		return ResearchState{
			Status: "Report unavailable",
			Errors: []string{"Cannot write report without personName and linkedinUrl"},
		}, nil
	}

    // Dedupe locally before report generation
    seen := make(map[string]bool)
	var deduped []WebSummary
	for _, s := range state.WebSummaries {
		if !seen[s.Url] {
			seen[s.Url] = true
			deduped = append(deduped, s)
		}
	}

	bundle := ResearchDataBundle{
		PersonName:    state.PersonName,
		LinkedinUrl:   state.LinkedinUrl,
		LinkedinData:  state.LinkedinData,
		WebSummaries:  deduped,
		SearchResults: state.SearchResults,
		Metadata: map[string]interface{}{
			"status": state.Status,
			"errors": state.Errors,
		},
	}

	result, err := generateResearchReport(bundle)
	if err != nil {
		return ResearchState{
			Status: "Report generation failed",
			Errors: []string{fmt.Sprintf("Report error: %v", err)},
		}, nil
	}

	return ResearchState{
		FinalReport: result.Report,
		Status:      "Report ready",
	}, nil
}