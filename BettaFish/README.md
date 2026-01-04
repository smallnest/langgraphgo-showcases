# BettaFish (Go Implementation)

This is a **complete replication** of the [BettaFish](https://github.com/666ghj/BettaFish) project in Go, using [langgraphgo](https://github.com/smallnest/langgraphgo) and [langchaingo](https://github.com/tmc/langchaingo).

It implements the full multi-agent architecture for deep public opinion analysis, synchronized with the latest upstream Python implementation.

## Latest Updates (2026-01)

### New Features
- **Enhanced Error Handling**: Added retry mechanism with exponential backoff and JSON repair logic
- **GraphRAG Foundation**: Added basic structures for knowledge graph and RAG support
- **Improved Search Tools**: Extended search options with platform-specific parameters
- **Structured Logging**: New logging utility with execution tracking
- **Sentiment Analysis Support**: Added sentiment fields to search results

### Synchronized with Upstream
- Updated prompts to match the latest Python implementation
- Enhanced search tool descriptions with platform-specific guidance
- Improved "ground-level" keyword optimization in prompts

## Features

- **QueryEngine**:
  - Generates a structured research plan (outline).
  - Performs deep web search using Tavily API.
  - Implements a **Reflection Loop** to iteratively refine search results and summaries.
  - Uses specialized prompts for searching, summarizing, and reflecting.
  - Supports 6 search tools: basic_search_news, deep_search_news, search_news_last_24_hours, search_news_last_week, search_images_for_news, search_news_by_date

- **MediaEngine**:
  - Searches for relevant images using Tavily's image search capabilities.

- **InsightEngine**:
  - (Simulated) Mines internal data for insights.
  - Supports sentiment analysis integration.

- **ForumEngine**:
  - Facilitates an LLM-driven discussion between "NewsAgent", "MediaAgent", and "Moderator" to synthesize findings.

- **ReportEngine**:
  - Compiles all findings into a comprehensive Markdown report.

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                          BettaFish Go                            │
├─────────────────────────────────────────────────────────────────┤
│                                                                   │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐     │
│  │ QueryEngine  │───▶│ MediaEngine  │───▶│ InsightEngine│     │
│  └──────────────┘    └──────────────┘    └──────────────┘     │
│         │                                        │            │
│         ▼                                        ▼            │
│  ┌──────────────┐                         ┌──────────────┐     │
│  │ ForumEngine  │◀────────────────────────│  (State)    │     │
│  └──────────────┘                         └──────────────┘     │
│         │                                                       │
│         ▼                                                       │
│  ┌──────────────┐                                              │
│  │ ReportEngine │───▶ Final Report                             │
│  └──────────────┘                                              │
│                                                                   │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │              GraphRAG Support (Foundation)              │   │
│  │  • GraphRAGNode, GraphRAGEdge                           │   │
│  │  • Knowledge graph state management                     │   │
│  │  • Sentiment analysis integration                        │   │
│  └─────────────────────────────────────────────────────────┘   │
│                                                                   │
└─────────────────────────────────────────────────────────────────┘
```

## Project Structure

```
BettaFish/
├── main.go                 # Entry point
├── schema/
│   └── state.go           # State management with GraphRAG support
├── query_engine/
│   ├── agent.go           # Query analysis node
│   ├── prompts.go         # LLM prompts (synced with upstream)
│   └── tools.go           # Enhanced Tavily search integration
├── media_engine/
│   ├── agent.go           # Media search node
│   └── prompts.go         # Image search prompts
├── insight_engine/
│   ├── agent.go           # Insight generation node
│   └── prompts.go         # Analysis prompts (synced with upstream)
├── forum_engine/
│   └── agent.go           # Multi-agent discussion simulation
├── report_engine/
│   ├── agent.go           # Report generation node
│   ├── prompts.go         # Report compilation prompts
│   └── template.go        # Go template-based Markdown generation
├── utils/
│   ├── errors.go          # Enhanced error handling with retry
│   └── logger.go          # Structured logging with execution tracking
├── mind_spider/
│   └── spider.go          # Social media search simulation
└── sentiment_model/
    └── model.go           # LLM-based sentiment analysis
```

## Prerequisites

You need the following API keys:
- `OPENAI_API_KEY`: For LLM inference (GPT-4o recommended, or any OpenAI-compatible API).
- `TAVILY_API_KEY`: For web search and image search.

**Optional**: For using alternative LLM providers (e.g., DeepSeek, Azure OpenAI, or any OpenAI-compatible API):
- `OPENAI_API_BASE`: Set to your custom API endpoint. This allows you to use any OpenAI-compatible service. For example:
  - DeepSeek: `https://api.deepseek.com/v1`
  - Azure OpenAI: `https://YOUR_RESOURCE_NAME.openai.azure.com/openai/deployments/YOUR_DEPLOYMENT_NAME`
  - Local models (Ollama, vLLM, etc.): `http://localhost:11434/v1`
- `OPENAI_MODEL`: Override the default model name if needed. This is particularly useful when:
  - Switching between different OpenAI models (e.g., `gpt-4o`, `gpt-4o-mini`, `gpt-4-turbo`)
  - Using alternative providers with specific model names (e.g., `deepseek-chat`, `claude-3-haiku`, etc.)

## Usage

### Basic Usage (OpenAI)

```bash
export OPENAI_API_KEY="sk-..."
export TAVILY_API_KEY="tvly-..."
go run showcases/BettaFish/main.go "Your Research Topic"
```

### Using Alternative Providers (e.g., DeepSeek)

```bash
export OPENAI_API_KEY="your-deepseek-api-key"
export OPENAI_API_BASE="https://api.deepseek.com/v1"
export OPENAI_MODEL="deepseek-chat"  # Specify the model name
export TAVILY_API_KEY="tvly-..."
go run showcases/BettaFish/main.go "Your Research Topic"
```

### Using Different OpenAI Models

```bash
export OPENAI_API_KEY="sk-..."
export OPENAI_MODEL="gpt-4o-mini"  # or gpt-4o, gpt-4-turbo, etc.
export TAVILY_API_KEY="tvly-..."
go run showcases/BettaFish/main.go "Your Research Topic"
```

### Using Local Models (Ollama example)

```bash
export OPENAI_API_KEY="ollama"  # Can be any value for local models
export OPENAI_API_BASE="http://localhost:11434/v1"
export OPENAI_MODEL="llama3.1"  # Specify the model name
export TAVILY_API_KEY="tvly-..."
go run showcases/BettaFish/main.go "Your Research Topic"
```

## Key Differences from Python Implementation

| Aspect | Python Version | Go Version |
|--------|----------------|------------|
| **Framework** | LangGraph (Python) | langgraphgo (Go) |
| **UI** | Web UI (Flask/Streamlit) | CLI tool |
| **Database** | Local PostgreSQL/MySQL | Simulated via Tavily API |
| **Sentiment Analysis** | Local ML models | LLM-based |
| **Report Format** | HTML/PDF/Markdown | Markdown |
| **Concurrency** | True async agents | Sequential graph execution |
| **Deployment** | Docker Compose | Single binary |

## Roadmap

- [ ] Add GraphRAG query engine integration
- [ ] Implement true concurrent agent execution
- [ ] Add HTML/PDF report generation
- [ ] Implement local database support
- [ ] Add Web UI (using Go web frameworks)

## Contributing

This project is synchronized with the upstream Python BettaFish implementation. When updating:

1. Check the [upstream repository](https://github.com/666ghj/BettaFish) for prompt changes
2. Update corresponding `prompts.go` files
3. Ensure error handling follows the new retry mechanism
4. Test with different LLM providers

## License

This project follows the same license as the upstream BettaFish project (GPL-2.0).
