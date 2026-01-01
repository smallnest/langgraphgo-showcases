# GPT Researcher - Go Implementation

A Go implementation of [gpt-researcher](https://github.com/assafelovic/gpt-researcher) using the langgraphgo framework and langchaingo library. This is an autonomous research agent designed to conduct comprehensive research on any given topic and generate detailed, factual reports with citations.

## Overview

GPT Researcher is a multi-agent system that automates the research process by:
1. **Planning**: Generating focused research questions from a query
2. **Executing**: Gathering information from multiple web sources
3. **Publishing**: Synthesizing findings into a comprehensive research report

The system produces detailed reports (2000+ words) with information aggregated from 20+ sources, complete with citations and references.

## Architecture

The system consists of three main agents working in a pipeline:

```
┌─────────────────────────────────────────────────────────────┐
│                    GPT Researcher                           │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌──────────────┐      ┌──────────────┐      ┌──────────┐ │
│  │   Planner    │─────▶│  Execution   │─────▶│Publisher │ │
│  │    Agent     │      │    Agent     │      │  Agent   │ │
│  └──────────────┘      └──────────────┘      └──────────┘ │
│        │                      │                     │      │
│        ▼                      ▼                     ▼      │
│  Generate Research       Gather & Summarize    Generate   │
│  Questions              Information from Web    Report    │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### 1. Planner Agent

**Responsibility**: Generate comprehensive research questions

- Analyzes the research query
- Creates 5-10 focused research questions
- Ensures questions cover multiple perspectives
- Questions collectively form an objective understanding

**Example Questions** for "What are the latest advances in LLMs in 2024?":
1. What are the major architectural innovations in LLMs released in 2024?
2. How have reasoning capabilities evolved in recent LLM models?
3. What improvements have been made in LLM efficiency and cost?
4. What are the latest applications and use cases for LLMs?
5. What ethical and safety advances have been introduced?

### 2. Execution Agent

**Responsibility**: Research each question and gather information

For each research question:
- Performs web search using Tavily API
- Retrieves top relevant sources (up to 20 per question)
- Scrapes and extracts content from web pages
- Summarizes each source using LLM
- Tracks citations and relevance scores

**Tools Used**:
- `tavily_search`: Web search with relevance ranking
- `web_scraper`: Content extraction from URLs
- `summarizer`: LLM-based summarization

### 3. Publisher Agent

**Responsibility**: Synthesize findings into final report

- Aggregates all summaries and findings
- Organizes information by themes
- Generates comprehensive 2000+ word report
- Includes citations and references
- Provides objective analysis and insights

## Features

✅ **Multi-agent Research Pipeline**: Planner → Executor → Publisher
✅ **Comprehensive Reports**: 2000+ words with detailed analysis
✅ **Source Aggregation**: Information from 20+ credible sources
✅ **Automatic Citations**: Numbered references throughout report
✅ **Web Search Integration**: Tavily API for relevant results
✅ **Flexible Configuration**: Customize models, parameters, output
✅ **Progress Tracking**: Verbose mode shows research progress
✅ **Multiple Report Types**: Research, outline, or resource reports

## Requirements

- Go 1.21 or higher
- OpenAI API key
- Tavily API key (for web search)

## Installation

```bash
# Navigate to the showcase directory
cd showcases/gpt_researcher

# Set up environment variables
export OPENAI_API_KEY="your-openai-api-key"
export TAVILY_API_KEY="your-tavily-api-key"

# Install dependencies
go mod download
```

## Usage

### Basic Usage

```bash
# Run with default query
go run *.go

# Run with custom query
go run *.go "Your research question here"
```

### Example Queries

```bash
# Technology research
go run *.go "What are the latest advances in quantum computing?"

# Market research
go run *.go "What is the current state of the electric vehicle market?"

# Academic research
go run *.go "What are the recent breakthroughs in CRISPR gene editing?"

# Business research
go run *.go "What are the emerging trends in sustainable energy?"
```

### Configuration

Customize behavior using environment variables:

```bash
# Model configuration
export GPT_MODEL="gpt-4"                    # Main model for research
export REPORT_MODEL="gpt-4"                 # Model for final report
export SUMMARY_MODEL="gpt-3.5-turbo"        # Model for summarization

# Search parameters
export MAX_SEARCH_RESULTS="10"              # Results per search
export MAX_SOURCES_TO_USE="20"              # Total sources to use
export MAX_QUESTIONS="5"                    # Research questions to generate

# Report configuration
export REPORT_TYPE="research_report"        # Type of report
export REPORT_FORMAT="markdown"             # Output format
export OUTPUT_DIR="./output"                # Save location
export SAVE_INTERMEDIATE="true"             # Save report to file

# Verbosity
export VERBOSE="true"                       # Show progress
```

## How It Works

### Research Workflow

1. **Initialize**: User provides research query
2. **Plan**: Planner Agent generates 5 research questions
3. **Execute**: For each question:
   - Search web using Tavily
   - Retrieve top 10-20 results
   - Scrape and summarize each source
   - Track citations
4. **Publish**: Publisher Agent:
   - Groups findings by question
   - Synthesizes comprehensive report
   - Adds citations and references
5. **Complete**: Return final report

### Example Output

```
================================================================================
GPT RESEARCHER
================================================================================

📋 Research Query: What are the latest advances in large language models in 2024?

🎯 [Planner Agent] Generating research questions...
✅ [Planner Agent] Generated 5 research questions:
   1. What are the major architectural innovations in LLMs released in 2024?
   2. How have reasoning capabilities evolved in recent LLM models?
   3. What improvements have been made in LLM efficiency and cost?
   4. What are the latest applications and use cases for LLMs?
   5. What ethical and safety advances have been introduced?

📚 [Execution Agent] Starting research for 5 questions...

--- Question 1/5 ---
🔍 [Execution Agent] Researching: What are the major architectural innovations...
   Found 10 search results
   ✅ Summarized: Transformer Architecture Evolution in 2024
   ✅ Summarized: Mixture of Experts: New Paradigms
   ...

📝 [Publisher Agent] Generating final research report...
✅ [Publisher Agent] Report generated (8547 characters)

================================================================================
RESEARCH COMPLETE
================================================================================

Statistics:
- Research Questions: 5
- Sources Consulted: 23
- Summaries Generated: 23
- Report Length: 8547 characters
- Duration: 3.2 minutes

================================================================================
FINAL RESEARCH REPORT
================================================================================

# Research Report

## Metadata

- **Research Query**: What are the latest advances in large language models in 2024?
- **Date**: December 6, 2024
- **Total Sources**: 23
- **Research Duration**: 3.2 minutes

---

## Executive Summary

The year 2024 has witnessed remarkable advances in large language models (LLMs)...

[Full report continues...]

## References

[1] Transformer Architecture Evolution in 2024 - https://...
[2] Mixture of Experts: New Paradigms - https://...
...
```

## Report Types

### Research Report (Default)
Comprehensive academic-style report with:
- Executive summary
- Detailed sections organized by themes
- In-depth analysis and evidence
- Clear conclusions and recommendations
- 2000+ words

### Outline Report
Structured outline format with:
- Hierarchical organization
- Bullet points and headings
- Concise key points
- Roadmap for deeper exploration

### Resource Report
Curated resource guide with:
- Categorized sources by type
- Brief annotations for each resource
- Highlighted authoritative sources
- Access information and context

Set via: `export REPORT_TYPE="outline_report"`

## Project Structure

```
gpt_researcher/
├── config.go              # Configuration management
├── state.go               # Research state definitions
├── tools.go               # Web search, scraper, summarizer
├── planner_agent.go       # Question generation agent
├── execution_agent.go     # Information gathering agent
├── publisher_agent.go     # Report generation agent
├── gpt_researcher.go      # Main workflow orchestration
├── main.go                # Example application
├── go.mod                 # Go module definition
├── README.md              # This file
└── README_CN.md           # Chinese documentation
```

## Comparison with Original Python Implementation

| Feature | Python (assafelovic/gpt-researcher) | Go (This Implementation) |
|---------|-----------------------------------|--------------------------|
| Planner Agent | ✅ | ✅ |
| Execution Agent | ✅ | ✅ |
| Publisher Agent | ✅ | ✅ |
| Tavily Search | ✅ | ✅ |
| Web Scraping | ✅ | ✅ (Simplified) |
| PDF Support | ✅ | ⚠️ (Planned) |
| Source Citations | ✅ | ✅ |
| Multiple Report Types | ✅ | ✅ |
| FastAPI Backend | ✅ | ❌ (CLI only) |
| NextJS Frontend | ✅ | ❌ (CLI only) |
| Export to PDF/DOCX | ✅ | ⚠️ (Planned) |
| Local Documents | ✅ | ⚠️ (Planned) |

## Best Practices

### 1. Write Clear Queries

✅ **Good**:
- "What are the latest advances in quantum computing in 2024?"
- "How is artificial intelligence being used in healthcare?"
- "What are the economic impacts of remote work?"

❌ **Too Vague**:
- "Tell me about AI"
- "What's new in tech?"

### 2. Adjust Configuration

For **quick research** (faster, cheaper):
```bash
export MAX_QUESTIONS="3"
export MAX_SOURCES_TO_USE="10"
export SUMMARY_MODEL="gpt-3.5-turbo"
export GPT_MODEL="gpt-3.5-turbo"
```

For **deep research** (thorough, detailed):
```bash
export MAX_QUESTIONS="10"
export MAX_SOURCES_TO_USE="30"
export GPT_MODEL="gpt-4"
export REPORT_MODEL="gpt-4"
```

### 3. Monitor API Costs

- Each research session makes ~20-50 API calls
- Use cheaper models (gpt-3.5-turbo) for summaries
- Limit `MAX_QUESTIONS` and `MAX_SOURCES_TO_USE`
- Enable `VERBOSE="true"` to track progress

### 4. Review Citations

Always verify sources in the References section:
- Check URL validity
- Assess source credibility
- Review original content for accuracy

## Troubleshooting

### API Key Errors

```
Warning: OPENAI_API_KEY not set
Warning: TAVILY_API_KEY not set
```

**Solution**: Set environment variables:
```bash
export OPENAI_API_KEY="sk-..."
export TAVILY_API_KEY="tvly-..."
```

### Rate Limiting

If you hit rate limits:
- Reduce `MAX_QUESTIONS` (e.g., to 3)
- Reduce `MAX_SOURCES_TO_USE` (e.g., to 10)
- Add delays between requests
- Use different API tiers

### Empty or Poor Reports

If reports are inadequate:
- Verify Tavily API key is valid
- Make query more specific
- Increase `MAX_QUESTIONS` and `MAX_SOURCES_TO_USE`
- Try different models (gpt-4 vs gpt-3.5-turbo)

### Web Scraping Failures

Some sites may block scraping:
- This is normal behavior
- System will skip failed sources
- Increase `MAX_SOURCES_TO_USE` for redundancy

## Advanced Usage

### Programmatic Usage

```go
package main

import (
    "context"
    "fmt"
    "log"
)

func main() {
    // Create config
    config := NewConfig()
    config.Verbose = false
    config.MaxQuestions = 3

    // Create researcher
    researcher, err := NewGPTResearcher(config)
    if err != nil {
        log.Fatal(err)
    }

    // Conduct research
    ctx := context.Background()
    state, err := researcher.ConductResearch(ctx, "Your query here")
    if err != nil {
        log.Fatal(err)
    }

    // Access results
    fmt.Printf("Questions: %v\n", state.Questions)
    fmt.Printf("Sources: %d\n", len(state.Sources))
    fmt.Printf("Report: %s\n", state.FinalReport)
}
```

### Custom Tool Integration

Extend with custom tools:

```go
// Add a custom search tool
type CustomSearchTool struct{}

func (t *CustomSearchTool) Name() string {
    return "custom_search"
}

func (t *CustomSearchTool) Description() string {
    return "Search using custom API"
}

func (t *CustomSearchTool) Call(ctx context.Context, input string) (string, error) {
    // Your custom implementation
    return results, nil
}
```

## Performance

### Typical Research Session

- **Duration**: 3-5 minutes
- **API Calls**: 25-50 (depending on configuration)
- **Sources**: 15-25 unique sources
- **Report Length**: 2000-4000 words
- **Cost**: ~$0.50-2.00 (using GPT-4)

### Optimization Tips

1. **Use gpt-3.5-turbo** for summaries (80% cost reduction)
2. **Limit questions** to 3-5 for most queries
3. **Batch requests** when possible
4. **Cache results** for repeated queries

## Future Enhancements

Planned features:
- [ ] PDF document analysis
- [ ] Local document search
- [ ] Export to PDF/DOCX
- [ ] Image analysis and inclusion
- [ ] Multi-language support
- [ ] Custom report templates
- [ ] Web UI / API server
- [ ] Parallel question execution
- [ ] Source quality scoring
- [ ] Fact verification

## License

This implementation follows the same license as the langgraphgo project.

## References

- [Original Python gpt-researcher](https://github.com/assafelovic/gpt-researcher)
- [LangGraph Documentation](https://python.langchain.com/docs/langgraph)
- [Tavily Search API](https://www.tavily.com/)
- [LangChain Go](https://github.com/tmc/langchaingo)

## Contributing

Contributions are welcome! Areas for improvement:
- Enhanced web scraping
- PDF/document processing
- Additional export formats
- Performance optimizations
- Test coverage

## Support

For issues and questions:
- Check the troubleshooting section above
- Review the examples in this README
- Open an issue on GitHub

---

**Built with**:
- [langgraphgo](https://github.com/smallnest/langgraphgo) - Graph-based agent orchestration
- [langchaingo](https://github.com/tmc/langchaingo) - LLM integration
- [Tavily](https://www.tavily.com/) - Web search API
- [OpenAI](https://openai.com/) - Language models
