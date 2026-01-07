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
	"sync"
	"time"

	"github.com/google/uuid"
)

//go:embed web
var webFS embed.FS

// Shared report storage
type SharedReport struct {
	ID        string    `json:"id"`
	HTML      string    `json:"html"`
	Query     string    `json:"query"`
	CreatedAt time.Time `json:"created_at"`
}

var (
	reportsMap = make(map[string]*SharedReport)
	reportsMu  sync.RWMutex
)

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
	fmt.Printf("正在启动 Insight 研究代理，查询内容：%s\n", query)

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
	http.HandleFunc("/api/share", handleShare)
	http.HandleFunc("/reports/", handleReportView)

	fmt.Println("🚀 Insight Web Server running at http://localhost:8085")
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
	dataDir := filepath.Join("data", sanitizedQuery)
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
	dataRoot := filepath.Join("data")
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

func handleShare(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		HTML  string `json:"html"`
		Query string `json:"query"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Generate unique ID for this report
	reportID := uuid.New().String()

	report := &SharedReport{
		ID:        reportID,
		HTML:      req.HTML,
		Query:     req.Query,
		CreatedAt: time.Now(),
	}

	// Store the report
	reportsMu.Lock()
	reportsMap[reportID] = report
	reportsMu.Unlock()

	// Determine scheme (http or https)
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	} else if proto := r.Header.Get("X-Forwarded-Proto"); proto != "" {
		scheme = proto
	}

	// Build the share URL using the request's host
	shareURL := fmt.Sprintf("%s://%s/reports/%s", scheme, r.Host, reportID)

	// Return the share URL
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"id":  reportID,
		"url": shareURL,
	})
}

func handleReportView(w http.ResponseWriter, r *http.Request) {
	// Extract report ID from URL
	// URL format: /reports/{id}
	id := strings.TrimPrefix(r.URL.Path, "/reports/")
	if id == "" {
		http.Error(w, "Report ID is required", http.StatusBadRequest)
		return
	}

	// Retrieve the report
	reportsMu.RLock()
	report, exists := reportsMap[id]
	reportsMu.RUnlock()

	if !exists {
		http.Error(w, "Report not found", http.StatusNotFound)
		return
	}

	// Render the shared report page
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s - Insight 研究报告</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Inter', sans-serif;
            background-color: #f7f5f0;
            color: #333;
            line-height: 1.6;
        }
        .header {
            background: white;
            padding: 20px 40px;
            border-bottom: 1px solid #e6e4dd;
            display: flex;
            justify-content: space-between;
            align-items: center;
        }
        .header h1 {
            font-size: 18px;
            font-weight: 600;
            color: #333;
        }
        .header .brand {
            font-size: 14px;
            color: #666;
        }
        .content {
            max-width: 850px;
            margin: 40px auto;
            padding: 60px 60px;
            background: white;
            box-shadow: 0 12px 40px rgba(0, 0, 0, 0.12);
            border-radius: 2px;
            min-height: calc(100vh - 160px);
        }
        .content h1 {
            font-size: 2.8em;
            font-weight: 800;
            margin-bottom: 40px;
            padding-bottom: 40px;
            border-bottom: 3px solid #333;
            line-height: 1.5;
        }
        .content h2 {
            font-size: 2.0em;
            font-weight: 700;
            margin-top: 30px;
            margin-bottom: 20px;
            color: #333;
        }
        .content h3 {
            font-size: 1.5em;
            font-weight: 600;
            margin-top: 20px;
            margin-bottom: 15px;
        }
        .content p {
            margin-bottom: 16px;
            color: #333;
            line-height: 1.7;
        }
        .content ul, .content ol {
            margin-bottom: 24px;
            padding-left: 28px;
        }
        .content li {
            margin-bottom: 8px;
        }
        .content code {
            background-color: #f6f8fa;
            padding: 0.2em 0.4em;
            border-radius: 3px;
            font-family: 'Menlo', 'Monaco', monospace;
            font-size: 0.9em;
        }
        .content pre {
            background-color: #f6f8fa;
            padding: 16px;
            border-radius: 6px;
            overflow-x: auto;
            margin-bottom: 24px;
        }
        .content pre code {
            background: none;
            padding: 0;
        }
        .content blockquote {
            border-left: 4px solid #d97757;
            padding: 16px 24px;
            margin: 24px 0;
            background-color: #f8f9fa;
            color: #666;
        }
        .content strong {
            font-weight: 600;
        }
        .mermaid-diagram {
            text-align: center;
            margin: 30px 0;
            padding: 20px;
            background: #fafafa;
            border-radius: 8px;
        }
        .katex {
            font-size: 1.1em;
        }
    </style>
    <script src="https://s4.zstatic.net/ajax/libs/KaTeX/0.16.9/katex.min.js"></script>
    <script src="https://s4.zstatic.net/ajax/libs/KaTeX/0.16.9/contrib/auto-render.min.js"></script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.11.1/highlight.min.js"></script>
    <script src="https://s4.zstatic.net/ajax/libs/mermaid/11.12.0/mermaid.min.js"></script>
</head>
<body>
    <div class="header">
        <h1>%s</h1>
        <div class="brand">由 <a href="https://insight.rpcx.io" target="_blank" style="color: #d97757; text-decoration: none; font-weight: 500;" onmouseover="this.style.color='#c56245'" onmouseout="this.style.color='#d97757'">Insight</a> AI 研究助手生成</div>
    </div>
    <div class="content">
        %s
    </div>
    <script>
        // Initialize Mermaid
        mermaid.initialize({
            startOnLoad: true,
            theme: 'default',
            securityLevel: 'loose'
        });

        // Render math
        renderMathInElement(document.body, {
            delimiters: [
                {left: '$$', right: '$$', display: true},
                {left: '$', right: '$', display: false}
            ]
        });

        // Highlight code
        document.querySelectorAll('pre code').forEach((block) => {
            hljs.highlightElement(block);
        });
    </script>
</body>
</html>`, report.Query, report.Query, report.HTML)
}
