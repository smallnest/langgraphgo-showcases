package sentiment_model

import (
	"context"
	"fmt"
	"os"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

func getLLM() (llms.Model, error) {
	if os.Getenv("OPENAI_API_KEY") == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY not set")
	}
	opts := []openai.Option{}
	if base := os.Getenv("OPENAI_API_BASE"); base != "" {
		opts = append(opts, openai.WithBaseURL(base))
	}
	if model := os.Getenv("OPENAI_MODEL"); model != "" {
		opts = append(opts, openai.WithModel(model))
	}
	return openai.New(opts...)
}

// AnalyzeSentiment analyzes the sentiment of a text.
func AnalyzeSentiment(ctx context.Context, text string) (string, error) {
	llm, err := getLLM()
	if err != nil {
		return "", err
	}

	prompt := fmt.Sprintf(`Analyze the sentiment of the following text. 
Classify it as "Positive", "Negative", or "Neutral".
Provide a brief explanation (1 sentence).

Text: "%s"

Output format: "Sentiment: [Label] - [Explanation]"`, text)

	completion, err := llms.GenerateFromSinglePrompt(ctx, llm, prompt)
	if err != nil {
		return "", err
	}

	return completion, nil
}
