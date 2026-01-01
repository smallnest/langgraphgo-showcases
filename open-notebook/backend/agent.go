package backend

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
	ollamallm "github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/prompts"
)

// Agent handles AI operations for generating notes and chat responses
type Agent struct {
	vectorStore *VectorStore
	llm         llms.Model
	cfg         Config
}

// NewAgent creates a new agent
func NewAgent(cfg Config, vectorStore *VectorStore) (*Agent, error) {
	llm, err := createLLM(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create LLM: %w", err)
	}

	return &Agent{
		vectorStore: vectorStore,
		llm:         llm,
		cfg:         cfg,
	}, nil
}

// createLLM creates an LLM based on configuration
func createLLM(cfg Config) (llms.Model, error) {
	if cfg.IsOllama() {
		return ollamallm.New(
			ollamallm.WithModel(cfg.OllamaModel),
			ollamallm.WithServerURL(cfg.OllamaBaseURL),
		)
	}

	opts := []openai.Option{
		openai.WithToken(cfg.OpenAIAPIKey),
		openai.WithModel(cfg.OpenAIModel),
	}
	if cfg.OpenAIBaseURL != "" {
		opts = append(opts, openai.WithBaseURL(cfg.OpenAIBaseURL))
	}

	return openai.New(opts...)
}

