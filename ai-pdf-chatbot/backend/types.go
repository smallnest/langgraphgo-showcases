package main

import (
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
)

// ============================================
// Ingestion Graph State
// ============================================

// IngestionState represents the state for the document ingestion graph.
// It follows the original TypeScript implementation's state structure.
type IngestionState struct {
	// Docs contains the documents to be ingested.
	// Can be a slice of Document or the string "delete" to clear the index.
	Docs interface{}
}

// ============================================
// Retrieval Graph State
// ============================================

// RetrievalState represents the state for the retrieval/chat graph.
// It follows the original TypeScript implementation's state structure.
type RetrievalState struct {
	// Messages contains the conversation history.
	Messages []llms.MessageContent

	// Documents contains the retrieved relevant documents.
	Documents []schema.Document `reducer:"append"`

	// Question is the user's input question.
	Question string

	// Route determines the execution path (retrieve or direct).
	Route string
}

// ============================================
// Router Types
// ============================================

// RouteResponse is the structured output from the routing LLM call.
type RouteResponse struct {
	// Route determines whether to retrieve documents ("retrieve") or answer directly ("direct").
	Route string `json:"route"`

	// DirectAnswer contains the pre-computed answer for direct responses (optional).
	DirectAnswer string `json:"directAnswer,omitempty"`
}

// ============================================
// API Request/Response Types
// ============================================

// IngestRequest represents a request to ingest documents.
type IngestRequest struct {
	// Files contains the file paths or URLs to process.
	Files []string `json:"files"`

	// Action can be "add" or "delete".
	Action string `json:"action"`
}

// IngestResponse represents the response from document ingestion.
type IngestResponse struct {
	// Success indicates if the ingestion was successful.
	Success bool `json:"success"`

	// Message contains a status message.
	Message string `json:"message"`

	// DocumentsIngested is the count of documents processed.
	DocumentsIngested int `json:"documentsIngested,omitempty"`

	// Error contains any error message.
	Error string `json:"error,omitempty"`
}

// ChatRequest represents a chat request from the frontend.
type ChatRequest struct {
	// Message is the user's question.
	Message string `json:"message"`

	// SessionID identifies the conversation session for checkpointing.
	SessionID string `json:"sessionId"`
}

// ChatEvent represents a server-sent event for streaming responses.
type ChatEvent struct {
	// Event type: "message", "source", "metadata", "error", or "done".
	Event string `json:"event"`

	// Data contains the event-specific data.
	Data any `json:"data"`
}

// MessageEventData contains data for message events.
type MessageEventData struct {
	// Content is the text content being streamed.
	Content string `json:"content"`

	// Role indicates who sent the message (user/assistant).
	Role string `json:"role"`
}

// SourceEventData contains data for source document events.
type SourceEventData struct {
	// Content is the document text content.
	Content string `json:"content"`

	// Metadata contains the document metadata.
	Metadata map[string]any `json:"metadata"`

	// Score is the relevance score.
	Score float64 `json:"score,omitempty"`
}

// ============================================
// Thread State (for checkpointing)
// ============================================

// ThreadState represents a conversation thread with checkpointing support.
type ThreadState struct {
	// SessionID uniquely identifies this session.
	SessionID string

	// Messages contains the conversation history.
	Messages []llms.MessageContent

	// RetrievedDocuments contains the documents used in the last query.
	RetrievedDocuments []schema.Document
}
