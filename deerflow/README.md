# DeerFlow - Deep Research Agent

A Go implementation of the [ByteDance DeerFlow](https://github.com/bytedance/deer-flow) deep research agent, built using [langgraphgo](https://github.com/smallnest/langgraphgo) and [langchaingo](https://github.com/tmc/langchaingo).

DeerFlow is an intelligent multi-agent research system that autonomously conducts deep research on any topic, generates comprehensive reports, and optionally creates podcast scripts for engaging content delivery.

## Overview

DeerFlow orchestrates multiple AI agents to perform structured research:

```
User Query → Planner → Researcher → Reporter → (Optional) Podcast → Final Output
```

The system breaks down complex research tasks, gathers information systematically, and synthesizes findings into professional, well-formatted reports.

## Features

### 🎯 Multi-Agent Architecture
- **Planner Agent**: Decomposes queries into structured research plans
- **Researcher Agent**: Executes each research step using LLM
- **Reporter Agent**: Synthesizes findings into comprehensive HTML reports
- **Podcast Agent**: Generates engaging podcast scripts (optional)

### 🌐 Modern Web Interface
- **Real-time Progress**: Live updates using Server-Sent Events (SSE)
- **Dark Theme UI**: Professional, eye-friendly interface
- **Research History**: View and replay past research sessions
- **Result Caching**: Instant replay of previous queries

### 💻 Dual Operation Modes
- **Web Server**: Interactive browser-based interface
- **CLI Mode**: Quick command-line execution

### 📊 Rich Output Formats
- **HTML Reports**: Well-structured, styled research reports
- **Podcast Scripts**: Conversational content for audio production
- **Persistent Storage**: Automatic saving of research results

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                       DeerFlow                              │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌──────────┐    ┌──────────┐    ┌──────────┐    ┌──────┐│
│  │ Planner  │───▶│Researcher│───▶│ Reporter │───▶│Podcast││
│  │  Agent   │    │  Agent   │    │  Agent   │    │ Agent ││
│  └──────────┘    └──────────┘    └──────────┘    └──────┘│
│       │               │                │              │    │
│       ▼               ▼                ▼              ▼    │
│  Generate Plan   Execute Steps   Create Report   Generate │
│  from Query      Using LLM       in HTML         Script   │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### Workflow

1. **Planning Phase**
   - User submits a research query
   - Planner Agent analyzes and creates step-by-step research plan
   - Detects if podcast generation is requested

2. **Research Phase**
   - Researcher Agent executes each step
   - Gathers information using LLM
   - Collects findings for each research step

3. **Reporting Phase**
   - Reporter Agent synthesizes all research results
   - Generates well-formatted HTML report
   - Includes proper structure and styling

4. **Podcast Phase** (Optional)
   - Podcast Agent creates conversational script
   - Formats content for audio delivery
   - Maintains engagement and flow

## Prerequisites

- **Go**: Version 1.21 or higher
- **API Key**: OpenAI-compatible API (OpenAI, DeepSeek, etc.)
- **Browser**: Modern web browser for UI (Chrome, Firefox, Safari, Edge)

## Installation

```bash
# Navigate to the deerflow directory
cd showcases/deerflow

# Set up environment variables
export OPENAI_API_KEY="your-api-key-here"

# Optional: If using DeepSeek or another provider
export OPENAI_API_BASE="https://api.deepseek.com/v1"

# Build the application
go build -o deerflow .
```

## Usage

### Web Interface (Recommended)

Start the web server:

```bash
./deerflow
```

Then open your browser and navigate to:
```
http://localhost:8085
```

**Web Interface Features:**
- Enter your research query in the input box
- Watch real-time progress updates
- View formatted HTML reports
- Access research history
- Replay previous searches instantly

### Command Line Interface

For quick, one-off queries:

```bash
# Basic usage
./deerflow "Your research question here"

# Example queries
./deerflow "What are the latest advances in quantum computing?"
./deerflow "Explain the impact of AI on healthcare"
./deerflow "What is the current state of renewable energy?"
```

### Example Queries

**Technology Research:**
```bash
./deerflow "What are the breakthrough developments in AI in 2024?"
```

**Scientific Research:**
```bash
./deerflow "What are the recent discoveries about Mars exploration?"
```

**Business Research:**
```bash
./deerflow "What are the emerging trends in e-commerce?"
```

**With Podcast Generation:**
```bash
./deerflow "Create a podcast about blockchain technology"
./deerflow "生成关于人工智能的播客脚本"
```

## Configuration

### Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `OPENAI_API_KEY` | OpenAI API key | None | ✅ Yes |
| `OPENAI_API_BASE` | API base URL | OpenAI default | ❌ No |

### Server Configuration

The web server runs on port **8085** by default. To change this, modify `main.go`:

```go
server := &http.Server{
    Addr: ":8085",  // Change port here
    ReadHeaderTimeout: 3 * time.Second,
}
```

## Project Structure

```
deerflow/
├── main.go              # Entry point, HTTP server, CLI handler
├── graph.go             # Graph structure and state definitions
├── nodes.go             # Agent implementations (Planner, Researcher, Reporter, Podcast)
├── nginx.conf           # Nginx configuration (for production deployment)
├── web/                 # Frontend assets
│   ├── index.html       # Main web interface
│   ├── styles.css       # UI styling
│   └── script.js        # Client-side JavaScript
├── data/                # Research results storage (auto-created)
│   └── [query]/         # One folder per unique query
│       ├── metadata.json    # Query metadata
│       ├── logs.json        # Research process logs
│       ├── report.html      # Generated HTML report
│       └── podcast.txt      # Podcast script (if generated)
└── README.md            # This file
```

## How It Works

### 1. Planner Agent

**Input**: User query
**Process**:
- Analyzes the query to understand research scope
- Creates a structured, step-by-step research plan
- Detects podcast generation intent from keywords

**Output**:
```json
{
  "plan": ["Step 1: ...", "Step 2: ...", "Step 3: ..."],
  "generate_podcast": true/false
}
```

**Example Plan:**
For query "What are the latest advances in quantum computing?":
1. 搜索量子计算的最新研究进展
2. 调查主要的量子计算公司和项目
3. 分析量子计算的实际应用案例
4. 总结未来发展趋势和挑战

### 2. Researcher Agent

**Input**: Research plan
**Process**:
- Executes each step sequentially
- Uses LLM to gather detailed information
- Collects comprehensive findings

**Output**: Array of research results for each step

### 3. Reporter Agent

**Input**: All research results
**Process**:
- Synthesizes findings into coherent report
- Formats content in HTML with proper structure
- Adds styling for professional appearance
- Optionally includes image placeholders

**Output**: Complete HTML report

### 4. Podcast Agent (Optional)

**Input**: Research results and final report
**Process**:
- Converts technical content to conversational format
- Creates engaging dialogue or monologue
- Maintains informational accuracy

**Output**: Podcast script in conversational style

## Web Interface Features

### Real-Time Progress Updates

The web interface provides live updates during research:
- Initial planning phase
- Each research step execution
- Report generation
- Podcast script creation

### Research History

- Automatically saves all research sessions
- Browse previous queries by timestamp
- Instant replay of cached results
- No redundant API calls for repeated queries

### Caching System

DeerFlow intelligently caches research results:
- Each unique query is saved in `data/[sanitized-query]/`
- Subsequent requests for the same query use cached data
- Fast replay with simulated progress for better UX

## API Endpoints

### POST /api/run

Execute a research query.

**Query Parameters:**
- `query` (required): The research question

**Response**: Server-Sent Events stream

**Event Types:**
- `update`: Progress updates
- `log`: Research process logs
- `result`: Final report and podcast script
- `error`: Error messages

**Example:**
```javascript
const eventSource = new EventSource('/api/run?query=Your+question');
eventSource.onmessage = (event) => {
  const data = JSON.parse(event.data);
  // Handle different event types
};
```

### GET /api/history

Retrieve research history.

**Response:**
```json
[
  {
    "query": "Research question",
    "timestamp": "2024-12-06T10:30:00Z",
    "dir_name": "Research_question"
  }
]
```

## Advanced Usage

### Custom LLM Models

Modify `nodes.go` to use different models:

```go
func getLLM() (llms.Model, error) {
    return openai.New(
        openai.WithModel("gpt-4"),  // Change model here
    )
}
```

### Extending Agents

Add new agent nodes by:

1. Define node function in `nodes.go`:
```go
func MyCustomNode(ctx context.Context, state any) (any, error) {
    s := state.(*State)
    // Your logic here
    return s, nil
}
```

2. Register node in `graph.go`:
```go
workflow.AddNode("custom", "Custom node description", MyCustomNode)
workflow.AddEdge("previous_node", "custom")
```

### Production Deployment

For production, use the included nginx.conf:

```bash
# Copy nginx config
sudo cp nginx.conf /etc/nginx/sites-available/deerflow
sudo ln -s /etc/nginx/sites-available/deerflow /etc/nginx/sites-enabled/

# Start DeerFlow
./deerflow &

# Restart nginx
sudo systemctl restart nginx
```

## Troubleshooting

### API Key Not Set

```
Please set OPENAI_API_KEY environment variable
```

**Solution**:
```bash
export OPENAI_API_KEY="sk-..."
```

### Connection Refused

If web interface doesn't load:
- Check if port 8085 is available
- Verify the application is running
- Check firewall settings

### Empty or Incomplete Reports

If reports are inadequate:
- Verify API key is valid and has credits
- Check API base URL if using non-OpenAI provider
- Try with more specific queries
- Check network connectivity

### JSON Parsing Errors

The system includes fallback parsing:
- If LLM returns malformed JSON, it uses simple text parsing
- Check logs for parsing issues
- Consider using more capable models (GPT-4 vs GPT-3.5)

## Performance Considerations

### Response Times

- **Planning**: 2-5 seconds
- **Research**: 5-15 seconds (depends on plan steps)
- **Reporting**: 5-10 seconds
- **Podcast**: 5-10 seconds (if enabled)
- **Total**: Typically 15-40 seconds

### Cost Optimization

- Use cheaper models (gpt-3.5-turbo) for research steps
- Use premium models (gpt-4) for final report
- Cache results to avoid repeated API calls
- Limit research plan steps for simpler queries

### Caching Benefits

- **Zero cost** for repeated queries
- **Instant results** (200ms per log replay)
- **Consistent output** for same questions
- **Bandwidth savings** for users

## Future Enhancements

Planned features:
- [ ] Real web search integration (Tavily, Google, Bing)
- [ ] Multi-language support
- [ ] PDF export
- [ ] Audio generation from podcast scripts
- [ ] Collaborative research sessions
- [ ] Custom report templates
- [ ] Image search and inclusion
- [ ] Source citations with links
- [ ] Export to various formats (Markdown, Word, etc.)

## Comparison with ByteDance DeerFlow

| Feature | ByteDance DeerFlow (Python) | This Implementation (Go) |
|---------|----------------------------|-------------------------|
| Multi-agent architecture | ✅ | ✅ |
| Research planning | ✅ | ✅ |
| Web search | ✅ | ⚠️ LLM-based (planned) |
| Report generation | ✅ | ✅ |
| Web interface | ✅ | ✅ |
| CLI support | ✅ | ✅ |
| Podcast generation | ❌ | ✅ |
| Result caching | ⚠️ | ✅ |
| SSE real-time updates | ⚠️ | ✅ |
| History browsing | ⚠️ | ✅ |
| Language | Python | Go |

## License

MIT License - Same as the parent langgraphgo project

## References

- [ByteDance DeerFlow](https://github.com/bytedance/deer-flow) - Original Python implementation
- [LangGraph Go](https://github.com/smallnest/langgraphgo) - Graph-based agent framework
- [LangChain Go](https://github.com/tmc/langchaingo) - LLM integration library

## Contributing

Contributions are welcome! Areas for improvement:
- Real web search integration
- Enhanced UI/UX
- Additional export formats
- Performance optimizations
- Test coverage
- Documentation improvements

## Support

For issues and questions:
- Check the troubleshooting section
- Review the examples in this README
- Open an issue on the langgraphgo GitHub repository

---

**Built with**:
- [langgraphgo](https://github.com/smallnest/langgraphgo) - Graph-based agent orchestration
- [langchaingo](https://github.com/tmc/langchaingo) - LLM integration
- Server-Sent Events for real-time updates
- Embedded Go web server for simplicity
