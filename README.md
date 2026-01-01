# LangGraph Go Showcases

A collection of AI agent showcases and demonstrations built with [LangGraph Go](https://github.com/smallnest/langgraphgo). This repository was extracted from the showcases directory of the main LangGraph Go project to provide standalone, production-ready examples of AI agent implementations.

## Overview

This repository contains 9 comprehensive AI agent implementations showcasing different use cases and patterns for building intelligent systems with LangGraph Go and LangChain Go. Each showcase is a complete, runnable project with its own README, examples, and documentation.

## Table of Contents

- [Showcases](#showcases)
  - [BettaFish - Public Opinion Analysis](#bettafish)
  - [GPT Researcher - Autonomous Research Agent](#gpt-researcher)
  - [Health Insights Agent - Medical Report Analyzer](#health-insights-agent)
  - [PeopleHub - Person Research Agent](#peoplehub)
  - [DeepAgents - Filesystem-Aware AI Agent](#deepagents)
  - [DeerFlow - Deep Research Agent](#deerflow)
  - [LangManus - Multi-Agent Automation Framework](#langmanus)
  - [Profile - Digital Footprint Analyzer](#profile)
  - [AI PDF Chatbot - Document Q&A System](#ai-pdf-chatbot)
- [Getting Started](#getting-started)
- [Requirements](#requirements)
- [License](#license)

## Showcases

### BettaFish

**Deep Public Opinion Analysis with Multi-Agent Architecture**

A complete Go implementation of the [BettaFish](https://github.com/666ghj/BettaFish) project, featuring:
- QueryEngine with reflection loop for iterative search refinement
- MediaEngine for relevant image search
- ForumEngine with LLM-driven discussion between agents
- ReportEngine for comprehensive Markdown reports

[📂 View BettaFish →](./BettaFish)

**Key Technologies**: Tavily API, Multi-agent discussion, Reflection patterns

---

### GPT Researcher

**Autonomous Research Agent for Comprehensive Reports**

A Go port of [gpt-researcher](https://github.com/assafelovic/gpt-researcher) that automates research tasks:
- Generates focused research questions from queries
- Gathers information from 20+ web sources
- Produces detailed 2000+ word reports with citations
- Supports multiple report types (research, outline, resource)

[📂 View GPT Researcher →](./gpt_researcher)

**Key Technologies**: Tavily search, Web scraping, Multi-source aggregation

---

### Health Insights Agent

**AI-Powered Blood Report Analysis**

Intelligent health analysis system that processes blood test reports:
- Extracts parameters from text and PDF reports
- Identifies potential health risks with severity levels
- Provides personalized recommendations (diet, lifestyle, medical)
- Generates structured analysis with confidence scores

[📂 View Health Insights Agent →](./health_insights_agent)

**Key Technologies**: LLM-based extraction, Medical knowledge, PDF processing

---

### PeopleHub

**Automated Person Research Agent**

A Go implementation of [PeopleHub](https://github.com/MeirKaD/pepolehub) for researching individuals:
- Fetches LinkedIn profiles via web search
- Performs comprehensive Google searches
- Scrapes and summarizes relevant web pages
- Generates detailed person research reports

[📂 View PeopleHub →](./pepolehub)

**Key Technologies**: Parallel execution, Web scraping, Profile aggregation

---

### DeepAgents

**Filesystem-Aware AI Agent with Task Management**

A powerful agent framework with filesystem access:
- Read, write, and manage files within workspace
- Built-in todo list management
- SubAgent delegation for complex tasks
- Pattern-based file finding with glob support

[📂 View DeepAgents →](./deepagents)

**Key Technologies**: Filesystem tools, Task delegation, Hierarchical agents

---

### DeerFlow

**Deep Research Agent with Web Interface**

Go implementation of [ByteDance DeerFlow](https://github.com/bytedance/deer-flow):
- Multi-agent architecture (Planner → Researcher → Reporter → Podcast)
- Real-time progress updates via Server-Sent Events
- Modern dark theme web interface
- Research history and result caching
- Optional podcast script generation

[📂 View DeerFlow →](./deerflow)

**Key Technologies**: SSE streaming, Web UI, Podcast generation, Caching

---

### LangManus

**Multi-Agent AI Automation Framework**

A Go implementation of the [LangManus](https://github.com/Darwin-lfl/langmanus) framework:
- Layered architecture with 7 specialized agents
- Coordinator → Planner → Supervisor → Workers → Reporter
- Code execution (Python/Bash) with safety controls
- Web search integration via Tavily
- Streaming support for real-time updates

[📂 View LangManus →](./langmanus)

**Key Technologies**: Multi-agent orchestration, Code execution, Task planning

---

### Profile

**Digital Footprint Analyzer**

An OSINT tool for analyzing digital presence across the internet:
- Searches 1000+ social media platforms
- AI-powered psychological profiling
- Real-time streaming feedback via SSE
- Beautiful paper-textured UI design
- Privacy-focused (public data only)

[📂 View Profile →](./profile)

**Key Technologies**: OSINT, Social media search, Psychological analysis

**Live Demo**: http://profile.rpcx.io

---

### AI PDF Chatbot

**Intelligent Document Q&A System**

A full-stack RAG (Retrieval-Augmented Generation) application:
- Upload and ingest PDF documents
- Vector-based semantic search
- Conversational Q&A interface
- Backend: Go with LangChain Go
- Frontend: Modern web interface

[📂 View AI PDF Chatbot →](./ai-pdf-chatbot)

**Key Technologies**: RAG, Vector database, PDF processing, Chat interface

---

## Getting Started

Each showcase is self-contained with its own dependencies and setup instructions.

### General Prerequisites

- **Go**: Version 1.21 or higher
- **API Keys**: Most showcases require:
  - OpenAI API key (or compatible provider)
  - Tavily API key (for web search features)

### Quick Start

1. Choose a showcase from the list above
2. Navigate to its directory:
   ```bash
   cd <showcase-name>
   ```
3. Follow the README in that directory for specific setup instructions
4. Set required environment variables:
   ```bash
   export OPENAI_API_KEY="your-key-here"
   export TAVILY_API_KEY="your-tavily-key"  # if needed
   ```
5. Run the showcase:
   ```bash
   go run *.go
   # or
   go build && ./<showcase-name>
   ```

## Requirements

### Common Dependencies

All showcases use:
- [LangGraph Go](https://github.com/smallnest/langgraphgo) - Graph-based agent orchestration
- [LangChain Go](https://github.com/tmc/langchaingo) - LLM integration library

### API Services

Most showcases integrate with:
- **OpenAI API** - For LLM capabilities (GPT-4, GPT-3.5, etc.)
- **Tavily API** - For web search functionality
- **Alternative Providers** - Many support OpenAI-compatible APIs (DeepSeek, Azure, Ollama, etc.)

### Environment Configuration

Common environment variables:
```bash
# Required
OPENAI_API_KEY="sk-..."

# Optional
OPENAI_API_BASE="https://api.openai.com/v1"  # Custom endpoint
OPENAI_MODEL="gpt-4o"                        # Model selection
TAVILY_API_KEY="tvly-..."                    # For search features
```

## Project Structure

```
langgraphgo-showcases/
├── BettaFish/              # Public opinion analysis
├── gpt_researcher/         # Autonomous research agent
├── health_insights_agent/  # Medical report analyzer
├── pepolehub/             # Person research agent
├── deepagents/            # Filesystem-aware agent
├── deerflow/              # Deep research with web UI
├── langmanus/             # Multi-agent framework
├── profile/               # Digital footprint analyzer
├── ai-pdf-chatbot/        # PDF Q&A system
├── LICENSE                # MIT License
└── README.md              # This file
```

## Features Across Showcases

### Agent Patterns
- ✅ Multi-agent orchestration
- ✅ Hierarchical task delegation
- ✅ Reflection and self-improvement loops
- ✅ Parallel execution
- ✅ State management with LangGraph

### Capabilities
- ✅ Web search and scraping
- ✅ Document processing (PDF, text)
- ✅ Code generation and execution
- ✅ Report generation (Markdown, HTML)
- ✅ Real-time streaming
- ✅ Filesystem operations
- ✅ Task planning and decomposition

### Interfaces
- ✅ Command-line interfaces (CLI)
- ✅ Web interfaces with modern UI
- ✅ Programmatic APIs
- ✅ Streaming support (SSE)

## Use Cases

These showcases demonstrate solutions for:

- 📊 **Research & Analysis**: Automated research, data gathering, report generation
- 🏥 **Healthcare**: Medical report analysis, health insights
- 🔍 **OSINT**: Public data aggregation, digital footprint analysis
- 📝 **Document Processing**: PDF analysis, Q&A systems
- 💼 **Automation**: Task orchestration, code execution, workflow automation
- 🌐 **Web Intelligence**: Search, scraping, content synthesis

## Learning Path

**Beginner**: Start with simpler showcases
1. PeopleHub - Basic multi-step workflow
2. GPT Researcher - Classic research pattern

**Intermediate**: Explore more complex patterns
3. Health Insights Agent - Structured extraction
4. DeepAgents - Filesystem and tools
5. AI PDF Chatbot - RAG implementation

**Advanced**: Multi-agent systems
6. BettaFish - Reflection loops
7. DeerFlow - Full-stack with UI
8. LangManus - Complex orchestration

## Contributing

Contributions are welcome! Whether it's:
- Bug fixes
- New showcases
- Documentation improvements
- Performance optimizations
- Test coverage

Please feel free to open issues or submit pull requests.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

These showcases are inspired by and port several excellent open-source projects:
- [BettaFish](https://github.com/666ghj/BettaFish) - Original Python implementation
- [gpt-researcher](https://github.com/assafelovic/gpt-researcher) - Research automation
- [PeopleHub](https://github.com/MeirKaD/pepolehub) - Person research
- [ByteDance DeerFlow](https://github.com/bytedance/deer-flow) - Deep research
- [LangManus](https://github.com/Darwin-lfl/langmanus) - Multi-agent framework
- [hia](https://github.com/harshhh28/hia) - Health insights

Special thanks to:
- [LangGraph Go](https://github.com/smallnest/langgraphgo) - Framework foundation
- [LangChain Go](https://github.com/tmc/langchaingo) - LLM integration
- [Tavily](https://www.tavily.com/) - Web search API
- [OpenAI](https://openai.com/) - Language models

## Related Resources

- [LangGraph Go Documentation](https://github.com/smallnest/langgraphgo)
- [LangChain Go Documentation](https://github.com/tmc/langchaingo)
- [LangGraph (Python) Docs](https://python.langchain.com/docs/langgraph)

## Support

For questions, issues, or discussions:
- Open an issue in this repository
- Check individual showcase READMEs
- Visit the main [LangGraph Go repository](https://github.com/smallnest/langgraphgo)

---

**Built with ❤️ using LangGraph Go and LangChain Go**
