package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/smallnest/langgraphgo/showcases/deepagents/agent"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

func main() {
	if os.Getenv("OPENAI_API_KEY") == "" {
		log.Println("OPENAI_API_KEY not set, skipping example execution")
		return
	}

	ctx := context.Background()

	// Initialize LLM
	model, err := openai.New()
	if err != nil {
		log.Fatalf("Failed to create LLM: %v", err)
	}

	// Subagent handler example
	subAgentHandler := func(ctx context.Context, task string) (string, error) {
		log.Printf("[SubAgent] Received task: %s", task)
		return "Subagent completed task: " + task, nil
	}

	// Create Deep Agent
	deepAgent, err := agent.CreateDeepAgent(model,
		agent.WithRootDir("./workspace"),
		agent.WithSystemPrompt("You are a capable assistant with filesystem access. Use tools to complete tasks."),
		agent.WithSubAgentHandler(subAgentHandler),
	)
	if err != nil {
		log.Fatalf("Failed to create deep agent: %v", err)
	}

	// Run the agent
	inputs := map[string]any{
		"messages": []llms.MessageContent{
			llms.TextParts(llms.ChatMessageTypeHuman, "Create a file named 'hello.txt' with content 'Hello, DeepAgents!', then read it back to confirm."),
		},
	}

	log.Println("Starting Deep Agent...")
	result, err := deepAgent.Invoke(ctx, inputs)
	if err != nil {
		log.Fatalf("Agent execution failed: %v", err)
	}

	// Print result - result is already map[string]any
	mState := result
	messages := mState["messages"].([]llms.MessageContent)
	lastMsg := messages[len(messages)-1]

	if len(lastMsg.Parts) > 0 {
		if textPart, ok := lastMsg.Parts[0].(llms.TextContent); ok {
			fmt.Printf("Agent Response: %s\n", textPart.Text)
		}
	}
}
