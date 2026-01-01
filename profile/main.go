package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/microcosm-cc/bluemonday"
	"github.com/smallnest/langgraphgo/showcases/profile/graph"
)

//go:embed web
var webFS embed.FS

// 限流控制：确保同时只能有一个画像操作正在进行
var (
	profilingMutex sync.Mutex
	isProfiling    bool
)

func main() {
	webMode := flag.Bool("web", false, "以 Web 模式运行")
	addr := flag.String("addr", ":8086", "Web 服务器地址")
	username := flag.String("username", "", "要分析的用户名 (CLI 模式)")
	flag.Parse()

	if *webMode {
		runWeb(*addr)
	} else {
		if *username == "" {
			fmt.Println("请使用 -username 提供用户名，或使用 -web 运行 Web 模式")
			os.Exit(1)
		}
		runCLI(*username)
	}
}

func runCLI(username string) {
	ctx := context.Background()
	g, err := graph.NewGraph()
	if err != nil {
		log.Fatalf("创建图失败: %v", err)
	}

	initialState := &graph.State{
		Username: username,
	}

	res, err := g.Invoke(ctx, initialState)
	if err != nil {
		log.Fatalf("运行图失败: %v", err)
	}

	fmt.Println("\n=== 生成的画像 ===")
	fmt.Println(res.ProfileText)
}

func runWeb(addr string) {
	// Serve static files from web directory
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(webFS))))

	http.HandleFunc("/stream", streamHandler)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFS(webFS, "web/templates/index.html")
		if err != nil {
			http.Error(w, fmt.Sprintf("解析模板失败: %v", err), http.StatusInternalServerError)
			return
		}
		_ = tmpl.Execute(w, nil)
	})

	fmt.Println("服务器启动于", addr)
	server := &http.Server{
		Addr:              addr,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       60 * time.Second,
		// WriteTimeout is explicitly set to 0 (unlimited) to support Server-Sent Events (SSE).
		// Setting a short WriteTimeout would break the streaming functionality.
		WriteTimeout: 0,
		IdleTimeout:  120 * time.Second,
	}
	log.Fatal(server.ListenAndServe())
}

func streamHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	username := r.URL.Query().Get("username")
	if username == "" {
		fmt.Fprintf(w, "event: error\ndata: 用户名是必填的\n\n")
		flusher.Flush()
		return
	}

	// 限流检查：如果正在为其他用户生成画像，返回限流消息
	profilingMutex.Lock()
	if isProfiling {
		profilingMutex.Unlock()
		fmt.Fprintf(w, "event: error\ndata: 限流: 正在为其他人画像，请等候几分钟再试\n\n")
		flusher.Flush()
		return
	}
	// 设置为正在处理状态
	isProfiling = true
	profilingMutex.Unlock()

	// 确保在函数退出时释放锁
	defer func() {
		profilingMutex.Lock()
		isProfiling = false
		profilingMutex.Unlock()
	}()

	logChan := make(chan string, 10)
	ctx := context.Background()

	initialState := &graph.State{
		Username: username,
		LogChan:  logChan,
	}

	resultChan := make(chan string, 1)

	go func() {
		defer close(logChan)
		defer close(resultChan)

		g, err := graph.NewGraph()
		if err != nil {
			logChan <- fmt.Sprintf("Error: 创建图失败: %v", err)
			return
		}

		res, err := g.Invoke(ctx, initialState)
		if err != nil {
			logChan <- fmt.Sprintf("Error: 运行图失败: %v", err)
			return
		}

		// 渲染结果 HTML
		tmpl, err := template.ParseFS(webFS, "web/templates/index.html")
		if err != nil {
			logChan <- fmt.Sprintf("Error: 解析模板失败: %v", err)
			return
		}

		profileText := res.ProfileText
		socialData := res.SocialData

		// Convert Markdown to HTML
		extensions := parser.CommonExtensions | parser.AutoHeadingIDs
		p := parser.NewWithExtensions(extensions)
		doc := p.Parse([]byte(profileText))

		htmlFlags := html.CommonFlags | html.HrefTargetBlank
		opts := html.RendererOptions{Flags: htmlFlags}
		renderer := html.NewRenderer(opts)
		htmlBytes := markdown.Render(doc, renderer)

		// Fix G203: Sanitize HTML before determining it is safe
		sanitizer := bluemonday.UGCPolicy()
		htmlBytes = sanitizer.SanitizeBytes(htmlBytes)

		var sb strings.Builder
		data := struct {
			Username    string
			Profile     string
			ProfileHTML template.HTML
			Results     any
		}{
			Username:    username,
			Profile:     profileText,
			ProfileHTML: template.HTML(htmlBytes), // #nosec G203
			Results:     socialData,
		}

		if err := tmpl.ExecuteTemplate(&sb, "result", data); err != nil {
			logChan <- fmt.Sprintf("Error: 渲染模板失败: %v", err)
			return
		}

		resultChan <- sb.String()
	}()

	for msg := range logChan {
		if strings.HasPrefix(msg, "Error:") {
			fmt.Fprintf(w, "event: error\ndata: %s\n\n", strings.TrimPrefix(msg, "Error: "))
		} else {
			fmt.Fprintf(w, "data: %s\n\n", msg)
		}
		flusher.Flush()
	}

	// 检查是否有结果
	resHtml := <-resultChan
	if resHtml != "" {
		fmt.Fprintf(w, "event: result\n")
		// Use manual split instead of splitSeq for older go versions or just for reliability
		lines := strings.SplitSeq(resHtml, "\n")
		for line := range lines {
			fmt.Fprintf(w, "data: %s\n", line)
		}
		fmt.Fprintf(w, "\n")
		flusher.Flush()
	}
}
