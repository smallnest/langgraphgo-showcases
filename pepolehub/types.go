package main

// LinkedInData represents the profile data from LinkedIn
type LinkedInData struct {
	LinkedinUrl         string `json:"linkedinUrl"`
	LinkedinId          string `json:"linkedinId"`
	LinkedinNumId       string `json:"linkedinNumId,omitempty"`
	FirstName           string `json:"firstName"`
	LastName            string `json:"lastName"`
	FullName            string `json:"fullName"`
	Headline            string `json:"headline,omitempty"`
	About               string `json:"about,omitempty"`
	Location            string `json:"location,omitempty"`
	City                string `json:"city,omitempty"`
	CountryCode         string `json:"countryCode,omitempty"`
	ProfilePicUrl       string `json:"profilePicUrl,omitempty"`
	BannerImage         string `json:"bannerImage,omitempty"`
	DefaultAvatar       bool   `json:"defaultAvatar,omitempty"`
	CurrentCompany      string `json:"currentCompany,omitempty"`
	CurrentCompanyId    string `json:"currentCompanyId,omitempty"`
	Experience          []any  `json:"experience,omitempty"`
	Education           []any  `json:"education,omitempty"`
	Languages           []any  `json:"languages,omitempty"`
	Connections         int    `json:"connections,omitempty"`
	Followers           int    `json:"followers,omitempty"`
	MemorializedAccount bool   `json:"memorializedAccount,omitempty"`
}

// SearchResult represents a single search result
type SearchResult struct {
	Title       string `json:"title"`
	Url         string `json:"url"`
	Snippet     string `json:"snippet,omitempty"`
	Rank        int    `json:"rank"`
	Source      string `json:"source"`
	CountryCode string `json:"countryCode,omitempty"`
}

// ScrapedContent represents the content scraped from a URL
type ScrapedContent struct {
	Url         string                 `json:"url"`
	Status      int                    `json:"status,omitempty"`
	ContentType string                 `json:"contentType,omitempty"`
	Content     string                 `json:"content"`
	Bytes       int                    `json:"bytes"`
	FetchedAt   int64                  `json:"fetchedAt"`
	Error       string                 `json:"error,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// WebSummary represents a summary of a scraped web page
type WebSummary struct {
	Url            string   `json:"url"`
	Summary        string   `json:"summary"`
	KeyPoints      []string `json:"keyPoints"`
	MentionsPerson bool     `json:"mentionsPerson"`
	Confidence     float64  `json:"confidence,omitempty"`
	Sentiment      string   `json:"sentiment,omitempty"` // "positive", "neutral", "negative"
	Source         string   `json:"source,omitempty"`
	RawExcerpt     string   `json:"rawExcerpt,omitempty"`
}

// ResearchState represents the state of the research graph
type ResearchState struct {
	PersonName      string           `json:"personName"`
	LinkedinUrl     string           `json:"linkedinUrl"`
	LinkedinData    *LinkedInData    `json:"linkedinData,omitempty"`
	SearchQuery     string           `json:"searchQuery,omitempty"`
	SearchResults   []SearchResult   `json:"searchResults,omitempty"`
	ScrapedContents []ScrapedContent `json:"scrapedContents,omitempty"`
	WebSummaries    []WebSummary     `json:"webSummaries,omitempty"`
	FinalReport     string           `json:"finalReport,omitempty"`
	Errors          []string         `json:"errors,omitempty"`
	Status          string           `json:"status"`
}

// ResearchInput represents the initial input to the research process
type ResearchInput struct {
	PersonName   string `json:"personName"`
	LinkedinUrl  string `json:"linkedinUrl"`
	Context      string `json:"context,omitempty"`
	ForceRefresh bool   `json:"forceRefresh,omitempty"`
}
