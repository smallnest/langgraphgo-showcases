# LangManus - Go Implementation

A Go implementation of the LangManus multi-agent AI automation framework using [langgraphgo](https://github.com/smallnest/langgraphgo) and [langchaingo](https://github.com/tmc/langchaingo).

## Overview

LangManus is a community-driven AI automation framework that integrates language models with specialized tools to execute complex tasks through a multi-agent architecture. This Go implementation provides a faithful reproduction of the original Python version.

## Architecture

LangManus uses a **layered multi-agent architecture** with the following roles:

- **Coordinator**: Entry point that analyzes initial requests and routes to appropriate agents
- **Planner**: Analyzes complex tasks and creates execution strategies
- **Supervisor**: Orchestrates worker agents and monitors task completion
- **Researcher**: Conducts information gathering and data analysis using web search
- **Coder**: Handles code generation, modification, and execution (Python/Bash)
- **Browser**: Performs web interactions and information retrieval
- **Reporter**: Generates comprehensive final reports and summaries

## Features

- ✅ **Multi-Agent Orchestration**: Coordinated workflow between specialized agents
- ✅ **LLM Integration**: Support for OpenAI and compatible APIs with multiple model tiers
- ✅ **Web Search**: Integrated Tavily API for research capabilities
- ✅ **Code Execution**: Safe Python and Bash script execution
- ✅ **Task Planning**: Automatic decomposition of complex tasks
- ✅ **Streaming Support**: Real-time updates during execution
- ✅ **Configurable**: Environment-based configuration

## Installation

```bash
cd showcases/langmanus
go build
```

## Configuration

LangManus automatically loads configuration from a `.env` file in the current directory.

### Step 1: Create `.env` file

```bash
# Copy the example configuration
cp .env.example .env

# Edit with your settings
vim .env
```

### Step 2: Configure your `.env`

```bash
# Required
OPENAI_API_KEY=your-api-key-here

# Optional (these are the defaults)
OPENAI_BASE_URL=https://api.openai.com/v1
OPENAI_MODEL=gpt-4o
OPENAI_MODEL_SMALL=gpt-4o-mini
TEMPERATURE=0.7

# Search Configuration (recommended)
SEARCH_API_KEY=your-tavily-api-key
SEARCH_ENGINE=tavily

# Code Execution
ENABLE_CODE_EXECUTION=true
CODE_TIMEOUT=60

# Agent Configuration
MAX_ITERATIONS=15
VERBOSE=true
MAX_CONCURRENT_TASKS=3
```

**Note**: The `.env` file is automatically loaded at startup. You can still override settings with environment variables if needed.

## Usage

### Basic Usage

```bash
# Use default query
./langmanus

# Custom query
./langmanus "Research machine learning trends in 2024 and create a summary"

# Complex task with code execution
./langmanus "Analyze HuggingFace datasets and write Python code to visualize the results"
```

### Programmatic Usage

```go
package main

import (
    "context"
    "log"
)

func main() {
    // Create configuration
    config := NewConfig()

    // Create LangManus instance
    lm, err := NewLangManus(config)
    if err != nil {
        log.Fatal(err)
    }

    // Run query
    ctx := context.Background()
    state, err := lm.Run(ctx, "Your query here")
    if err != nil {
        log.Fatal(err)
    }

    // Access results
    println(state.FinalReport)
}
```

### Streaming Mode

```go
// Stream updates
stateChan, err := lm.Stream(ctx, "Your query here")
if err != nil {
    log.Fatal(err)
}

for state := range stateChan {
    fmt.Printf("Agent: %s\n", state.CurrentAgent)
    fmt.Println(state.Summary())
}
```

## Workflow

1. **Coordinator** receives the query and analyzes the task type
2. **Planner** (if needed) breaks down complex tasks into steps
3. **Supervisor** assigns tasks to specialized workers:
   - **Researcher** for information gathering
   - **Coder** for code generation/execution
   - **Browser** for web interactions
4. **Reporter** synthesizes all results into a final report

```
┌─────────────┐
│ Coordinator │
└──────┬──────┘
       │
       ├─────────────┐
       │             │
┌──────▼──────┐ ┌───▼────────┐
│   Planner   │ │ Researcher │
└──────┬──────┘ └────────────┘
       │
┌──────▼──────┐
│ Supervisor  │
└──────┬──────┘
       │
       ├──────────┬──────────┐
       │          │          │
┌──────▼──────┐ ┌▼─────┐ ┌──▼────┐
│ Researcher  │ │ Coder│ │Browser│
└──────┬──────┘ └┬─────┘ └──┬────┘
       │         │          │
       └─────────┴──────────┘
                 │
          ┌──────▼──────┐
          │  Reporter   │
          └─────────────┘
```

## Examples

### Example 1: Research Task

```bash
./langmanus "What are the latest developments in quantum computing?"
```

This will:
1. Route to Researcher agent
2. Conduct web searches
3. Analyze and synthesize results
4. Generate a comprehensive report

### Example 2: Code Analysis

```bash
./langmanus "Write Python code to analyze a CSV file and create visualizations"
```

This will:
1. Route to Planner
2. Create a plan with steps
3. Coder writes Python code
4. Execute code safely
5. Report results

### Example 3: Complex Multi-Step Task

```bash
./langmanus "Research Python testing frameworks, write example code, and create a comparison report"
```

This will:
1. Planner creates execution strategy
2. Researcher gathers information on testing frameworks
3. Coder writes example code
4. Reporter synthesizes findings and code into final report

## Components

### State Management

The `State` struct tracks:
- Query and messages
- Task planning and execution
- Agent routing history
- Research and code results
- Final report

### Tools

- **SearchTool**: Web search via Tavily API
- **CodeExecutor**: Safe Python/Bash execution with timeout
- **ToolRegistry**: Central registry for all tools

### Agents

Each agent is implemented with:
- Specific prompt templates
- LLM integration
- State transformation logic
- Routing decisions

## Comparison with Original

| Feature | Original (Python) | This (Go) | Status |
|---------|------------------|-----------|---------|
| Multi-agent architecture | ✅ | ✅ | Complete |
| LLM integration | ✅ | ✅ | Complete |
| Web search (Tavily) | ✅ | ✅ | Complete |
| Code execution | ✅ | ✅ | Complete |
| Graph orchestration | LangGraph | langgraphgo | Complete |
| Streaming | ✅ | ✅ | Complete |
| FastAPI server | ✅ | ❌ | Not implemented |
| Browser automation | ✅ | ⚠️ | Partial |

## Requirements

- Go 1.25+
- Python 3.x (for code execution)
- OpenAI API key or compatible endpoint
- Tavily API key (for search functionality)

## License

MIT License - Same as the original LangManus project

## Credits

This is a Go implementation inspired by the original [LangManus](https://github.com/Darwin-lfl/langmanus) by Darwin-lfl.

Built with:
- [langgraphgo](https://github.com/smallnest/langgraphgo) - Multi-agent workflow orchestration
- [langchaingo](https://github.com/tmc/langchaingo) - LLM integration

## Contributing

Contributions are welcome! This project follows the same philosophy as the original LangManus - giving back to the open source community.

## Troubleshooting

### Search not working
- Ensure `SEARCH_API_KEY` is set
- Check Tavily API quota

### Code execution fails
- Ensure Python 3 is installed and in PATH
- Check `CODE_TIMEOUT` setting
- Verify `ENABLE_CODE_EXECUTION=true`

### LLM errors
- Verify `OPENAI_API_KEY` is valid
- Check `OPENAI_BASE_URL` if using custom endpoint
- Ensure model names are correct

## Roadmap

- [ ] Add more search engines (Serp, Jina)
- [ ] Implement FastAPI server mode
- [ ] Add browser automation
- [ ] Enhanced error handling
- [ ] Visualization tools
- [ ] Persistence and checkpointing
- [ ] Multi-language code execution
