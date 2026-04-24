// Package reasoner synthesizes structured reasoning over text.
// Implementations: OpenAI (production), Fake (tests).
package reasoner

import (
	"context"
)

// StructuredFact represents an extracted fact with entity, property, and value.
type StructuredFact struct {
	// Entity is the subject (e.g., "Alice", "Bob").
	Entity string
	// Property is the attribute or predicate (e.g., "role", "location").
	Property string
	// Value is the fact value (e.g., "engineer", "Paris").
	Value string
	// Summary is the full natural language fact statement.
	Summary string
}

// StructuredRelationship represents an extracted relationship between two entities.
type StructuredRelationship struct {
	// FromEntity is the source entity (e.g., "Alice").
	FromEntity string
	// RelationType is the relationship type (e.g., "works_at", "located_in").
	RelationType string
	// ToEntity is the target entity (e.g., "TechCorp").
	ToEntity string
	// Confidence is how confident the LLM is in this relationship (0.0-1.0).
	Confidence float32
}

// Reasoner synthesizes structured reasoning over text input.
type Reasoner interface {
	// ReasonStructured takes a list of text inputs and returns a structured fact.
	ReasonStructured(ctx context.Context, texts []string) (*StructuredFact, error)

	// ReasonRelationships takes a fact and extracts relationships between entities.
	ReasonRelationships(ctx context.Context, factContent string) ([]*StructuredRelationship, error)
}
