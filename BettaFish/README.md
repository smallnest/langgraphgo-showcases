# BettaFish (Go Implementation)

This is a **complete replication** of the [BettaFish](https://github.com/666ghj/BettaFish) project in Go, using [langgraphgo](https://github.com/smallnest/langgraphgo) and [langchaingo](https://github.com/tmc/langchaingo).

It implements the full multi-agent architecture for deep public opinion analysis.

## Features

- **QueryEngine**: 
  - Generates a structured research plan (outline).
  - Performs deep web search using Tavily API.
  - Implements a **Reflection Loop** to iteratively refine search results and summaries.
  - Uses specialized prompts for searching, summarizing, and reflecting.
- **MediaEngine**: 
  - Searches for relevant images using Tavily's image search capabilities.
- **InsightEngine**: 
  - (Simulated) Mines internal data for insights.
- **ForumEngine**: 
  - Facilitates an LLM-driven discussion between "NewsAgent", "MediaAgent", and "Moderator" to synthesize findings.
- **ReportEngine**: 
  - Compiles all findings into a comprehensive Markdown report.

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
