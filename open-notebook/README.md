# Open Notebook - LangGraphGo Edition

<div align="center">

**A privacy-first, open-source alternative to NotebookLM**

[![Go](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/license-MIT-blue)](./LICENSE)

An AI-powered knowledge management application built with [LangGraphGo](https://github.com/smallnest/langgraphgo) that lets you create intelligent notebooks from your documents.

</div>

## ✨ Features

- 📚 **Multiple Source Types** - Upload PDFs, text files, Markdown, DOCX, and HTML documents
- 🤖 **AI-Powered Chat** - Ask questions and get answers based on your sources
- ✨ **Multiple Transformations** - Generate summaries, FAQs, study guides, outlines, timelines, glossaries, quizzes, and podcast scripts
- 🎙️ **Podcast Generation** - Create engaging podcast scripts from your content
- 💾 **Full Privacy** - Local SQLite storage, optional cloud backends
- 🔄 **Multi-Model Support** - Works with OpenAI, Ollama, and other compatible APIs
- 🎨 **Academic Brutalist Design** - Distinctive, research-focused interface

## 🚀 Quick Start

### Prerequisites

- Go 1.23 or later
- An LLM API key (OpenAI) or Ollama running locally

### Installation

```bash
# Navigate to the showcase directory
cd langgraphgo/showcases/open-notebook

# Install dependencies
go mod tidy

# Set your API key
export OPENAI_API_KEY=your_key_here

# Run the server
go run . -server
```

Open your browser to `http://localhost:8080`

### Using Ollama (Local, Free)

```bash
# Make sure Ollama is running
ollama serve

# Run with Ollama
export OLLAMA_BASE_URL=http://localhost:11434
go run . -server
```

### Alternative: Build and Run

```bash
# Build the binary
go build -o open-notebook .

# Run with OpenAI
export OPENAI_API_KEY=your_key_here
./open-notebook -server

# Or run with Ollama
export OLLAMA_BASE_URL=http://localhost:11434
./open-notebook -server
```

## 📖 Usage

### Creating Notebooks

1. Click "New Notebook" in the header
2. Enter a name and optional description
3. Click "Create Notebook"

### Adding Sources

You can add content to your notebook in three ways:

**File Upload**
- Click the "+" button in the Sources panel
- Drag and drop or browse for files
- Supported: PDF, TXT, MD, DOCX, HTML

**Paste Text**
- Select the "Text" tab
- Enter a title and paste your content

**From URL**
- Select the "URL" tab
- Enter the URL and optional title

### Chatting with Sources

1. Switch to the "CHAT" tab
2. Ask questions about your content
3. Responses include references to relevant sources

### Transformations

Click any transformation card to generate:

| Transformation | Description |
|---------------|-------------|
| 📝 Summary | Condensed overview of your sources |
| ❓ FAQ | Common questions and answers |
| 📚 Study Guide | Educational material with learning objectives |
| 🗂️ Outline | Hierarchical structure of topics |
| 🎙️ Podcast | Conversational script for audio content |
| 📅 Timeline | Chronological events from sources |
| 📖 Glossary | Key terms and definitions |
| ✍️ Quiz | Assessment questions with answer key |

Or use the custom prompt field for any other transformation.

## ⚙️ Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `OPENAI_API_KEY` | OpenAI API key | Required (unless using Ollama) |
| `OPENAI_BASE_URL` | Custom API base URL | OpenAI default |
| `OPENAI_MODEL` | Model name | `gpt-4o-mini` |
| `EMBEDDING_MODEL` | Embedding model | `text-embedding-3-small` |
| `OLLAMA_BASE_URL` | Ollama server URL | `http://localhost:11434` |
| `OLLAMA_MODEL` | Ollama model name | `llama3.2` |
| `SERVER_HOST` | Server host | `0.0.0.0` |
| `SERVER_PORT` | Server port | `8080` |
| `VECTOR_STORE_TYPE` | Vector store backend | `sqlite` |
| `STORE_PATH` | Database path | `./data/checkpoints.db` |
| `MAX_SOURCES` | Max sources for RAG | `5` |
| `CHUNK_SIZE` | Document chunk size | `1000` |
| `CHUNK_OVERLAP` | Chunk overlap | `200` |

### Vector Store Options

- `sqlite` - Local SQLite database (default)
- `memory` - In-memory vectors
- `supabase` - Supabase vector store
- `postgres` / `pgvector` - PostgreSQL with pgvector
- `redis` - Redis with RediSearch

### Example Configuration Files

**docker-compose.yml** (for PostgreSQL + pgvector)

```yaml
version: '3.8'
services:
  postgres:
    image: pgvector/pgvector:pg16
    environment:
      POSTGRES_DB: notebook
      POSTGRES_USER: notebook
      POSTGRES_PASSWORD: secret
    ports:
      - "5432:5432"

  app:
    build: .
    environment:
      - POSTGRES_URL=postgres://notebook:secret@postgres:5432/notebook
      - VECTOR_STORE_TYPE=postgres
    ports:
      - "8080:8080"
```

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        Frontend                              │
│              (HTML/CSS/JS - Brutalist UI)                    │
└────────────────────┬────────────────────────────────────────┘
                     │ HTTP API
┌────────────────────▼────────────────────────────────────────┐
│                      Server Layer                            │
│                    (Gin Router)                              │
└────────────────────┬────────────────────────────────────────┘
                     │
        ┌────────────┼────────────┐
        │            │            │
┌───────▼──────┐ ┌──▼──────┐ ┌──▼──────────┐
│  VectorStore │ │  Store  │ │   Agent     │
│              │ │         │ │             │
│ - Embeddings │ │ SQLite  │ │ - LLM calls │
│ - Search     │ │         │ │ - Prompts   │
│ - Chunks     │ │         │ │ - RAG       │
└──────────────┘ └─────────┘ └─────────────┘
```

## 🎨 Design Philosophy

This application uses an **Academic Brutalist** aesthetic:

- **Warm paper tones** with sharp black ink - like archival documents
- **Monospace technical fonts** paired with elegant serif headings
- **Visible grid structure** - showing the "structure of knowledge"
- **High contrast typography** for readability and focus
- **Subtle grain texture** for warmth and depth

The design emphasizes function over form, making the content the hero while maintaining a distinctive, memorable character.

## 📁 Project Structure

```
open-notebook/
├── backend/
│   ├── main.go          # CLI entry point
│   ├── config.go        # Configuration management
│   ├── types.go         # Data structures
│   ├── store.go         # Database persistence
│   ├── vector.go        # Vector search
│   ├── agent.go         # AI operations
│   └── server.go        # HTTP server
├── frontend/
│   ├── index.html       # Main HTML
│   └── static/
│       ├── style.css    # Brutalist styles
│       └── app.js       # Application logic
├── agent/               # LangGraph agents (optional)
├── store/               # Checkpoint stores (optional)
├── go.mod
└── README.md
```

## 🔧 Development

### Running Tests

```bash
go test -v ./...
```

### Building

```bash
go build -o open-notebook .
```

### Code Quality

```bash
# Format
go fmt ./...

# Lint
golangci-lint run

# Vet
go vet ./...
```

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 📄 License

MIT License - see [LICENSE](../../LICENSE) for details.

## 🙏 Acknowledgments

- Inspired by [Google's NotebookLM](https://notebooklm.google.com/)
- Built with [LangGraphGo](https://github.com/smallnest/langgraphgo)
- Powered by [LangChain Go](https://github.com/tmc/langchaingo)

## 📞 Support

- Report issues on [GitHub](https://github.com/smallnest/langgraphgo/issues)
- Join discussions in the [LangGraphGo community](https://github.com/smallnest/langgraphgo/discussions)

---

Made with ❤️ using [LangGraphGo](https://github.com/smallnest/langgraphgo)
