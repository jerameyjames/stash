// Package reasoner synthesizes structured reasoning over text.
// Implementations: OpenAI (production), Fake (tests).
package reasoner

import (
	"context"
)

// Reasoner synthesizes structured reasoning over text input.
// Implementations: OpenAI (production), Fake (tests).
type Reasoner interface {
	// Reason takes a list of text inputs and returns synthesized reasoning output.
	// Implementation determines how to combine inputs, query the LLM, and format the result.
	Reason(ctx context.Context, texts []string) (string, error)

	// Model returns the model identifier as passed at construction.
	// Examples: "gpt-4o-mini", "gpt-4".
	// Used for logging and debugging.
	Model() string

	// Driver returns the driver name as passed at construction.
	// Examples: "openai".
	Driver() string
}
