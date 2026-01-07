package tool

import (
	"context"
	"fmt"
	"sync"
)

// Registry manages all available search tools.
type Registry struct {
	mu     sync.RWMutex
	tools  map[string]SearchTool
	toolsByCategory map[string][]SearchTool
}

// Category represents a search source category.
type Category string

const (
	CategorySocial     Category = "social"      // 社交媒体和社区 (知乎、微信)
	CategoryCode       Category = "code"        // 代码和开发 (GitHub)
	CategoryAcademic   Category = "academic"    // 学术和科研 (Google Scholar)
	CategoryGeneral    Category = "general"     // 通用搜索
	CategoryAll        Category = "all"         // 所有分类
)

// NewRegistry creates a new tool registry.
func NewRegistry() *Registry {
	r := &Registry{
		tools:  make(map[string]SearchTool),
		toolsByCategory: make(map[string][]SearchTool),
	}
	return r
}

// Register adds a search tool to the registry.
func (r *Registry) Register(category Category, tool SearchTool) error {
	if tool == nil {
		return fmt.Errorf("tool cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	name := tool.Name()
	if _, exists := r.tools[name]; exists {
		return fmt.Errorf("tool '%s' already registered", name)
	}

	r.tools[name] = tool
	r.toolsByCategory[string(category)] = append(r.toolsByCategory[string(category)], tool)

	return nil
}

// Get retrieves a tool by name.
func (r *Registry) Get(name string) (SearchTool, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tool, exists := r.tools[name]
	return tool, exists
}

// List returns all registered tool names.
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.tools))
	for name := range r.tools {
		names = append(names, name)
	}
	return names
}

// GetByCategory returns all tools in a specific category.
func (r *Registry) GetByCategory(category Category) []SearchTool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if category == CategoryAll {
		tools := make([]SearchTool, 0, len(r.tools))
		for _, tool := range r.tools {
			tools = append(tools, tool)
		}
		return tools
	}

	tools, exists := r.toolsByCategory[string(category)]
	if !exists {
		return []SearchTool{}
	}

	// Return a copy to avoid external modification
	result := make([]SearchTool, len(tools))
	copy(result, tools)
	return result
}

// SearchAll searches using all tools in the specified category.
func (r *Registry) SearchAll(ctx context.Context, category Category, query string, limitPerTool int) ([]SearchResult, error) {
	tools := r.GetByCategory(category)
	if len(tools) == 0 {
		return []SearchResult{}, nil
	}

	type result struct {
		results []SearchResult
		err     error
	}

	resultsCh := make(chan result, len(tools))

	// Search concurrently using all tools
	for _, tool := range tools {
		go func(t SearchTool) {
			searchResults, err := t.Search(ctx, query, limitPerTool)
			resultsCh <- result{results: searchResults, err: err}
		}(tool)
	}

	// Collect results
	var allResults []SearchResult
	var errs []error

	for range tools {
		res := <-resultsCh
		if res.err != nil {
			errs = append(errs, res.err)
			// Continue even if some tools fail
		} else if len(res.results) > 0 {
			allResults = append(allResults, res.results...)
		}
	}

	return allResults, nil
}

// DescribeAll returns descriptions of all registered tools.
func (r *Registry) DescribeAll() string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var output string
	output += "# 可用的搜索工具\n\n"

	for category, tools := range r.toolsByCategory {
		output += fmt.Sprintf("## %s\n\n", categoryDisplayName(category))
		for _, tool := range tools {
			output += fmt.Sprintf("- **%s**: %s\n", tool.Name(), tool.Description())
		}
		output += "\n"
	}

	return output
}

// categoryDisplayName returns a human-readable name for a category.
func categoryDisplayName(category string) string {
	switch category {
	case "social":
		return "社交媒体和社区"
	case "code":
		return "代码和开发"
	case "academic":
		return "学术和科研"
	case "general":
		return "通用搜索"
	default:
		return category
	}
}

// DefaultRegistry is the default global registry.
var DefaultRegistry = NewRegistry()

// RegisterDefaultTools registers all default search tools.
func RegisterDefaultTools() error {
	// Register WeChat search
	wechat, err := NewWeChatSearch()
	if err != nil {
		return fmt.Errorf("failed to create WeChat search: %w", err)
	}
	if err := DefaultRegistry.Register(CategorySocial, wechat); err != nil {
		return err
	}

	// Register Zhihu search
	zhihu, err := NewZhihuSearch()
	if err != nil {
		return fmt.Errorf("failed to create Zhihu search: %w", err)
	}
	if err := DefaultRegistry.Register(CategorySocial, zhihu); err != nil {
		return err
	}

	// Register GitHub search
	github, err := NewGitHubSearch()
	if err != nil {
		return fmt.Errorf("failed to create GitHub search: %w", err)
	}
	if err := DefaultRegistry.Register(CategoryCode, github); err != nil {
		return err
	}

	// Register Scholar search
	scholar, err := NewScholarSearch()
	if err != nil {
		return fmt.Errorf("failed to create Scholar search: %w", err)
	}
	if err := DefaultRegistry.Register(CategoryAcademic, scholar); err != nil {
		return err
	}

	// Register Tavily search (general purpose AI search)
	tavily, err := NewTavilySearch()
	if err != nil {
		return fmt.Errorf("failed to create Tavily search: %w", err)
	}
	if err := DefaultRegistry.Register(CategoryGeneral, tavily); err != nil {
		return err
	}

	return nil
}

func init() {
	// Automatically register default tools on package init
	if err := RegisterDefaultTools(); err != nil {
		// Log error but don't panic - this is in init
		fmt.Printf("Warning: failed to register default search tools: %v\n", err)
	}
}
