package agent

import (
	"fmt"
	"os"

	"github.com/smallnest/langgraphgo/graph"
	"github.com/smallnest/langgraphgo/prebuilt"
	"github.com/smallnest/langgraphgo/showcases/deepagents/tools"
	"github.com/tmc/langchaingo/llms"
	ltools "github.com/tmc/langchaingo/tools"
)

// DeepAgentOptions contains options for creating a deep agent
type DeepAgentOptions struct {
	RootDir         string
	SystemPrompt    string
	SubAgentHandler tools.SubAgentHandler
}

// Option is a function that configures DeepAgentOptions
type Option func(*DeepAgentOptions)

// WithRootDir sets the root directory for filesystem tools
func WithRootDir(rootDir string) Option {
	return func(o *DeepAgentOptions) {
		o.RootDir = rootDir
	}
}

// WithSystemPrompt sets the system prompt
func WithSystemPrompt(prompt string) Option {
	return func(o *DeepAgentOptions) {
		o.SystemPrompt = prompt
	}
}

// WithSubAgentHandler sets the handler for subagent tasks
func WithSubAgentHandler(handler tools.SubAgentHandler) Option {
	return func(o *DeepAgentOptions) {
		o.SubAgentHandler = handler
	}
}

// CreateDeepAgent creates a new deep agent
func CreateDeepAgent(model llms.Model, opts ...Option) (*graph.StateRunnable[map[string]any], error) {
	options := DeepAgentOptions{
		RootDir:      ".",
		SystemPrompt: "You are a helpful deep agent.",
	}
	for _, opt := range opts {
		opt(&options)
	}

	// Ensure root dir exists
	if err := os.MkdirAll(options.RootDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create root dir: %w", err)
	}

	// Initialize tools
	todoManager := tools.NewTodoManager()

	agentTools := []ltools.Tool{
		&tools.LsTool{RootDir: options.RootDir},
		&tools.ReadFileTool{RootDir: options.RootDir},
		&tools.WriteFileTool{RootDir: options.RootDir},
		&tools.GlobTool{RootDir: options.RootDir},
		&tools.WriteTodosTool{Manager: todoManager},
		&tools.ReadTodosTool{Manager: todoManager},
		&tools.TaskTool{Handler: options.SubAgentHandler},
	}

	// Create agent using map state convenience function
	agent, err := prebuilt.CreateAgentMap(model, agentTools,
		prebuilt.WithSystemMessage(options.SystemPrompt),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create agent: %w", err)
	}

	return agent, nil
}
