package main

import (
	"context"
	"fmt"
	"sync"

	"github.com/smallnest/langgraphgo/rag"
	"github.com/smallnest/langgraphgo/rag/store"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/schema"
)

// VectorStore wraps the RAG vector store for use in the PDF chatbot.
// It provides methods compatible with the original TypeScript implementation.
type VectorStore struct {
	store    rag.VectorStore
	embedder rag.Embedder
	mu       sync.RWMutex
}

// NewVectorStore creates a new vector store based on the configuration.
func NewVectorStore(cfg Config) (*VectorStore, error) {
	switch cfg.VectorStoreType {
	case "memory", "":
		return newInMemoryVectorStore(cfg)
	case "supabase":
		return newSupabaseVectorStore(cfg)
	default:
		return nil, fmt.Errorf("unsupported vector store type: %s", cfg.VectorStoreType)
	}
}

// newInMemoryVectorStore creates an in-memory vector store with OpenAI embeddings.
func newInMemoryVectorStore(cfg Config) (*VectorStore, error) {
	// Initialize OpenAI client for embeddings
	opts := []openai.Option{
		openai.WithEmbeddingModel("embedding-v1"),
		openai.WithBaseURL("https://qianfan.baidubce.com/v2"),
	}
	if cfg.OpenAIBaseURL != "" {
		opts = append(opts, openai.WithBaseURL(cfg.OpenAIBaseURL))
	}

	llm, err := openai.New(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize OpenAI: %w", err)
	}

	// Create embedder using langchaingo
	lcEmbedder, err := embeddings.NewEmbedder(llm)
	if err != nil {
		return nil, fmt.Errorf("failed to create embedder: %w", err)
	}

	// Wrap with our RAG adapter
	embedder := rag.NewLangChainEmbedder(lcEmbedder)

	// Create in-memory vector store
	vs := store.NewInMemoryVectorStore(embedder)

	return &VectorStore{
		store:    vs,
		embedder: embedder,
	}, nil
}

// newSupabaseVectorStore creates a Supabase-backed vector store.
// Note: This is a placeholder for future Supabase integration.
func newSupabaseVectorStore(cfg Config) (*VectorStore, error) {
	// TODO: Implement Supabase vector store integration
	// For now, fall back to in-memory with a warning
	fmt.Println("Warning: Supabase vector store not yet implemented, using in-memory store")
	return newInMemoryVectorStore(cfg)
}

// AddDocuments adds documents to the vector store with embeddings.
// This mimics the original TypeScript implementation's behavior.
func (vs *VectorStore) AddDocuments(ctx context.Context, docs []schema.Document) error {
	vs.mu.Lock()
	defer vs.mu.Unlock()

	// Convert schema.Document to rag.Document
	ragDocs := make([]rag.Document, len(docs))
	for i, doc := range docs {
		ragDocs[i] = rag.Document{
			Content:  doc.PageContent,
			Metadata: doc.Metadata,
		}

		// Generate ID from metadata or use content hash
		if source, ok := doc.Metadata["source"].(string); ok {
			ragDocs[i].ID = source
		} else {
			ragDocs[i].ID = fmt.Sprintf("doc_%d", i)
		}
	}

	return vs.store.Add(ctx, ragDocs)
}

// SimilaritySearch performs a similarity search for the given query.
// Returns the top-k most relevant documents.
func (vs *VectorStore) SimilaritySearch(ctx context.Context, query string, k int) ([]schema.Document, error) {
	vs.mu.RLock()
	defer vs.mu.RUnlock()

	// Embed the query using our stored embedder
	queryEmb, err := vs.embedder.EmbedDocument(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to embed query: %w", err)
	}

	// Search
	results, err := vs.store.Search(ctx, queryEmb, k)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	// Convert results back to schema.Document
	docs := make([]schema.Document, len(results))
	for i, result := range results {
		docs[i] = schema.Document{
			PageContent: result.Document.Content,
			Metadata:    result.Document.Metadata,
		}
	}

	return docs, nil
}

// SearchWithScores performs similarity search and returns documents with scores.
func (vs *VectorStore) SearchWithScores(ctx context.Context, query string, k int) ([]rag.DocumentSearchResult, error) {
	vs.mu.RLock()
	defer vs.mu.RUnlock()

	// Embed the query
	queryEmb, err := vs.embedder.EmbedDocument(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to embed query: %w", err)
	}

	// Search
	return vs.store.Search(ctx, queryEmb, k)
}

// Delete removes documents from the vector store.
func (vs *VectorStore) Delete(ctx context.Context, ids []string) error {
	vs.mu.Lock()
	defer vs.mu.Unlock()
	return vs.store.Delete(ctx, ids)
}

// GetStats returns statistics about the vector store.
func (vs *VectorStore) GetStats(ctx context.Context) (*rag.VectorStoreStats, error) {
	vs.mu.RLock()
	defer vs.mu.RUnlock()
	return vs.store.GetStats(ctx)
}
