package main

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

//go:embed web
var webFS embed.FS

type RunMetadata struct {
	Query     string    `json:"query"`
	Timestamp time.Time `json:"timestamp"`
	DirName   string    `json:"dir_name"` // To know which folder to load if needed, though query is enough if unique
}

func main() {
	// Check for API key
	if os.Getenv("OPENAI_API_KEY") == "" {
		log.Fatal("Please set OPENAI_API_KEY environment variable")
	}
	// Check for API Base if using DeepSeek (optional but recommended for non-OpenAI)
	if os.Getenv("OPENAI_API_BASE") == "" {
		fmt.Println("Warning: OPENAI_API_BASE not set. Defaulting to OpenAI. If using DeepSeek, set this to their API URL.")
	}

	// If arguments are provided, run in CLI mode
	if len(os.Args) > 1 {
		runCLI(os.Args[1])
		return
	}

	// Otherwise, run in Web Server mode
	runServer()
}

func runCLI(query string) {
	fmt.Printf("正在启动 Deer-Flow 研究代理，查询内容：%s\n", query)

	graph, err := NewGraph()
	if err != nil {
		log.Fatalf("Failed to create graph: %v", err)
	}

	initialState := &State{
		Request: Request{
			Query: query,
		},
	}

	result, err := graph.Invoke(context.Background(), initialState)
	if err != nil {
		log.Fatalf("Graph execution failed: %v", err)
	}

	// result is now *State, no type assertion needed
	fmt.Println("\n=== Final Report ===")
	fmt.Println(result.FinalReport)
}

func runServer() {
	subFS, err := fs.Sub(webFS, "web")
	if err != nil {
		log.Fatal(err)
	}
	http.Handle("/", http.FileServer(http.FS(subFS)))

	http.HandleFunc("/api/run", handleRun)
	http.HandleFunc("/api/history", handleHistory)

	fmt.Println("🚀 DeerFlow Web Server running at http://localhost:8085")
	server := &http.Server{
		Addr:              ":8085",
		ReadHeaderTimeout: 3 * time.Second,
	}
	log.Fatal(server.ListenAndServe())
}

// Refactored handleRun to support concurrent logging and result retrieval
func handleRun(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	if query == "" {
		http.Error(w, "Query parameter is required", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	// Check if we have a saved run for this query
	sanitizedQuery := sanitizeFilename(query)
	dataDir := filepath.Join("showcases", "deerflow", "data", sanitizedQuery)
	if _, err := os.Stat(dataDir); err == nil {
		// Data exists, replay it
		replayRun(w, flusher, dataDir)
		return
	}

	// Send initial status
	sendSSE(w, flusher, "update", map[string]string{"step": "正在初始化..."})

	g, err := NewGraph()
	if err != nil {
		sendSSE(w, flusher, "error", map[string]string{"message": err.Error()})
		return
	}

	initialState := &State{
		Request: Request{
			Query: query,
		},
	}

	logChan := make(chan string, 100)
	resultChan := make(chan *State, 1)
	errChan := make(chan error, 1)

	ctx := context.WithValue(context.Background(), logKey{}, logChan)

	go func() {
		defer close(logChan)
		defer close(resultChan)
		defer close(errChan)

		res, err := g.Invoke(ctx, initialState)
		if err != nil {
			errChan <- err
			return
		}
		resultChan <- res // res is already *State, no type assertion needed
	}()

	var capturedLogs []string

	// Loop to handle logs and result
	for {
		select {
		case msg, ok := <-logChan:
			if !ok {
				logChan = nil // Channel closed
			} else {
				capturedLogs = append(capturedLogs, msg)
				sendSSE(w, flusher, "log", map[string]string{"message": msg})
			}
		case res, ok := <-resultChan:
			if !ok {
				resultChan = nil
			} else {
				// Save run data
				saveRun(dataDir, query, capturedLogs, res.FinalReport, res.PodcastScript)
				sendSSE(w, flusher, "result", map[string]string{
					"report":         res.FinalReport,
					"podcast_script": res.PodcastScript,
				})
				return // Done
			}
		case err, ok := <-errChan:
			if !ok {
				errChan = nil
			} else {
				sendSSE(w, flusher, "error", map[string]string{"message": err.Error()})
				return
			}
		}

		if logChan == nil && resultChan == nil && errChan == nil {
			break
		}
	}
}

func sanitizeFilename(name string) string {
	// Replace invalid characters with underscore
	reg := regexp.MustCompile(`[^a-zA-Z0-9\p{Han}]+`)
	safe := reg.ReplaceAllString(name, "_")
	// Trim underscores
	safe = strings.Trim(safe, "_")
	// Limit length
	if len(safe) > 100 {
		safe = safe[:100]
	}
	return safe
}

func saveRun(dir string, query string, logs []string, report string, podcastScript string) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Printf("Failed to create data dir: %v", err)
		return
	}

	// Save metadata
	meta := RunMetadata{
		Query:     query,
		Timestamp: time.Now(),
		DirName:   filepath.Base(dir),
	}
	metaData, _ := json.Marshal(meta)
	if err := os.WriteFile(filepath.Join(dir, "metadata.json"), metaData, 0600); err != nil {
		log.Printf("Failed to save metadata: %v", err)
	}

	// Save logs
	logsData, _ := json.Marshal(logs)
	if err := os.WriteFile(filepath.Join(dir, "logs.json"), logsData, 0600); err != nil {
		log.Printf("Failed to save logs: %v", err)
	}

	// Save report
	if err := os.WriteFile(filepath.Join(dir, "report.html"), []byte(report), 0600); err != nil {
		log.Printf("Failed to save report: %v", err)
	}

	// Save podcast script if exists
	if podcastScript != "" {
		if err := os.WriteFile(filepath.Join(dir, "podcast.txt"), []byte(podcastScript), 0600); err != nil {
			log.Printf("Failed to save podcast script: %v", err)
		}
	}
}

