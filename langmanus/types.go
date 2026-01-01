package main

import (
	"time"
)

// MessageType represents the type of message in the conversation
type MessageType string

const (
	MessageTypeHuman  MessageType = "human"
	MessageTypeAI     MessageType = "ai"
	MessageTypeSystem MessageType = "system"
	MessageTypeTool   MessageType = "tool"
)

// Message represents a single message in the conversation
type Message struct {
	Type       MessageType    `json:"type"`
	Content    string         `json:"content"`
	Name       string         `json:"name,omitempty"`
	ToolCallID string         `json:"tool_call_id,omitempty"`
	ToolCalls  []ToolCall     `json:"tool_calls,omitempty"`
	Metadata   map[string]any `json:"metadata,omitempty"`
}

// ToolCall represents a tool invocation
type ToolCall struct {
	ID   string         `json:"id"`
	Name string         `json:"name"`
	Args map[string]any `json:"args"`
}

// Task represents a research or coding task
type Task struct {
	ID          string    `json:"id"`
	Description string    `json:"description"`
	Status      string    `json:"status"` // "pending", "in_progress", "completed", "failed"
	AssignedTo  string    `json:"assigned_to"`
	Result      string    `json:"result,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	CompletedAt time.Time `json:"completed_at,omitempty"`
}

// ResearchResult represents the result of a research task
type ResearchResult struct {
	Query   string   `json:"query"`
	Sources []Source `json:"sources"`
	Summary string   `json:"summary"`
}

// Source represents a search result source
type Source struct {
	Title   string  `json:"title"`
	URL     string  `json:"url"`
	Content string  `json:"content"`
	Score   float64 `json:"score,omitempty"`
}

// CodeExecutionResult represents the result of code execution
type CodeExecutionResult struct {
	Code     string `json:"code"`
	Output   string `json:"output"`
	Error    string `json:"error,omitempty"`
	ExitCode int    `json:"exit_code"`
}

// Plan represents a task execution plan
type Plan struct {
	Steps       []string `json:"steps"`
	Description string   `json:"description"`
	Strategy    string   `json:"strategy"`
}

// AgentType represents the type of agent
type AgentType string

const (
	AgentTypeCoordinator AgentType = "coordinator"
	AgentTypePlanner     AgentType = "planner"
	AgentTypeSupervisor  AgentType = "supervisor"
	AgentTypeResearcher  AgentType = "researcher"
	AgentTypeCoder       AgentType = "coder"
	AgentTypeBrowser     AgentType = "browser"
	AgentTypeReporter    AgentType = "reporter"
)

// NextAgent represents the routing decision
type NextAgent struct {
	Agent  AgentType `json:"agent"`
	Reason string    `json:"reason"`
}
