package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/smallnest/langgraphgo/graph"
	"github.com/smallnest/langgraphgo/showcases/BettaFish/forum_engine"
	"github.com/smallnest/langgraphgo/showcases/BettaFish/insight_engine"
	"github.com/smallnest/langgraphgo/showcases/BettaFish/media_engine"
	"github.com/smallnest/langgraphgo/showcases/BettaFish/query_engine"
	"github.com/smallnest/langgraphgo/showcases/BettaFish/report_engine"
	"github.com/smallnest/langgraphgo/showcases/BettaFish/schema"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("用法: go run main.go <查询>")
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
	initialState := schema.NewBettaFishState(query)

	// Create graph with typed state *schema.BettaFishState
	workflow := graph.NewStateGraph[*schema.BettaFishState]()

	// Helper to wrap untyped nodes (func(ctx, any) (any, error)) to typed nodes
	// The engines were written with untyped signature but expect *BettaFishState inside.
	// So we can just call them, but we need to cast the result back to *BettaFishState if it returns any.
	wrapNode := func(fn func(ctx context.Context, state any) (any, error)) func(ctx context.Context, state *schema.BettaFishState) (*schema.BettaFishState, error) {
		return func(ctx context.Context, s *schema.BettaFishState) (*schema.BettaFishState, error) {
			res, err := fn(ctx, s)
			if err != nil {
				return nil, err
			}
			if typedRes, ok := res.(*schema.BettaFishState); ok {
				return typedRes, nil
			}
			// If it returns something else (or nil), we return the original state modified in place (since it's a pointer)
			// But wait, if fn modifies *s in place and returns it as 'any', typedRes assertion works.
			return s, nil
		}
	}

	// Add nodes
	workflow.AddNode("query_engine", "Query analysis engine", wrapNode(query_engine.QueryEngineNode))
	workflow.AddNode("media_engine", "Media search engine", wrapNode(media_engine.MediaEngineNode))
	workflow.AddNode("insight_engine", "Insight generation engine", wrapNode(insight_engine.InsightEngineNode))
	workflow.AddNode("forum_engine", "Forum search engine", wrapNode(forum_engine.ForumEngineNode))
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
	fmt.Printf("报告已生成，包含 %d 个段落。\n", len(finalState.Paragraphs))
}
