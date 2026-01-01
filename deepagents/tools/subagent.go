package tools

import (
	"context"
	"fmt"
)

// SubAgentHandler is a function that handles a subagent task
type SubAgentHandler func(ctx context.Context, task string) (string, error)

// TaskTool delegates a task to a subagent
type TaskTool struct {
	Handler SubAgentHandler
}

func (t *TaskTool) Name() string {
	return "task"
}

func (t *TaskTool) Description() string {
	return "Delegate a task to a subagent. Input should be the task description."
}

func (t *TaskTool) Call(ctx context.Context, input string) (string, error) {
	if t.Handler == nil {
		return "", fmt.Errorf("no subagent handler configured")
	}
	return t.Handler(ctx, input)
}
