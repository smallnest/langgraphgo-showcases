package main

import (
	"fmt"
)

// State represents the overall state of the LangManus workflow
type State struct {
	// Core fields
	Query    string    `json:"query"`
	Messages []Message `json:"messages"`

	// Planning and execution
	Plan        *Plan  `json:"plan,omitempty"`
	Tasks       []Task `json:"tasks"`
	CurrentTask *Task  `json:"current_task,omitempty"`

	// Agent routing
	CurrentAgent AgentType   `json:"current_agent"`
	NextAgent    *NextAgent  `json:"next_agent,omitempty"`
	AgentHistory []AgentType `json:"agent_history"`

	// Research and coding results
	ResearchResults []ResearchResult      `json:"research_results"`
	CodeResults     []CodeExecutionResult `json:"code_results"`

	// Final output
	FinalReport string `json:"final_report,omitempty"`
	Status      string `json:"status"` // "in_progress", "completed", "failed"

	// Metadata
	Metadata map[string]any `json:"metadata"`
}

// NewState creates a new state with the given query
func NewState(query string) *State {
	return &State{
		Query:           query,
		Messages:        []Message{},
		Tasks:           []Task{},
		AgentHistory:    []AgentType{},
		ResearchResults: []ResearchResult{},
		CodeResults:     []CodeExecutionResult{},
		Status:          "in_progress",
		Metadata:        make(map[string]any),
	}
}

// AddMessage adds a message to the state
func (s *State) AddMessage(msgType MessageType, content string, name ...string) {
	msg := Message{
		Type:    msgType,
		Content: content,
	}
	if len(name) > 0 {
		msg.Name = name[0]
	}
	s.Messages = append(s.Messages, msg)
}

// AddSystemMessage adds a system message to the state
func (s *State) AddSystemMessage(content string) {
	s.AddMessage(MessageTypeSystem, content)
}

// AddHumanMessage adds a human message to the state
func (s *State) AddHumanMessage(content string) {
	s.AddMessage(MessageTypeHuman, content)
}

// AddAIMessage adds an AI message to the state
func (s *State) AddAIMessage(content string, name string) {
	s.AddMessage(MessageTypeAI, content, name)
}

// AddTask adds a new task to the state
func (s *State) AddTask(task Task) {
	s.Tasks = append(s.Tasks, task)
}

// UpdateCurrentAgent updates the current agent and adds to history
func (s *State) UpdateCurrentAgent(agent AgentType) {
	s.CurrentAgent = agent
	s.AgentHistory = append(s.AgentHistory, agent)
}

// GetLastNMessages returns the last n messages
func (s *State) GetLastNMessages(n int) []Message {
	if len(s.Messages) <= n {
		return s.Messages
	}
	return s.Messages[len(s.Messages)-n:]
}

// FormatMessages formats messages for LLM consumption
func (s *State) FormatMessages() string {
	var result string
	for _, msg := range s.Messages {
		prefix := ""
		switch msg.Type {
		case MessageTypeHuman:
			prefix = "Human"
		case MessageTypeAI:
			prefix = "AI"
			if msg.Name != "" {
				prefix = fmt.Sprintf("AI (%s)", msg.Name)
			}
		case MessageTypeSystem:
			prefix = "System"
		case MessageTypeTool:
			prefix = fmt.Sprintf("Tool (%s)", msg.Name)
		}
		result += fmt.Sprintf("%s: %s\n\n", prefix, msg.Content)
	}
	return result
}

// GetPendingTasks returns all pending tasks
func (s *State) GetPendingTasks() []Task {
	var pending []Task
	for _, task := range s.Tasks {
		if task.Status == "pending" {
			pending = append(pending, task)
		}
	}
	return pending
}

// GetCompletedTasks returns all completed tasks
func (s *State) GetCompletedTasks() []Task {
	var completed []Task
	for _, task := range s.Tasks {
		if task.Status == "completed" {
			completed = append(completed, task)
		}
	}
	return completed
}

// Summary returns a summary of the current state
func (s *State) Summary() string {
	return fmt.Sprintf(
		"State Summary:\n"+
			"  Query: %s\n"+
			"  Current Agent: %s\n"+
			"  Messages: %d\n"+
			"  Tasks: %d (pending: %d, completed: %d)\n"+
			"  Research Results: %d\n"+
			"  Code Results: %d\n"+
			"  Status: %s\n",
		s.Query,
		s.CurrentAgent,
		len(s.Messages),
		len(s.Tasks),
		len(s.GetPendingTasks()),
		len(s.GetCompletedTasks()),
		len(s.ResearchResults),
		len(s.CodeResults),
		s.Status,
	)
}
