package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/smallnest/langgraphgo/graph"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/schema"
)

// RetrievalGraphBuilder builds the retrieval graph that handles user questions.
// The graph follows the original TypeScript implementation:
// START -> checkQueryType (router) -> [retrieveDocuments -> generateResponse | directAnswer] -> END
func RetrievalGraphBuilder(vs *VectorStore, cfg Config) *graph.StateGraph[RetrievalState] {
	g := graph.NewStateGraph[RetrievalState]()

	// Node 1: checkQueryType - Router that decides if we need to retrieve documents
	g.AddNode("checkQueryType", "Determine if query needs document retrieval", func(ctx context.Context, state RetrievalState) (RetrievalState, error) {
		log.Printf("Node: checkQueryType - Question: %s", state.Question)

		opts := []openai.Option{openai.WithModel(cfg.Model)}
		if cfg.OpenAIBaseURL != "" {
			opts = append(opts, openai.WithBaseURL(cfg.OpenAIBaseURL))
		}

		llm, err := openai.New(opts...)
		if err != nil {
			return state, fmt.Errorf("failed to create LLM: %w", err)
		}

		// Create routing prompt
		routePrompt := fmt.Sprintf(`You are a router. Determine if the user's query requires retrieving information from PDF documents or if it can be answered directly as a generic chat question.

User Query: %s

Analyze the query and decide:
1. If the query asks about specific information that would be contained in documents (facts, details, explanations from PDFs), respond with {"route": "retrieve"}
2. If the query is a generic greeting, question about capabilities, or can be answered without document context, respond with {"route": "direct"}

Respond with ONLY a JSON object, nothing else: {"route": "retrieve"}`, state.Question)

		// Get routing decision
		resp, err := llm.Call(ctx, routePrompt)
		if err != nil {
			// Fallback: check if question seems document-related
			route := "direct"
			if containsDocKeywords(state.Question) {
				route = "retrieve"
			}
			log.Printf("LLM error, using fallback route: %s", route)
			return RetrievalState{Question: state.Question, Route: route, Messages: state.Messages}, nil
		}

		// Parse JSON response
		var routeResp RouteResponse
		jsonStr := strings.TrimSpace(resp)

		// Clean up markdown code blocks if present
		if strings.HasPrefix(jsonStr, "```json") {
			jsonStr = strings.TrimPrefix(jsonStr, "```json")
			jsonStr = strings.TrimSuffix(jsonStr, "```")
			jsonStr = strings.TrimSpace(jsonStr)
		} else if strings.HasPrefix(jsonStr, "```") {
			jsonStr = strings.TrimPrefix(jsonStr, "```")
			jsonStr = strings.TrimSuffix(jsonStr, "```")
			jsonStr = strings.TrimSpace(jsonStr)
		}

		if err := json.Unmarshal([]byte(jsonStr), &routeResp); err != nil {
			// Fallback: keyword-based routing
			route := "direct"
			if containsDocKeywords(state.Question) {
				route = "retrieve"
			}
			log.Printf("JSON parse error, using fallback route: %s", route)
			return RetrievalState{Question: state.Question, Route: route, Messages: state.Messages}, nil
		}

		log.Printf("Route decided: %s", routeResp.Route)
		return RetrievalState{Question: state.Question, Route: routeResp.Route, Messages: state.Messages}, nil
	})

	// Node 2: retrieveDocuments - Search vector store for relevant documents
	g.AddNode("retrieveDocuments", "Retrieve relevant documents from vector store", func(ctx context.Context, state RetrievalState) (RetrievalState, error) {
		log.Printf("Node: retrieveDocuments - Query: %s", state.Question)

		docs, err := vs.SimilaritySearch(ctx, state.Question, cfg.TopK)
		if err != nil {
			return state, fmt.Errorf("failed to search vector store: %w", err)
		}

		log.Printf("Retrieved %d document(s)", len(docs))
		return RetrievalState{
			Question:  state.Question,
			Route:     state.Route,
			Documents: docs,
			Messages:  state.Messages,
		}, nil
	})

	// Node 3: generateResponse - Generate response using retrieved documents
	g.AddNode("generateResponse", "Generate response with document context", func(ctx context.Context, state RetrievalState) (RetrievalState, error) {
		log.Println("Node: generateResponse")

		opts := []openai.Option{openai.WithModel(cfg.Model)}
		if cfg.OpenAIBaseURL != "" {
			opts = append(opts, openai.WithBaseURL(cfg.OpenAIBaseURL))
		}

		llm, err := openai.New(opts...)
		if err != nil {
			return state, fmt.Errorf("failed to create LLM: %w", err)
		}

		// Format context from retrieved documents
		var contextParts []string
		for i, doc := range state.Documents {
			contextParts = append(contextParts, fmt.Sprintf("[Source %d]\n%s", i+1, doc.PageContent))
		}
		contextStr := strings.Join(contextParts, "\n\n---\n\n")

		// Build messages
		messages := state.Messages

		// Add system message
		systemMsg := llms.TextParts(llms.ChatMessageTypeSystem,
			`You are an AI assistant that answers questions based on the provided PDF document context.

Rules:
1. Use ONLY the provided context to answer questions
2. If the answer is not in the context, say "I don't have enough information in the documents to answer this question."
3. Be concise and accurate
4. Cite the source number when using information from a specific document`)
		messages = append(messages, systemMsg)

		// Add user question with context
		userPrompt := fmt.Sprintf(`Context from documents:
%s

Question: %s

Please answer the question using only the context provided above.`, contextStr, state.Question)
		messages = append(messages, llms.TextParts(llms.ChatMessageTypeHuman, userPrompt))

		// Generate response
		resp, err := llm.GenerateContent(ctx, messages)
		if err != nil {
			return state, fmt.Errorf("failed to generate response: %w", err)
		}

		// Extract response content
		content := ""
		if len(resp.Choices) > 0 {
			content = resp.Choices[0].Content
		}

		// Create AI message
		aiMsg := llms.TextParts(llms.ChatMessageTypeAI, content)

		// Update messages with user question and AI response
		newMessages := append(state.Messages,
			llms.TextParts(llms.ChatMessageTypeHuman, state.Question),
			aiMsg,
		)

		log.Printf("Generated response: %d chars", len(content))
		return RetrievalState{
			Question:  state.Question,
			Route:     state.Route,
			Documents: state.Documents,
			Messages:  newMessages,
		}, nil
	})

	// Node 4: directAnswer - Answer without document retrieval
	g.AddNode("directAnswer", "Answer query directly without documents", func(ctx context.Context, state RetrievalState) (RetrievalState, error) {
		log.Println("Node: directAnswer")

		opts := []openai.Option{openai.WithModel(cfg.Model)}
		if cfg.OpenAIBaseURL != "" {
			opts = append(opts, openai.WithBaseURL(cfg.OpenAIBaseURL))
		}

		llm, err := openai.New(opts...)
		if err != nil {
			return state, fmt.Errorf("failed to create LLM: %w", err)
		}

		// Build messages
		messages := state.Messages

		// Add system message for direct chat
		systemMsg := llms.TextParts(llms.ChatMessageTypeSystem,
			`You are a helpful AI assistant. Answer the user's questions directly and conversationally.`)
		messages = append(messages, systemMsg)

		// Add user question
		messages = append(messages, llms.TextParts(llms.ChatMessageTypeHuman, state.Question))

		// Generate response
		resp, err := llm.GenerateContent(ctx, messages)
		if err != nil {
			return state, fmt.Errorf("failed to generate response: %w", err)
		}

		// Extract response content
		content := ""
		if len(resp.Choices) > 0 {
			content = resp.Choices[0].Content
		}

		// Create AI message
		aiMsg := llms.TextParts(llms.ChatMessageTypeAI, content)

		// Update messages with user question and AI response
		newMessages := append(state.Messages,
			llms.TextParts(llms.ChatMessageTypeHuman, state.Question),
			aiMsg,
		)

		log.Printf("Direct answer: %d chars", len(content))
		return RetrievalState{
			Question: state.Question,
			Route:    state.Route,
			Messages: newMessages,
		}, nil
	})

	// Set up edges
	g.SetEntryPoint("checkQueryType")

	// Conditional edge from checkQueryType
	g.AddConditionalEdge("checkQueryType", func(ctx context.Context, state RetrievalState) string {
		if state.Route == "retrieve" {
			return "retrieveDocuments"
		}
		return "directAnswer"
	})

	// Edges to END
	g.AddEdge("retrieveDocuments", "generateResponse")
	g.AddEdge("generateResponse", graph.END)
	g.AddEdge("directAnswer", graph.END)

	return g
}

