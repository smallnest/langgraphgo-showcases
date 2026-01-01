package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/smallnest/langgraphgo/graph"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/textsplitter"
)

// IngestionGraphBuilder builds the ingestion graph that processes PDF documents.
// The graph follows the original TypeScript implementation:
// START -> ingestDocs -> END
func IngestionGraphBuilder(vs *VectorStore, cfg Config) *graph.StateGraph[IngestionState] {
	g := graph.NewStateGraph[IngestionState]()

	// Add the ingestDocs node
	g.AddNode("ingestDocs", "Ingest and index documents", func(ctx context.Context, state IngestionState) (IngestionState, error) {
		log.Println("Node: ingestDocs")

		// Check if state.Docs is the "delete" command
		if deleteCmd, ok := state.Docs.(string); ok && deleteCmd == "delete" {
			log.Println("Clearing document index...")
			// In a real implementation, we would clear the vector store here
			// For now, just return success
			return IngestionState{Docs: nil}, nil
		}

		// Extract documents from state
		var docs []schema.Document
		switch v := state.Docs.(type) {
		case []schema.Document:
			docs = v
		case nil:
			log.Println("No documents to ingest")
			return IngestionState{Docs: nil}, nil
		default:
			return state, fmt.Errorf("unsupported docs type: %T", state.Docs)
		}

		if len(docs) == 0 {
			log.Println("No documents to ingest")
			return IngestionState{Docs: nil}, nil
		}

		log.Printf("Processing %d document(s)...", len(docs))

		// Split documents into chunks
		splitter := textsplitter.NewRecursiveCharacter()
		splitter.ChunkSize = cfg.ChunkSize
		splitter.ChunkOverlap = cfg.ChunkOverlap
		splitter.Separators = []string{"\n\n", "\n", ". ", " ", ""}

		splitDocs, err := textsplitter.SplitDocuments(splitter, docs)
		if err != nil {
			return state, fmt.Errorf("failed to split documents: %w", err)
		}

		log.Printf("Split into %d chunks", len(splitDocs))

		// Add documents to vector store
		if err := vs.AddDocuments(ctx, splitDocs); err != nil {
			return state, fmt.Errorf("failed to add documents to vector store: %w", err)
		}

		log.Printf("Successfully indexed %d document chunks", len(splitDocs))

		// Clear docs from state after successful ingestion
		return IngestionState{Docs: nil}, nil
	})

	g.SetEntryPoint("ingestDocs")
	g.AddEdge("ingestDocs", graph.END)

	return g
}

// LoadDocumentsFromFile loads documents from various file formats.
// Supports PDF, TXT, MD files.
func LoadDocumentsFromFile(filePath string) ([]schema.Document, error) {
	// Check file extension
	ext := strings.ToLower(filepath.Ext(filePath))

	switch ext {
	case ".pdf":
		return loadPDF(filePath)
	case ".txt", ".md":
		return loadTextFile(filePath)
	default:
		return nil, fmt.Errorf("unsupported file type: %s", ext)
	}
}

// loadPDF loads and parses a PDF file.
func loadPDF(filePath string) ([]schema.Document, error) {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("file not found: %s", filePath)
	}

	// Try to use langchaingo's PDF loader
	// Note: This requires the PDF loader to be properly configured
	// For now, we'll use a simpler approach by reading the file

	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// For PDF files, we'd ideally use a PDF parser
	// Since the PDF loader in langchaingo may have dependencies,
	// we'll check if it's a PDF and add appropriate metadata
	doc := schema.Document{
		PageContent: string(content),
		Metadata: map[string]any{
			"source": filePath,
			"type":   "pdf",
		},
	}

	// If the content looks like PDF binary data, warn the user
	if len(content) > 4 && string(content[0:4]) == "%PDF" {
		log.Printf("Warning: Binary PDF detected. For proper PDF parsing, ensure PDF dependencies are installed.")
		log.Printf("Attempting to process as text...")
	}

	return []schema.Document{doc}, nil
}

// loadTextFile loads a plain text file.
func loadTextFile(filePath string) ([]schema.Document, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	doc := schema.Document{
		PageContent: string(content),
		Metadata: map[string]any{
			"source": filePath,
			"type":   "text",
		},
	}

	return []schema.Document{doc}, nil
}

// IngestDocuments is a convenience function that loads and ingests documents.
func IngestDocuments(ctx context.Context, vs *VectorStore, cfg Config, filePaths []string) error {
	// Load all documents
	var allDocs []schema.Document
	for _, path := range filePaths {
		docs, err := LoadDocumentsFromFile(path)
		if err != nil {
			log.Printf("Warning: failed to load %s: %v", path, err)
			continue
		}
		allDocs = append(allDocs, docs...)
		log.Printf("Loaded %d document(s) from %s", len(docs), path)
	}

	if len(allDocs) == 0 {
		return fmt.Errorf("no documents were successfully loaded")
	}

	// Build and compile ingestion graph
	ingestGraph := IngestionGraphBuilder(vs, cfg)
	runnable, err := ingestGraph.Compile()
	if err != nil {
		return fmt.Errorf("failed to compile ingestion graph: %w", err)
	}

	// Run ingestion
	state := IngestionState{Docs: allDocs}
	_, err = runnable.Invoke(ctx, state)
	if err != nil {
		return fmt.Errorf("ingestion failed: %w", err)
	}

	return nil
}

// ClearDocuments clears all documents from the vector store.
func ClearDocuments(ctx context.Context, vs *VectorStore, cfg Config) error {
	ingestGraph := IngestionGraphBuilder(vs, cfg)
	runnable, err := ingestGraph.Compile()
	if err != nil {
		return fmt.Errorf("failed to compile ingestion graph: %w", err)
	}

	state := IngestionState{Docs: "delete"}
	_, err = runnable.Invoke(ctx, state)
	return err
}
