package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/smallnest/langgraphgo/graph"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
)

// Server represents the HTTP server for the AI PDF Chatbot.
type Server struct {
	cfg         Config
	vs          *VectorStore
	ingestGraph *graph.StateRunnable[IngestionState]
	retrieveGraph *graph.StateRunnable[RetrievalState]

	// Session management
	sessions   map[string][]llms.MessageContent
	sessionsMu sync.RWMutex
}

// NewServer creates a new HTTP server.
func NewServer(cfg Config) (*Server, error) {
	// Initialize vector store
	vs, err := NewVectorStore(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create vector store: %w", err)
	}

	// Build graphs
	ingestGraph := IngestionGraphBuilder(vs, cfg)
	ingestRunnable, err := ingestGraph.Compile()
	if err != nil {
		return nil, fmt.Errorf("failed to compile ingestion graph: %w", err)
	}

	retrieveGraph := RetrievalGraphBuilder(vs, cfg)
	retrieveRunnable, err := retrieveGraph.Compile()
	if err != nil {
		return nil, fmt.Errorf("failed to compile retrieval graph: %w", err)
	}

	return &Server{
		cfg:          cfg,
		vs:           vs,
		ingestGraph:  ingestRunnable,
		retrieveGraph: retrieveRunnable,
		sessions:     make(map[string][]llms.MessageContent),
	}, nil
}

// Start starts the HTTP server.
func (s *Server) Start() error {
	addr := s.cfg.ServerHost + ":" + s.cfg.ServerPort
	log.Printf("Starting AI PDF Chatbot server on %s", addr)

	// Set up routes
	http.HandleFunc("/", s.handleIndex)
	http.HandleFunc("/api/ingest", s.handleIngest)
	http.HandleFunc("/api/chat", s.handleChat)
	http.HandleFunc("/api/health", s.handleHealth)

	// Serve static files
	http.HandleFunc("/static/", s.handleStatic)

	return http.ListenAndServe(addr, nil)
}

// handleIndex serves the main HTML page.
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	// Read the HTML file
	htmlPath := filepath.Join("frontend", "index.html")
	http.ServeFile(w, r, htmlPath)
}

// handleStatic serves static files.
func (s *Server) handleStatic(w http.ResponseWriter, r *http.Request) {
	filePath := filepath.Join("frontend", r.URL.Path[1:])
	http.ServeFile(w, r, filePath)
}

// handleHealth returns health status.
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	stats, _ := s.vs.GetStats(r.Context())

	response := map[string]any{
		"status": "ok",
		"documents": stats.TotalDocuments,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleIngest handles document ingestion requests.
func (s *Server) handleIngest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request
	var req IngestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Printf("Ingest request: %d files, action: %s", len(req.Files), req.Action)

	// Handle delete action
	if req.Action == "delete" {
		ctx := r.Context()
		state := IngestionState{Docs: "delete"}
		_, err := s.ingestGraph.Invoke(ctx, state)
		if err != nil {
			log.Printf("Delete failed: %v", err)
			sendJSONError(w, "Failed to clear documents", http.StatusInternalServerError)
			return
		}

		sendJSONResponse(w, IngestResponse{
			Success: true,
			Message: "Documents cleared successfully",
		})
		return
	}

	// Handle file upload
	ctx := r.Context()

	// Load documents
	var allDocs []schema.Document
	for _, file := range req.Files {
		docs, err := LoadDocumentsFromFile(file)
		if err != nil {
			log.Printf("Failed to load %s: %v", file, err)
			continue
		}
		allDocs = append(allDocs, docs...)
	}

	if len(allDocs) == 0 {
		sendJSONError(w, "No documents were successfully loaded", http.StatusBadRequest)
		return
	}

	// Ingest documents
	state := IngestionState{Docs: allDocs}
	_, err := s.ingestGraph.Invoke(ctx, state)
	if err != nil {
		log.Printf("Ingestion failed: %v", err)
		sendJSONError(w, "Ingestion failed", http.StatusInternalServerError)
		return
	}

	// Get updated stats
	stats, _ := s.vs.GetStats(ctx)

	sendJSONResponse(w, IngestResponse{
		Success:           true,
		Message:           "Documents ingested successfully",
		DocumentsIngested: len(allDocs),
	})
	log.Printf("Ingestion complete: %d documents, total in store: %d", len(allDocs), stats.TotalDocuments)
}

