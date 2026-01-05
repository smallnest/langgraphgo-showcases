package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/smallnest/langgraphgo-showcases/DeepInsight/forum_engine"
	"github.com/smallnest/langgraphgo-showcases/DeepInsight/insight_engine"
	"github.com/smallnest/langgraphgo-showcases/DeepInsight/media_engine"
	"github.com/smallnest/langgraphgo-showcases/DeepInsight/query_engine"
	"github.com/smallnest/langgraphgo-showcases/DeepInsight/report_engine"
	"github.com/smallnest/langgraphgo-showcases/DeepInsight/schema"
	"github.com/smallnest/langgraphgo/graph"
)

func main() {
	// Define flags
	var outputFile string
	var simpleMode bool
	flag.StringVar(&outputFile, "o", "", "输出文件路径 (例如: -o report.md)")
	flag.BoolVar(&simpleMode, "simple", false, "简单模式，跳过深度洞察和专家讨论")

	// Parse flags
	flag.Parse()

	// Get the research topic (remaining arguments after flags)
	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("用法: go run main.go [-o 输出文件] [-simple] <研究主题>")
		fmt.Println()
		fmt.Println("参数:")
		fmt.Println("  -o <文件>    指定输出文件路径 (可选)")
		fmt.Println("  -simple      简单模式，跳过深度洞察和专家讨论 (可选)")
		fmt.Println("  <研究主题>    要研究的主题 (必需)")
		fmt.Println()
		fmt.Println("示例:")
		fmt.Println("  go run main.go \"人工智能的发展趋势\"")
		fmt.Println("  go run main.go -o report.md \"人工智能的发展趋势\"")
		fmt.Println("  go run main.go -simple \"人工智能的发展趋势\"")
		fmt.Println("  go run main.go -o output.md -simple \"Claude Code使用经验总结\"")
		return
	}

	if os.Getenv("OPENAI_API_KEY") == "" {
		log.Fatal("错误: 未设置 OPENAI_API_KEY 环境变量。")
	}
	if os.Getenv("TAVILY_API_KEY") == "" {
		log.Fatal("错误: 未设置 TAVILY_API_KEY 环境变量。")
	}

	// Join remaining args as the query (in case topic has spaces without quotes)
	query := strings.Join(args, " ")

	// Initialize state
	initialState := schema.NewDeepInsightState(query)
	initialState.OutputFile = outputFile // Set output file if specified
	initialState.SimpleMode = simpleMode // Set simple mode if specified

	// Create graph with typed state *schema.DeepInsightState
	workflow := graph.NewStateGraph[*schema.DeepInsightState]()

	// Helper to wrap untyped nodes (func(ctx, any) (any, error)) to typed nodes
	wrapNode := func(fn func(ctx context.Context, state any) (any, error)) func(ctx context.Context, state *schema.DeepInsightState) (*schema.DeepInsightState, error) {
		return func(ctx context.Context, s *schema.DeepInsightState) (*schema.DeepInsightState, error) {
			res, err := fn(ctx, s)
			if err != nil {
				return nil, err
			}
			if typedRes, ok := res.(*schema.DeepInsightState); ok {
				return typedRes, nil
			}
			return s, nil
		}
	}

	// Add nodes
	workflow.AddNode("query_engine", "Query research engine", wrapNode(query_engine.QueryEngineNode))
	workflow.AddNode("media_engine", "Media search engine", wrapNode(media_engine.MediaEngineNode))
	workflow.AddNode("report_engine", "Report generation engine", wrapNode(report_engine.ReportEngineNode))

	// Add edges based on mode
	workflow.SetEntryPoint("query_engine")
	workflow.AddEdge("query_engine", "media_engine")

	if simpleMode {
		// Simple mode: skip insight_engine and forum_engine
		workflow.AddEdge("media_engine", "report_engine")
	} else {
		// Full mode: include insight_engine and forum_engine
		workflow.AddNode("insight_engine", "Insight generation engine", wrapNode(insight_engine.InsightEngineNode))
		workflow.AddNode("forum_engine", "Expert forum discussion", wrapNode(forum_engine.ForumEngineNode))
		workflow.AddEdge("media_engine", "insight_engine")
		workflow.AddEdge("insight_engine", "forum_engine")
		workflow.AddEdge("forum_engine", "report_engine")
	}

	workflow.AddEdge("report_engine", graph.END)

	// Compile graph
	app, err := workflow.Compile()
	if err != nil {
		log.Fatalf("编译图失败: %v", err)
	}

	// Run graph
	ctx := context.Background()
	finalState, err := app.Invoke(ctx, initialState)
	if err != nil {
		log.Fatalf("运行图失败: %v", err)
	}

	// Print result
	fmt.Println("\n=== 执行完成 ===")
	fmt.Printf("深度洞察报告已生成，包含 %d 个段落。\n", len(finalState.Paragraphs))
	if finalState.OutputFile != "" {
		fmt.Printf("报告已保存至: %s\n", finalState.OutputFile)
	} else {
		fmt.Printf("报告已保存到文件系统（默认文件名格式: deep_insight_report_<主题>_<时间戳>.md）\n")
	}
}
