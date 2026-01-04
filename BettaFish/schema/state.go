package schema

import (
	"encoding/json"
	"sync"
	"time"
)

// SearchResult represents a single search result.
type SearchResult struct {
	Title         string  `json:"title"`
	URL           string  `json:"url"`
	Content       string  `json:"content"`
	Score         float64 `json:"score"`
	RawContent    string  `json:"raw_content,omitempty"`
	PublishedDate string  `json:"published_date,omitempty"`
	// Sentiment analysis (for GraphRAG)
	Sentiment      string  `json:"sentiment,omitempty"`       // positive, negative, neutral
	SentimentScore float64 `json:"sentiment_score,omitempty"` // Confidence score
	Platform       string  `json:"platform,omitempty"`        // Source platform (weibo, zhihu, etc.)
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

// GraphRAGNode represents a node in the knowledge graph.
type GraphRAGNode struct {
	ID          string            `json:"id"`
	Type        string            `json:"type"` // entity, concept, event
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Attributes  map[string]string `json:"attributes"`
	Source      string            `json:"source"` // Which engine/node created this
	CreatedAt   time.Time         `json:"created_at"`
}

// GraphRAGEdge represents a relationship between nodes.
type GraphRAGEdge struct {
	ID        string    `json:"id"`
	Source    string    `json:"source"`   // Source node ID
	Target    string    `json:"target"`   // Target node ID
	Relation  string    `json:"relation"` // Type of relationship
	Weight    float64   `json:"weight"`   // Strength of relationship
	CreatedAt time.Time `json:"created_at"`
}

// GraphRAGConfig holds GraphRAG configuration.
type GraphRAGConfig struct {
	Enabled       bool     `json:"enabled"`
	MaxQueries    int      `json:"max_queries"`    // Max queries for knowledge retrieval
	NodeTypes     []string `json:"node_types"`     // Types of nodes to extract
	RelationTypes []string `json:"relation_types"` // Types of relations to extract
}

// DefaultGraphRAGConfig returns default GraphRAG configuration.
func DefaultGraphRAGConfig() GraphRAGConfig {
	return GraphRAGConfig{
		Enabled:       false,
		MaxQueries:    3,
		NodeTypes:     []string{"entity", "concept", "event"},
		RelationTypes: []string{"related_to", "causes", "part_of", "mentions"},
	}
}

// GraphRAGState holds the knowledge graph and related state.
type GraphRAGState struct {
	Nodes     []GraphRAGNode `json:"nodes"`
	Edges     []GraphRAGEdge `json:"edges"`
	Queries   []string       `json:"queries"`   // Queries made for knowledge retrieval
	Responses []string       `json:"responses"` // Responses from graph queries
	Config    GraphRAGConfig `json:"config"`
	mu        sync.RWMutex
}

// NewGraphRAGState creates a new GraphRAG state.
func NewGraphRAGState() *GraphRAGState {
	return &GraphRAGState{
		Nodes:     make([]GraphRAGNode, 0),
		Edges:     make([]GraphRAGEdge, 0),
		Queries:   make([]string, 0),
		Responses: make([]string, 0),
		Config:    DefaultGraphRAGConfig(),
	}
}

// AddNode adds a node to the graph.
func (g *GraphRAGState) AddNode(node GraphRAGNode) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.Nodes = append(g.Nodes, node)
}

// AddEdge adds an edge to the graph.
func (g *GraphRAGState) AddEdge(edge GraphRAGEdge) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.Edges = append(g.Edges, edge)
}

// AddQuery records a graph query and its response.
func (g *GraphRAGState) AddQuery(query, response string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.Queries = append(g.Queries, query)
	g.Responses = append(g.Responses, response)
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

	// GraphRAG State (new)
	GraphRAG *GraphRAGState `json:"graphrag,omitempty"`

	// Execution metadata
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time,omitempty"`

	// Configuration
	Config map[string]any `json:"config,omitempty"`
}

func NewBettaFishState(query string) *BettaFishState {
	return &BettaFishState{
		Query:      query,
		Paragraphs: make([]*Paragraph, 0),
		GraphRAG:   NewGraphRAGState(),
		StartTime:  time.Now(),
		Config:     make(map[string]any),
	}
}

// Helper to serialize state for LLM prompts
func (s *BettaFishState) ToJSON() string {
	b, _ := json.MarshalIndent(s, "", "  ")
	return string(b)
}
