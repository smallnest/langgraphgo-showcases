package schema

import (
	"encoding/json"
	"sync"
)

// SearchResult represents a single search result.
type SearchResult struct {
	Title         string  `json:"title"`
	URL           string  `json:"url"`
	Content       string  `json:"content"`
	Score         float64 `json:"score"`
	RawContent    string  `json:"raw_content,omitempty"`
	PublishedDate string  `json:"published_date,omitempty"`
}

// ResearchState tracks the research progress for a paragraph.
type ResearchState struct {
	SearchQueries []string                  `json:"search_queries"`
	SearchResults map[string][]SearchResult `json:"search_results"` // Query -> Results
	LatestSummary string                    `json:"latest_summary"`
	Completed     bool                      `json:"completed"`
	mu            sync.RWMutex
}

func NewResearchState() *ResearchState {
	return &ResearchState{
		SearchQueries: make([]string, 0),
		SearchResults: make(map[string][]SearchResult),
	}
}

func (rs *ResearchState) AddSearchResults(query string, results []SearchResult) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	rs.SearchQueries = append(rs.SearchQueries, query)
	rs.SearchResults[query] = results
}

func (rs *ResearchState) MarkCompleted() {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	rs.Completed = true
}

// Paragraph represents a section of the report.
type Paragraph struct {
	Title    string         `json:"title"`
	Content  string         `json:"content"` // Expected content description
	Research *ResearchState `json:"research"`
}

// BettaFishState represents the global state of the BettaFish system.
type BettaFishState struct {
	// User input
	Query string `json:"query"`

	// QueryEngine State
	ReportTitle string       `json:"report_title"`
	Paragraphs  []*Paragraph `json:"paragraphs"`
	NewsResults []string     `json:"news_results"` // The compiled news report(s)
	FinalReport string       `json:"final_report"` // The final combined report

	// MediaEngine State
	MediaResults []string `json:"media_results"`

	// InsightEngine State
	InsightResults []string `json:"insight_results"`

	// ForumEngine State
	Discussion []string `json:"discussion"`
}

func NewBettaFishState(query string) *BettaFishState {
	return &BettaFishState{
		Query:      query,
		Paragraphs: make([]*Paragraph, 0),
	}
}

// Helper to serialize state for LLM prompts
func (s *BettaFishState) ToJSON() string {
	b, _ := json.MarshalIndent(s, "", "  ")
	return string(b)
}
