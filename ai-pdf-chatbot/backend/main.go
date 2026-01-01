package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/tmc/langchaingo/llms"
)

func main() {
	// Command line flags
	serverMode := flag.Bool("server", false, "Run in HTTP server mode")
	ingestFile := flag.String("ingest", "", "Path to a PDF/Text file to ingest")
	chatMode := flag.Bool("chat", false, "Start interactive chat mode")
	port := flag.String("port", "", "Server port (overrides SERVER_PORT env var)")
	flag.Parse()

	// Load and validate configuration
	cfg := LoadConfig()
	if *port != "" {
		cfg.ServerPort = *port
	}
	ValidateConfig(cfg)

	ctx := context.Background()

	// Initialize vector store
	vs, err := NewVectorStore(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize vector store: %v", err)
	}

	// Handle different modes
	switch {
	case *serverMode:
		// Server mode
		runServerMode(cfg)

	default:
		// Handle ingest if specified
		if *ingestFile != "" {
			fmt.Printf("📂 Ingesting file: %s...\n", *ingestFile)
			if err := IngestDocuments(ctx, vs, cfg, []string{*ingestFile}); err != nil {
				log.Fatalf("Ingestion failed: %v", err)
			}
			fmt.Println("✅ Ingestion complete!")

			// Show stats
			stats, _ := vs.GetStats(ctx)
			fmt.Printf("📊 Total documents in store: %d\n", stats.TotalDocuments)
		}

		// Handle chat mode
		if *chatMode {
			// Interactive chat mode
			runChatMode(ctx, vs, cfg)
		} else if *ingestFile == "" {
			// No mode specified, show usage
			printUsage()
		}
	}
}

// runServerMode starts the HTTP server.
func runServerMode(cfg Config) {
	server, err := NewServer(cfg)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	fmt.Printf("🚀 Starting AI PDF Chatbot server...\n")
	fmt.Printf("📍 Server will be available at: http://%s:%s\n", cfg.ServerHost, cfg.ServerPort)
	fmt.Println("💡 Upload documents via the web interface or API, then start chatting!")

	if err := server.Start(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

// runChatMode runs the interactive CLI chat mode.
func runChatMode(ctx context.Context, vs *VectorStore, cfg Config) {
	fmt.Println("\n🤖 AI PDF Chatbot - Interactive Mode")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("Commands:")
	fmt.Println("  Type your question to chat")
	fmt.Println("  'exit' or 'quit' to exit")
	fmt.Println("  'stats' to show document statistics")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Show initial stats
	stats, _ := vs.GetStats(ctx)
	fmt.Printf("📊 Documents loaded: %d\n\n", stats.TotalDocuments)

	if stats.TotalDocuments == 0 {
		fmt.Println("⚠️  Warning: No documents in store. Use --ingest to add documents first.")
		fmt.Println("   Example: go run backend/*.go --ingest your-document.pdf --chat\n")
	}

	reader := bufio.NewReader(os.Stdin)
	var history []llms.MessageContent

	for {
		fmt.Print("👤 You: ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		switch strings.ToLower(input) {
		case "exit", "quit":
			fmt.Println("👋 Goodbye!")
			return

		case "stats":
			stats, _ := vs.GetStats(ctx)
			fmt.Printf("📊 Documents: %d | Vectors: %d | Dimension: %d\n",
				stats.TotalDocuments, stats.TotalVectors, stats.Dimension)
			continue

		case "":
			continue
		}

		// Process question
		fmt.Print("\n🤖 Assistant: ")
		response, docs, err := Query(ctx, vs, cfg, input, history)
		if err != nil {
			fmt.Printf("❌ Error: %v\n\n", err)
			continue
		}

		fmt.Println(response)

		// Show sources if documents were retrieved
		if len(docs) > 0 {
			fmt.Printf("\n📚 Sources: %d document(s) referenced\n", len(docs))
			for i, doc := range docs {
				if source, ok := doc.Metadata["source"].(string); ok {
					fmt.Printf("   [%d] %s\n", i+1, source)
				}
			}
		}
		fmt.Println()

		// Update history
		history = append(history,
			llms.TextParts(llms.ChatMessageTypeHuman, input),
			llms.TextParts(llms.ChatMessageTypeAI, response),
		)

		// Limit history size
		if len(history) > 20 {
			history = history[len(history)-20:]
		}
	}
}

// printUsage shows command usage.
func printUsage() {
	fmt.Println("AI PDF Chatbot - Go Port of ai-pdf-chatbot-langchain")
	fmt.Println("\nUsage:")
	fmt.Println("  go run backend/*.go [options]")
	fmt.Println("\nOptions:")
	fmt.Println("  -server        Run in HTTP server mode (provides web UI)")
	fmt.Println("  -ingest <file> Ingest a PDF/Text file into the vector store")
	fmt.Println("  -chat          Start interactive chat mode")
	fmt.Println("  -port <port>   Server port (default: 8080)")
	fmt.Println("\nExamples:")
	fmt.Println("  # Ingest a document")
	fmt.Println("  go run backend/*.go -ingest manual.pdf")
	fmt.Println("\n  # Chat with ingested documents")
	fmt.Println("  go run backend/*.go -chat")
	fmt.Println("\n  # Ingest and chat in one command")
	fmt.Println("  go run backend/*.go -ingest manual.pdf -chat")
	fmt.Println("\n  # Start web server")
	fmt.Println("  go run backend/*.go -server")
	fmt.Println("\nEnvironment Variables:")
	fmt.Println("  OPENAI_API_KEY              Required - Your OpenAI API key")
	fmt.Println("  SERVER_HOST                 Server host (default: 0.0.0.0)")
	fmt.Println("  SERVER_PORT                 Server port (default: 8080)")
	fmt.Println("  VECTOR_STORE_TYPE           Vector store type (default: memory)")
	fmt.Println("  SUPABASE_URL                Supabase URL (if using Supabase)")
	fmt.Println("  SUPABASE_SERVICE_ROLE_KEY   Supabase key (if using Supabase)")
	fmt.Println("  LANGCHAIN_API_KEY           Optional - LangSmith tracing")
	fmt.Println("  LANGCHAIN_PROJECT           Optional - LangSmith project name")
}
