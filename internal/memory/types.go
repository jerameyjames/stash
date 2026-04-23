package memory

import (
	"fmt"
	"time"

	"github.com/alash3al/stash/internal/store"
)

// Event represents something that happened at a specific point in time.
// Stored as a store.Record with _memory.type = "event".
type Event struct {
	ID        string
	Namespace string
	Content   string
	Timestamp time.Time
	ExpiresAt *time.Time     // nil = forever, non-nil = expiration
	Metadata  map[string]any
	Score     float32
}

// WorkingMemory represents working memory — what is actively being thought about.
// Single global working memory for MVP, stored with fixed ID "_memory.context".
// Stored as a store.Record with _memory.type = "context".
type WorkingMemory struct {
	ID        string
	Focus     string
	EventIDs  []string
	CreatedAt time.Time
	UpdatedAt time.Time
	ExpiresAt time.Time
}

// Relation represents a directed semantic link between two events.
// Stored as a store.Record with _memory.type = "relationship".
type Relation struct {
	ID           string         // generated UUID
	Namespace    string         // same namespace as linked events
	FromEventID  string         // source event
	ToEventID    string         // target event
	RelationType string         // e.g., "contradicts", "caused_by"
	Metadata     map[string]any // optional caller metadata
	CreatedAt    time.Time
}

// Supported relation types (extensible)
const (
	RelationTypeContradicts = "contradicts" // A contradicts B
	RelationTypeCausedBy    = "caused_by"   // A caused B
	RelationTypeSimilarTo   = "similar_to"  // A is similar to B
	RelationTypeReferences  = "references"  // A references B
)

// BulkRemember represents a single event for batch import.
// Minimal structure: just content, optional metadata and TTL.
type BulkRemember struct {
	Content  string         // required, non-empty
	Metadata map[string]any // optional caller metadata
	TTL      *time.Duration // optional; nil = no expiry
}

// Fact represents a durable, synthesized belief derived from events.
// Stored as a store.Record with _memory.type = "fact".
// Facts are synthesized from clusters of similar events via LLM reasoning.
type Fact struct {
	ID              string         // UUID
	Namespace       string         // same as source events
	Content         string         // the synthesized fact text
	SynthesizedFrom []string       // event IDs used to create this fact
	ConflictWith    []string       // fact IDs with conflicting information (if any)
	CreatedAt       time.Time      // when fact was created
	Metadata        map[string]any // optional caller metadata
	Score           float32        // similarity score for retrieval
}

// FactFromRecord extracts a Fact from a store.Record.
// Returns error if record type is not "fact" or required fields are missing.
func FactFromRecord(r *store.Record) (*Fact, error) {
	// Check type
	memMeta, ok := r.Metadata["_memory"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("record metadata missing _memory field")
	}

	recType, ok := memMeta["type"].(string)
	if !ok || recType != "fact" {
		return nil, fmt.Errorf("record is not a fact (type=%q)", recType)
	}

	// Extract synthesized_from
	synthesizedFrom := []string{}
	if sf, ok := memMeta["synthesized_from"].([]any); ok {
		for _, id := range sf {
			if idStr, ok := id.(string); ok {
				synthesizedFrom = append(synthesizedFrom, idStr)
			}
		}
	}

	// Extract conflict_with
	conflictWith := []string{}
	if cw, ok := memMeta["conflict_with"].([]any); ok {
		for _, id := range cw {
			if idStr, ok := id.(string); ok {
				conflictWith = append(conflictWith, idStr)
			}
		}
	}

	// Extract timestamp
	createdAt := time.Now()
	if ts, ok := memMeta["created_at"].(string); ok {
		if parsed, err := time.Parse(time.RFC3339, ts); err == nil {
			createdAt = parsed
		}
	}

	return &Fact{
		ID:              r.ID,
		Namespace:       r.Namespace,
		Content:         r.Content,
		SynthesizedFrom: synthesizedFrom,
		ConflictWith:    conflictWith,
		CreatedAt:       createdAt,
		Metadata:        r.Metadata,
		Score:           0, // No score for fact record (not from search)
	}, nil
}
