package tools

import (
	"context"
	"strings"
	"sync"
)

// TodoManager manages a list of todos
type TodoManager struct {
	todos []string
	mu    sync.Mutex
}

func NewTodoManager() *TodoManager {
	return &TodoManager{
		todos: []string{},
	}
}

// WriteTodosTool writes a list of todos
type WriteTodosTool struct {
	Manager *TodoManager
}

func (t *WriteTodosTool) Name() string {
	return "write_todos"
}

func (t *WriteTodosTool) Description() string {
	return "Write a list of todos. Input should be a newline-separated list of todos. This overwrites the existing list."
}

func (t *WriteTodosTool) Call(ctx context.Context, input string) (string, error) {
	t.Manager.mu.Lock()
	defer t.Manager.mu.Unlock()

	// Split by newline and filter empty
	lines := strings.Split(input, "\n")
	var newTodos []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			newTodos = append(newTodos, trimmed)
		}
	}
	t.Manager.todos = newTodos
	return "Todos updated successfully", nil
}

// ReadTodosTool reads the current todo list
type ReadTodosTool struct {
	Manager *TodoManager
}

func (t *ReadTodosTool) Name() string {
	return "read_todos"
}

func (t *ReadTodosTool) Description() string {
	return "Read the current todo list."
}

func (t *ReadTodosTool) Call(ctx context.Context, input string) (string, error) {
	t.Manager.mu.Lock()
	defer t.Manager.mu.Unlock()

	if len(t.Manager.todos) == 0 {
		return "No todos", nil
	}
	return strings.Join(t.Manager.todos, "\n"), nil
}
