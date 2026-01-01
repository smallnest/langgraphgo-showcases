package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

// SearchTool provides web search functionality
type SearchTool struct {
	APIKey string
	Engine string
}

// NewSearchTool creates a new search tool
func NewSearchTool(apiKey, engine string) *SearchTool {
	return &SearchTool{
		APIKey: apiKey,
		Engine: engine,
	}
}

// Search performs a web search and returns results
func (t *SearchTool) Search(ctx context.Context, query string, maxResults int) ([]Source, error) {
	if t.APIKey == "" {
		return nil, fmt.Errorf("search API key not configured")
	}

	switch t.Engine {
	case "tavily":
		return t.searchTavily(ctx, query, maxResults)
	default:
		return nil, fmt.Errorf("unsupported search engine: %s", t.Engine)
	}
}

// searchTavily performs a search using Tavily API
func (t *SearchTool) searchTavily(ctx context.Context, query string, maxResults int) ([]Source, error) {
	url := "https://api.tavily.com/search"

	requestBody := map[string]any{
		"api_key":        t.APIKey,
		"query":          query,
		"max_results":    maxResults,
		"include_answer": false,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(jsonData)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("search request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Results []struct {
			Title   string  `json:"title"`
			URL     string  `json:"url"`
			Content string  `json:"content"`
			Score   float64 `json:"score"`
		} `json:"results"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	sources := make([]Source, len(result.Results))
	for i, r := range result.Results {
		sources[i] = Source{
			Title:   r.Title,
			URL:     r.URL,
			Content: r.Content,
			Score:   r.Score,
		}
	}

	return sources, nil
}

// CodeExecutor executes code
type CodeExecutor struct {
	Timeout time.Duration
	Verbose bool
}

// NewCodeExecutor creates a new code executor
func NewCodeExecutor(timeoutSeconds int, verbose bool) *CodeExecutor {
	return &CodeExecutor{
		Timeout: time.Duration(timeoutSeconds) * time.Second,
		Verbose: verbose,
	}
}

// ExecutePython executes Python code and returns the result
func (e *CodeExecutor) ExecutePython(ctx context.Context, code string) (*CodeExecutionResult, error) {
	if e.Verbose {
		fmt.Println("Executing Python code:")
		fmt.Println("```python")
		fmt.Println(code)
		fmt.Println("```")
	}

	// Create a temporary file for the code
	tmpFile, err := os.CreateTemp("", "langmanus_*.py")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(code); err != nil {
		return nil, fmt.Errorf("failed to write code: %w", err)
	}
	tmpFile.Close()

	// Execute with timeout
	execCtx, cancel := context.WithTimeout(ctx, e.Timeout)
	defer cancel()

	cmd := exec.CommandContext(execCtx, "python3", tmpFile.Name())
	output, err := cmd.CombinedOutput()

	result := &CodeExecutionResult{
		Code:   code,
		Output: string(output),
	}

	if err != nil {
		if execCtx.Err() == context.DeadlineExceeded {
			result.Error = "execution timeout"
			result.ExitCode = -1
		} else if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
			result.Error = err.Error()
		} else {
			result.Error = err.Error()
			result.ExitCode = -1
		}
	}

	if e.Verbose {
		fmt.Println("Execution result:")
		fmt.Println(result.Output)
		if result.Error != "" {
			fmt.Printf("Error: %s\n", result.Error)
		}
	}

	return result, nil
}

// ExecuteBash executes bash commands and returns the result
func (e *CodeExecutor) ExecuteBash(ctx context.Context, command string) (*CodeExecutionResult, error) {
	if e.Verbose {
		fmt.Println("Executing bash command:")
		fmt.Println("```bash")
		fmt.Println(command)
		fmt.Println("```")
	}

	// Execute with timeout
	execCtx, cancel := context.WithTimeout(ctx, e.Timeout)
	defer cancel()

	cmd := exec.CommandContext(execCtx, "bash", "-c", command)
	output, err := cmd.CombinedOutput()

	result := &CodeExecutionResult{
		Code:   command,
		Output: string(output),
	}

	if err != nil {
		if execCtx.Err() == context.DeadlineExceeded {
			result.Error = "execution timeout"
			result.ExitCode = -1
		} else if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
			result.Error = err.Error()
		} else {
			result.Error = err.Error()
			result.ExitCode = -1
		}
	}

	if e.Verbose {
		fmt.Println("Execution result:")
		fmt.Println(result.Output)
		if result.Error != "" {
			fmt.Printf("Error: %s\n", result.Error)
		}
	}

	return result, nil
}

// ToolRegistry holds all available tools
type ToolRegistry struct {
	Search   *SearchTool
	Executor *CodeExecutor
	Config   *Config
}

// NewToolRegistry creates a new tool registry
func NewToolRegistry(config *Config) *ToolRegistry {
	return &ToolRegistry{
		Search:   NewSearchTool(config.SearchAPIKey, config.SearchEngine),
		Executor: NewCodeExecutor(config.CodeTimeout, config.Verbose),
		Config:   config,
	}
}
