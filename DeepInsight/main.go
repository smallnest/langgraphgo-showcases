package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/smallnest/langgraphgo/graph"
	"github.com/smallnest/langgraphgo/showcases/DeepInsight/forum_engine"
	"github.com/smallnest/langgraphgo/showcases/DeepInsight/insight_engine"
	"github.com/smallnest/langgraphgo/showcases/DeepInsight/media_engine"
	"github.com/smallnest/langgraphgo/showcases/DeepInsight/query_engine"
	"github.com/smallnest/langgraphgo/showcases/DeepInsight/report_engine"
	"github.com/smallnest/langgraphgo/showcases/DeepInsight/schema"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("用法: go run main.go <研究主题>")
		fmt.Println("示例: go run main.go \"人工智能的发展趋势\"")
		return
	}

	if os.Getenv("OPENAI_API_KEY") == "" {
		log.Fatal("错误: 未设置 OPENAI_API_KEY 环境变量。")
	}
	if os.Getenv("TAVILY_API_KEY") == "" {
		log.Fatal("错误: 未设置 TAVILY_API_KEY 环境变量。")
	}

	query := os.Args[1]

	// Initialize state
	initialState := schema.NewDeepInsightState(query)

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
	workflow.AddNode("insight_engine", "Insight generation engine", wrapNode(insight_engine.InsightEngineNode))
	workflow.AddNode("forum_engine", "Expert forum discussion", wrapNode(forum_engine.ForumEngineNode))
	workflow.AddNode("report_engine", "Report generation engine", wrapNode(report_engine.ReportEngineNode))

	// Add edges
	workflow.SetEntryPoint("query_engine")
	workflow.AddEdge("query_engine", "media_engine")
	workflow.AddEdge("media_engine", "insight_engine")
	workflow.AddEdge("insight_engine", "forum_engine")
	workflow.AddEdge("forum_engine", "report_engine")
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
	fmt.Printf("报告已保存到文件系统中。\n")
}