// ============================================
// Helper Functions
// ============================================

// containsDocKeywords checks if a query seems to require document retrieval.
func containsDocKeywords(query string) bool {
	queryLower := strings.ToLower(query)
	docKeywords := []string{
		"what", "how", "explain", "describe", "tell me", "find", "search",
		"according", "document", "pdf", "file", "information", "details",
	}

	for _, keyword := range docKeywords {
		if strings.Contains(queryLower, keyword) {
			return true
		}
	}

	// Also check if it's a longer question (likely needs retrieval)
	return len(strings.Fields(query)) > 5
}

// Query processes a user question through the retrieval graph.
func Query(ctx context.Context, vs *VectorStore, cfg Config, question string, history []llms.MessageContent) (string, []schema.Document, error) {
	// Build retrieval graph
	retrievalGraph := RetrievalGraphBuilder(vs, cfg)
	runnable, err := retrievalGraph.Compile()
	if err != nil {
		return "", nil, fmt.Errorf("failed to compile retrieval graph: %w", err)
	}

	// Invoke graph
	state := RetrievalState{
		Question: question,
		Messages: history,
	}

	finalState, err := runnable.Invoke(ctx, state)
	if err != nil {
		return "", nil, fmt.Errorf("retrieval failed: %w", err)
	}

	// Extract response from last AI message
	var response string
	if len(finalState.Messages) > 0 {
		lastMsg := finalState.Messages[len(finalState.Messages)-1]
		if lastMsg.Role == llms.ChatMessageTypeAI && len(lastMsg.Parts) > 0 {
			if textPart, ok := lastMsg.Parts[0].(llms.TextContent); ok {
				response = textPart.Text
			}
		}
	}

	return response, finalState.Documents, nil
}