// GenerateTransformation generates a note based on transformation type
func (a *Agent) GenerateTransformation(ctx context.Context, req *TransformationRequest, sources []Source) (*TransformationResponse, error) {
	// Build context from sources
	var sourceContext strings.Builder
	for i, src := range sources {
		sourceContext.WriteString(fmt.Sprintf("\n## Source %d: %s\n", i+1, src.Name))
		if src.Content != "" && len(src.Content) < 10000 {
			sourceContext.WriteString(src.Content)
		} else {
			sourceContext.WriteString(fmt.Sprintf("[Source content: %s, type: %s]", src.Name, src.Type))
		}
		sourceContext.WriteString("\n")
	}

	// Build prompt using f-string format (no Go template reserved names issue)
	promptTemplate := a.getTransformationPrompt(req)

	prompt := prompts.NewPromptTemplate(
		promptTemplate,
		[]string{"sources", "type", "length", "format", "prompt"},
	)
	prompt.TemplateFormat = prompts.TemplateFormatFString

	promptValue, err := prompt.Format(map[string]any{
		"sources": sourceContext.String(),
		"type":    req.Type,
		"length":  req.Length,
		"format":  req.Format,
		"prompt":  req.Prompt,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to format prompt: %w", err)
	}

	// Generate response
	response, err := llms.GenerateFromSinglePrompt(ctx, a.llm, promptValue)
	if err != nil {
		return nil, fmt.Errorf("failed to generate response: %w", err)
	}

	// Build source summaries
	sourceSummaries := make([]SourceSummary, len(sources))
	for i, src := range sources {
		sourceSummaries[i] = SourceSummary{
			ID:   src.ID,
			Name: src.Name,
			Type: src.Type,
		}
	}

	return &TransformationResponse{
		Type:      req.Type,
		Content:   response,
		Sources:   sourceSummaries,
		CreatedAt: time.Now(),
		Metadata: map[string]interface{}{
			"length": req.Length,
			"format": req.Format,
		},
	}, nil
}

// getTransformationPrompt returns the prompt template for each transformation type
func (a *Agent) getTransformationPrompt(req *TransformationRequest) string {
	switch req.Type {
	case "summary":
		return `You are an expert at creating comprehensive summaries. Based on the following sources, create a {length} summary in {format} format.

Sources:
{sources}

Provide a well-structured summary that captures the key information, main themes, and important details from the sources.`

	case "faq":
		return `You are an expert at creating FAQ documents. Based on the following sources, generate a comprehensive FAQ in {format} format.

Sources:
{sources}

Create 10-15 frequently asked questions with detailed answers that cover the main topics and information from the sources.`

	case "study_guide":
		return `You are an expert educator. Create a comprehensive study guide based on the following sources in {format} format.

Sources:
{sources}

The study guide should include:
1. Learning objectives
2. Key concepts and definitions
3. Important themes and topics
4. Study questions and exercises
5. Summary of main points

Format it for {length} study session.`

	case "outline":
		return `You are an expert at creating structured outlines. Create a detailed hierarchical outline based on the following sources in {format} format.

Sources:
{sources}

The outline should:
- Use proper hierarchical structure (I, A, 1, a)
- Cover all main topics and subtopics
- Include brief descriptions for major sections
- Be {length} in detail`

	case "podcast":
		return `You are a podcast script writer. Create an engaging podcast script based on the following sources.

Sources:
{sources}

The script should:
- Be conversational and engaging
- Cover the main topics from the sources
- Include two hosts discussing the material
- Be approximately 10-15 minutes when spoken
- Include natural transitions and questions
- Have a clear introduction and conclusion

Format as a podcast script with speaker labels (Host 1, Host 2) and stage directions in [brackets].`

	case "timeline":
		return `You are an expert at creating chronological timelines. Create a timeline based on the following sources in {format} format.

Sources:
{sources}

Extract and organize events chronologically with:
- Dates or time periods
- Event descriptions
- Key figures involved
- Significance of each event`

	case "glossary":
		return `You are an expert at creating glossaries. Create a comprehensive glossary based on the following sources in {format} format.

Sources:
{sources}

Include:
- Important terms and concepts
- Clear, concise definitions
- Context from the sources
- Cross-references between related terms`

	case "quiz":
		return `You are an educator creating assessment materials. Create a quiz based on the following sources in {format} format.

Sources:
{sources}

The quiz should include:
- A mix of question types (multiple choice, true/false, short answer)
- Questions of varying difficulty
- An answer key
- Questions that test understanding, not just recall

Create {length} quiz with 10-20 questions.`

	case "custom":
		return `You are a helpful assistant. Based on the following sources and the custom request, generate the requested content.

Sources:
{sources}

Custom Request:
{prompt}

Generate the content in {format} format, keeping it {length}.`

	default:
		return `You are a helpful assistant. Based on the following sources, provide a {type} in {format} format.

Sources:
{sources}

Generate {length} content.`

	}
}

// Chat performs a chat query with RAG
func (a *Agent) Chat(ctx context.Context, notebookID, message string, history []ChatMessage) (*ChatResponse, error) {
	// Perform similarity search to find relevant sources
	docs, err := a.vectorStore.SimilaritySearch(ctx, message, a.cfg.MaxSources)
	if err != nil {
		return nil, fmt.Errorf("failed to search documents: %w", err)
	}

	// Build context from retrieved documents
	var contextBuilder strings.Builder
	if len(docs) > 0 {
		contextBuilder.WriteString("Relevant information from sources:\n\n")
		for i, doc := range docs {
			contextBuilder.WriteString(fmt.Sprintf("[Source %d] %s\n", i+1, doc.PageContent))
			if source, ok := doc.Metadata["source"].(string); ok {
				contextBuilder.WriteString(fmt.Sprintf("Source: %s\n\n", source))
			}
		}
	}

	// Build chat history
	var historyBuilder strings.Builder
	for i, msg := range history {
		if i >= 10 { // Limit history
			break
		}
		role := "User"
		if msg.Role == "assistant" {
			role = "Assistant"
		}
		historyBuilder.WriteString(fmt.Sprintf("%s: %s\n", role, msg.Content))
	}

	// Create RAG prompt using f-string format
	promptTemplate := prompts.NewPromptTemplate(
		`You are a helpful AI assistant for a notebook application. Answer the user's question based on the provided context and chat history. If the context doesn't contain enough information, say so and provide a general response.

Chat History:
{history}

Context:
{context}

User Question: {question}

Provide a helpful, accurate response. When referencing information from sources, mention which source it came from.`,
		[]string{"history", "context", "question"},
	)
	promptTemplate.TemplateFormat = prompts.TemplateFormatFString

	promptValue, err := promptTemplate.Format(map[string]any{
		"history":  historyBuilder.String(),
		"context":  contextBuilder.String(),
		"question": message,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to format prompt: %w", err)
	}

	// Generate response
	response, err := llms.GenerateFromSinglePrompt(ctx, a.llm, promptValue)
	if err != nil {
		return nil, fmt.Errorf("failed to generate response: %w", err)
	}

	// Build source summaries
	sourceSummaries := make([]SourceSummary, 0, len(docs))
	sourceMap := make(map[string]bool)
	for _, doc := range docs {
		if source, ok := doc.Metadata["source"].(string); ok {
			if !sourceMap[source] {
				sourceSummaries = append(sourceSummaries, SourceSummary{
					ID:   source,
					Name: source,
					Type: "file",
				})
				sourceMap[source] = true
			}
		}
	}

	return &ChatResponse{
		Message:   response,
		Sources:   sourceSummaries,
		SessionID: notebookID,
		Metadata: map[string]interface{}{
			"docs_retrieved": len(docs),
		},
	}, nil
}

// GeneratePodcastScript generates a podcast script from sources
func (a *Agent) GeneratePodcastScript(ctx context.Context, sources []Source, voice string) (string, error) {
	req := &TransformationRequest{
		Type:   "podcast",
		Length: "medium",
		Format: "markdown",
	}

	resp, err := a.GenerateTransformation(ctx, req, sources)
	if err != nil {
		return "", err
	}

	return resp.Content, nil
}

// GenerateOutline generates an outline from sources
func (a *Agent) GenerateOutline(ctx context.Context, sources []Source) (string, error) {
	req := &TransformationRequest{
		Type:   "outline",
		Length: "detailed",
		Format: "markdown",
	}

	resp, err := a.GenerateTransformation(ctx, req, sources)
	if err != nil {
		return "", err
	}

	return resp.Content, nil
}

// GenerateFAQ generates an FAQ from sources
func (a *Agent) GenerateFAQ(ctx context.Context, sources []Source) (string, error) {
	req := &TransformationRequest{
		Type:   "faq",
		Length: "comprehensive",
		Format: "markdown",
	}

	resp, err := a.GenerateTransformation(ctx, req, sources)
	if err != nil {
		return "", err
	}

	return resp.Content, nil
}

// GenerateStudyGuide generates a study guide from sources
func (a *Agent) GenerateStudyGuide(ctx context.Context, sources []Source) (string, error) {
	req := &TransformationRequest{
		Type:   "study_guide",
		Length: "comprehensive",
		Format: "markdown",
	}

	resp, err := a.GenerateTransformation(ctx, req, sources)
	if err != nil {
		return "", err
	}

	return resp.Content, nil
}

// GenerateSummary generates a summary from sources
func (a *Agent) GenerateSummary(ctx context.Context, sources []Source, length string) (string, error) {
	req := &TransformationRequest{
		Type:   "summary",
		Length: length,
		Format: "markdown",
	}

	resp, err := a.GenerateTransformation(ctx, req, sources)
	if err != nil {
		return "", err
	}

	return resp.Content, nil
}