func replayRun(w http.ResponseWriter, flusher http.Flusher, dir string) {
	// Read logs
	logsData, err := os.ReadFile(filepath.Join(dir, "logs.json"))
	if err != nil {
		sendSSE(w, flusher, "error", map[string]string{"message": "Failed to load saved logs"})
		return
	}
	var logs []string
	if err := json.Unmarshal(logsData, &logs); err != nil {
		sendSSE(w, flusher, "error", map[string]string{"message": "Failed to parse saved logs"})
		return
	}

	// Read report
	reportData, err := os.ReadFile(filepath.Join(dir, "report.html"))
	if err != nil {
		sendSSE(w, flusher, "error", map[string]string{"message": "Failed to load saved report"})
		return
	}

	// Read podcast script (optional)
	podcastScript := ""
	podcastData, err := os.ReadFile(filepath.Join(dir, "podcast.txt"))
	if err == nil {
		podcastScript = string(podcastData)
	}

	sendSSE(w, flusher, "update", map[string]string{"step": "正在从缓存回放..."})

	// Replay logs with simulated delay
	for _, msg := range logs {
		sendSSE(w, flusher, "log", map[string]string{"message": msg})
		// Simulate delay (faster than real-time but noticeable)
		time.Sleep(200 * time.Millisecond)
	}

	// Send result
	sendSSE(w, flusher, "result", map[string]string{
		"report":         string(reportData),
		"podcast_script": podcastScript,
	})
}

func handleHistory(w http.ResponseWriter, r *http.Request) {
	dataRoot := filepath.Join("showcases", "deerflow", "data")
	entries, err := os.ReadDir(dataRoot)
	if err != nil {
		http.Error(w, "Failed to read history", http.StatusInternalServerError)
		return
	}

	var history []RunMetadata
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		metaPath := filepath.Join(dataRoot, entry.Name(), "metadata.json")
		data, err := os.ReadFile(metaPath)
		if err != nil {
			continue // Skip if no metadata
		}

		var meta RunMetadata
		if err := json.Unmarshal(data, &meta); err == nil {
			history = append(history, meta)
		}
	}

	// Sort by timestamp desc
	sort.Slice(history, func(i, j int) bool {
		return history[i].Timestamp.After(history[j].Timestamp)
	})

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(history); err != nil {
		log.Printf("Failed to encode history: %v", err)
	}
}

func sendSSE(w http.ResponseWriter, flusher http.Flusher, eventType string, data any) {
	payload := map[string]any{
		"type": eventType,
	}

	// Merge data into payload
	if m, ok := data.(map[string]string); ok {
		for k, v := range m {
			payload[k] = v
		}
	}

	jsonPayload, _ := json.Marshal(payload)
	fmt.Fprintf(w, "data: %s\n\n", jsonPayload)
	flusher.Flush()
}
