package tools

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// LsTool lists files in a directory
type LsTool struct {
	RootDir string
}

func (t *LsTool) Name() string {
	return "ls"
}

func (t *LsTool) Description() string {
	return "List files in a directory. Input should be a directory path."
}

func (t *LsTool) Call(ctx context.Context, input string) (string, error) {
	path := filepath.Join(t.RootDir, input)
	entries, err := os.ReadDir(path)
	if err != nil {
		return "", fmt.Errorf("failed to read dir: %w", err)
	}

	var result strings.Builder
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}
		prefix := "F"
		if entry.IsDir() {
			prefix = "D"
		}
		result.WriteString(fmt.Sprintf("%s %s %d\n", prefix, entry.Name(), info.Size()))
	}
	return result.String(), nil
}

// ReadFileTool reads a file
type ReadFileTool struct {
	RootDir string
}

func (t *ReadFileTool) Name() string {
	return "read_file"
}

func (t *ReadFileTool) Description() string {
	return "Read a file. Input should be a file path."
}

func (t *ReadFileTool) Call(ctx context.Context, input string) (string, error) {
	path := filepath.Join(t.RootDir, input)
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}
	return string(content), nil
}

// WriteFileTool writes to a file
type WriteFileTool struct {
	RootDir string
}

func (t *WriteFileTool) Name() string {
	return "write_file"
}

func (t *WriteFileTool) Description() string {
	return "Write to a file. Input should be a json string with 'path' and 'content' fields."
}

func (t *WriteFileTool) Call(ctx context.Context, input string) (string, error) {
	// Simple parsing for now, assuming input is "path|content" or similar if we want to avoid complex json parsing in this simple example
	// But let's try to be a bit more robust or just take two args if the framework supports it.
	// For simplicity in this port, let's assume input is "path\ncontent"
	parts := strings.SplitN(input, "\n", 2)
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid input format, expected 'path\\ncontent'")
	}
	path := filepath.Join(t.RootDir, parts[0])
	content := parts[1]

	err := os.WriteFile(path, []byte(content), 0600)
	if err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}
	return "File written successfully", nil
}

// GlobTool finds files matching a pattern
type GlobTool struct {
	RootDir string
}

func (t *GlobTool) Name() string {
	return "glob"
}

func (t *GlobTool) Description() string {
	return "Find files matching a pattern. Input should be a glob pattern."
}

func (t *GlobTool) Call(ctx context.Context, input string) (string, error) {
	var matches []string
	err := filepath.WalkDir(t.RootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(t.RootDir, path)
		matched, err := filepath.Match(input, rel)
		if err != nil {
			return err
		}
		if matched {
			matches = append(matches, rel)
		}
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("failed to glob: %w", err)
	}
	return strings.Join(matches, "\n"), nil
}
