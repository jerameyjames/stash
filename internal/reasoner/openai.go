package reasoner

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

// OpenAI uses the OpenAI-compatible SDK for reasoning tasks.
// Works with any OpenAI-compatible endpoint: api.openai.com, OpenRouter, local endpoints, etc.
// The model string is passed as-is to the API — no stripping or transformation.
type OpenAI struct {
	client openai.Client
	model  string
	driver string
}

// NewOpenAI creates an OpenAI reasoner.
// baseURL: the API endpoint (e.g. "https://api.openai.com/v1")
// apiKey: the API key for the endpoint
// driver: the driver name (e.g. "openai")
// model: required — the model string for this endpoint (e.g. "gpt-4o-mini")
// Returns error if apiKey or model is empty.
func NewOpenAI(baseURL, apiKey, driver, model string) (*OpenAI, error) {
	if apiKey == "" {
		return nil, errors.New("reasoner: apiKey is required")
	}
	if model == "" {
		return nil, errors.New("reasoner: model is required")
	}
	if driver == "" {
		return nil, errors.New("reasoner: driver is required")
	}

	client := openai.NewClient(
		option.WithBaseURL(baseURL),
		option.WithAPIKey(apiKey),
	)

	return &OpenAI{
		client: client,
		model:  model,
		driver: driver,
	}, nil
}

// Model returns the model string as passed at construction.
func (o *OpenAI) Model() string {
	return o.model
}

// Driver returns the driver name as passed at construction.
func (o *OpenAI) Driver() string {
	return o.driver
}

// Reason synthesizes structured reasoning over the given texts using the OpenAI API.
// Combines all texts into a consolidation prompt and returns the LLM's response.
func (o *OpenAI) Reason(ctx context.Context, texts []string) (string, error) {
	if len(texts) == 0 {
		return "", errors.New("reasoner: texts must not be empty")
	}

	// Build prompt: ask LLM to synthesize events into a fact
	eventsList := strings.Join(texts, "\n- ")
	prompt := fmt.Sprintf(`You are a memory synthesis engine. Given raw observations (events), distill them into a single durable fact.

Events:
- %s

Output a single, declarative fact statement (1–2 sentences). Focus on what is true, not when or how often.
Example: "Mohamed prefers Go for systems programming" (not "Mohamed mentioned Go three times").

Fact:`, eventsList)

	resp, err := o.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model: o.model,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(prompt),
		},
	})
	if err != nil {
		return "", fmt.Errorf("chat.completions call failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", errors.New("reasoner: no response from LLM")
	}

	// Extract synthesized fact from response
	fact := resp.Choices[0].Message.Content
	fact = strings.TrimSpace(fact)

	return fact, nil
}
