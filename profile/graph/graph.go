package graph

import (
	"context"

	g "github.com/smallnest/langgraphgo/graph"
)

func NewGraph() (*g.StateRunnable[*State], error) {
	workflow := g.NewStateGraph[*State]()

	workflow.AddNode("account", "提取用户ID", func(ctx context.Context, state *State) (*State, error) {
		result, err := AccountNode(ctx, state)
		if err != nil {
			return nil, err
		}
		if resultState, ok := result.(*State); ok {
			return resultState, nil
		}
		return state, nil
	})
	workflow.AddNode("search", "搜索社交资料", func(ctx context.Context, state *State) (*State, error) {
		result, err := SearchNode(ctx, state)
		if err != nil {
			return nil, err
		}
		if resultState, ok := result.(*State); ok {
			return resultState, nil
		}
		return state, nil
	})
	workflow.AddNode("profile", "生成用户画像", func(ctx context.Context, state *State) (*State, error) {
		result, err := ProfileNode(ctx, state)
		if err != nil {
			return nil, err
		}
		if resultState, ok := result.(*State); ok {
			return resultState, nil
		}
		return state, nil
	})

	workflow.SetEntryPoint("account")
	workflow.AddEdge("account", "search")
	workflow.AddEdge("search", "profile")
	workflow.AddEdge("profile", g.END)

	return workflow.Compile()
}