// handleChat handles chat requests with SSE streaming.
func (s *Server) handleChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request
	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Printf("Chat request from session %s: %s", req.SessionID, req.Message)

	// Get or create session
	s.sessionsMu.Lock()
	if req.SessionID == "" {
		req.SessionID = generateSessionID()
	}
	messages := s.sessions[req.SessionID]
	s.sessionsMu.Unlock()

	// Set up SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	// Send session ID
	sseEvent(w, flusher, "metadata", map[string]string{
		"sessionId": req.SessionID,
	})

	// Run retrieval graph
	ctx := r.Context()
	state := RetrievalState{
		Question: req.Message,
		Messages: messages,
	}

	finalState, err := s.retrieveGraph.Invoke(ctx, state)
	if err != nil {
		log.Printf("Chat failed: %v", err)
		sseEvent(w, flusher, "error", map[string]string{
			"message": err.Error(),
		})
		sseEvent(w, flusher, "done", nil)
		return
	}

	// Stream response content
	if len(finalState.Messages) > 0 {
		lastMsg := finalState.Messages[len(finalState.Messages)-1]
		if lastMsg.Role == llms.ChatMessageTypeAI && len(lastMsg.Parts) > 0 {
			if textPart, ok := lastMsg.Parts[0].(llms.TextContent); ok {
				content := textPart.Text

				// Stream content word by word (simulated)
				words := strings.Fields(content)
				for i, word := range words {
					if i > 0 {
						sseEvent(w, flusher, "message", MessageEventData{
							Content: " " + word,
							Role:    "assistant",
						})
					} else {
						sseEvent(w, flusher, "message", MessageEventData{
							Content: word,
							Role:    "assistant",
						})
					}
				}
			}
		}
	}

	// Send source documents
	for _, doc := range finalState.Documents {
		sseEvent(w, flusher, "source", SourceEventData{
			Content:  doc.PageContent,
			Metadata: doc.Metadata,
		})
	}

	// Update session
	s.sessionsMu.Lock()
	s.sessions[req.SessionID] = finalState.Messages
	s.sessionsMu.Unlock()

	// Send done event
	sseEvent(w, flusher, "done", nil)
	log.Printf("Chat complete for session %s", req.SessionID)
}

// ============================================
// SSE Helpers
// ============================================

// sseEvent sends a server-sent event.
func sseEvent(w http.ResponseWriter, flusher http.Flusher, event string, data any) {
	var jsonData string
	if data != nil {
		bytes, err := json.Marshal(data)
		if err != nil {
			return
		}
		jsonData = string(bytes)
	} else {
		jsonData = "{}"
	}

	fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, jsonData)
	flusher.Flush()
}

// sendJSONResponse sends a JSON response.
func sendJSONResponse(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// sendJSONError sends a JSON error response.
func sendJSONError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]any{
		"error": message,
	})
}

// generateSessionID generates a unique session ID.
func generateSessionID() string {
	return fmt.Sprintf("session_%d", len("session"))
}

// ============================================
// File Upload Handler
// ============================================

// handleFileUpload handles multipart file uploads.
func (s *Server) handleFileUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form (max 32MB)
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	files := r.MultipartForm.File["files"]
	if len(files) == 0 {
		http.Error(w, "No files uploaded", http.StatusBadRequest)
		return
	}

	log.Printf("Received %d file(s)", len(files))

	// Process uploaded files
	var filePaths []string
	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			log.Printf("Failed to open file %s: %v", fileHeader.Filename, err)
			continue
		}

		// Save to temporary directory
		tempPath := filepath.Join(os.TempDir(), fileHeader.Filename)
		out, err := os.Create(tempPath)
		if err != nil {
			log.Printf("Failed to create temp file: %v", err)
			file.Close()
			continue
		}

		_, err = io.Copy(out, file)
		file.Close()
		out.Close()

		if err != nil {
			log.Printf("Failed to save file: %v", err)
			continue
		}

		filePaths = append(filePaths, tempPath)
		log.Printf("Saved file to: %s", tempPath)
	}

	// Ingest documents
	ctx := r.Context()
	var allDocs []schema.Document
	for _, path := range filePaths {
		docs, err := LoadDocumentsFromFile(path)
		if err != nil {
			log.Printf("Failed to load %s: %v", path, err)
			continue
		}
		allDocs = append(allDocs, docs...)
	}

	if len(allDocs) == 0 {
		sendJSONError(w, "No documents were successfully loaded", http.StatusBadRequest)
		return
	}

	state := IngestionState{Docs: allDocs}
	_, err := s.ingestGraph.Invoke(ctx, state)
	if err != nil {
		log.Printf("Ingestion failed: %v", err)
		sendJSONError(w, "Ingestion failed", http.StatusInternalServerError)
		return
	}

	sendJSONResponse(w, IngestResponse{
		Success:           true,
		Message:           fmt.Sprintf("Successfully ingested %d file(s)", len(filePaths)),
		DocumentsIngested: len(allDocs),
	})
}
