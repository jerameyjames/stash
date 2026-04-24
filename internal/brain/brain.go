package brain

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/alash3al/stash/internal/brain/store"
	"github.com/alash3al/stash/internal/embedder"
	"github.com/alash3al/stash/internal/reasoner"
	"github.com/google/uuid"
)

var (
	errMissingStore    = errors.New("brain: store is required")
	errMissingEmbedder = errors.New("brain: embedder is required")
	errMissingReasoner = errors.New("brain: reasoner is required")
)

// Brain is the agent's memory system.
type Brain struct {
	store      store.Store
	embedder   embedder.Embedder
	reasoner   reasoner.Reasoner
	pipelineCh chan string

	// Pipeline stats
	statsLock              sync.RWMutex
	pipelineQueueDepth     int
	pipelineLastRun        time.Time
	pipelineLastRunSuccess bool
	pipelineLastError      string
}

// New creates a Brain with the provided store, embedder, and reasoner.
func New(s store.Store, e embedder.Embedder, r reasoner.Reasoner) (*Brain, error) {
	if s == nil {
		return nil, errMissingStore
	}
	if e == nil {
		return nil, errMissingEmbedder
	}
	if r == nil {
		return nil, errMissingReasoner
	}
	return &Brain{
		store:      s,
		embedder:   e,
		reasoner:   r,
		pipelineCh: make(chan string, 100),
	}, nil
}

// Close releases resources.
func (b *Brain) Close() error {
	return b.store.Close()
}

// PipelineStats returns current pipeline statistics.
func (b *Brain) PipelineStats() (queueDepth int, lastRun time.Time, lastSuccess bool, lastError string) {
	b.statsLock.RLock()
	defer b.statsLock.RUnlock()
	return b.pipelineQueueDepth, b.pipelineLastRun, b.pipelineLastRunSuccess, b.pipelineLastError
}

func (b *Brain) updatePipelineStats(queueDepth int, success bool, err error) {
	b.statsLock.Lock()
	defer b.statsLock.Unlock()
	b.pipelineQueueDepth = queueDepth
	b.pipelineLastRun = time.Now()
	b.pipelineLastRunSuccess = success
	if err != nil {
		b.pipelineLastError = err.Error()
	} else {
		b.pipelineLastError = ""
	}
}

// vectorKey returns the key used for storing vectors.
func (b *Brain) vectorKey() string {
	return b.embedder.Model()
}

// calculateConfidence computes confidence from observation count.
func calculateConfidence(observationCount int) float32 {
	if observationCount == 0 {
		return 0.0
	}
	return float32(observationCount) / float32(observationCount+2)
}

// --- Internal helpers for consolidation ---

// queryRecentEventRecords returns event records from the last 7 days.
func (b *Brain) queryRecentEventRecords(ctx context.Context, namespace string, since time.Time) ([]store.Record, error) {
	var namespaces []string
	if namespace != "" {
		namespaces = []string{namespace}
	}

	records, err := b.store.List(ctx, store.Filter{
		Namespaces: namespaces,
		Where: &store.Predicate{
			And: []store.Predicate{
				{Field: "metadata._memory.type", Op: store.OpEq, Value: typeEvent},
				{Field: "metadata._memory.timestamp", Op: store.OpGte, Value: since.Format(time.RFC3339)},
			},
		},
		Limit: 1000,
	})
	if err != nil {
		return nil, err
	}
	return records, nil
}

// clusterRecordsBySimilarity groups records by cosine similarity.
func (b *Brain) clusterRecordsBySimilarity(records []store.Record, threshold float64) [][]store.Record {
	if len(records) == 0 {
		return nil
	}

	clusters := make([][]store.Record, 0)
	used := make(map[int]bool)

	for i, r1 := range records {
		if used[i] {
			continue
		}

		cluster := []store.Record{r1}
		used[i] = true

		vec1, ok1 := r1.Vectors[b.vectorKey()]
		if !ok1 {
			clusters = append(clusters, cluster)
			continue
		}

		for j := i + 1; j < len(records); j++ {
			if used[j] {
				continue
			}

			vec2, ok2 := records[j].Vectors[b.vectorKey()]
			if !ok2 {
				continue
			}

			sim := cosineSimilarity(vec1.Values, vec2.Values)
			if sim >= threshold {
				cluster = append(cluster, records[j])
				used[j] = true
			}
		}

		clusters = append(clusters, cluster)
	}

	return clusters
}

// cosineSimilarity computes cosine similarity between two vectors.
func cosineSimilarity(a, b []float32) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}

	var dot, normA, normB float64
	for i := 0; i < len(a); i++ {
		dot += float64(a[i]) * float64(b[i])
		normA += float64(a[i]) * float64(a[i])
		normB += float64(b[i]) * float64(b[i])
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}

// storeFact stores a consolidated fact in the store.
func (b *Brain) storeFact(ctx context.Context, namespace, content, factType string, observationCount int, source string, synthesizedFrom []string) error {
	now := time.Now().UTC()
	factID := uuid.New().String()

	vec, err := b.embedder.Embed(ctx, content)
	if err != nil {
		return fmt.Errorf("embed fact: %w", err)
	}

	confidence := calculateConfidence(observationCount)

	memMeta := map[string]any{
		"type":               typeFact,
		"fact_type":          factType,
		"confidence":         float64(confidence),
		"observation_count":  observationCount,
		"source":             source,
		"synthesized_from":   synthesizedFrom,
		"created_at":         now.Format(time.RFC3339),
		"valid_from":         now.Format(time.RFC3339),
	}

	record := store.Record{
		ID:        factID,
		Namespace: namespace,
		Content:   content,
		Vectors: map[string]store.Vector{
			b.vectorKey(): {
				Values: vec,
				Model:  b.embedder.Model(),
			},
		},
		Metadata: map[string]any{
			"_memory": memMeta,
		},
	}

	return b.store.Put(ctx, record)
}

// storeRelationship stores a relationship in the store.
func (b *Brain) storeRelationship(ctx context.Context, namespace, fromEntity, relationType, toEntity, source string, confidence float32) error {
	now := time.Now().UTC()
	relID := uuid.New().String()

	memMeta := map[string]any{
		"type":              typeRelationship,
		"from_entity":       fromEntity,
		"relationship_type": relationType,
		"to_entity":         toEntity,
		"confidence":        float64(confidence),
		"source":            source,
		"created_at":        now.Format(time.RFC3339),
	}

	record := store.Record{
		ID:        relID,
		Namespace: namespace,
		Content:   fmt.Sprintf("%s %s %s", fromEntity, relationType, toEntity),
		Metadata: map[string]any{
			"_memory": memMeta,
		},
	}

	return b.store.Put(ctx, record)
}

// queryFacts returns all facts in a namespace.
func (b *Brain) queryFacts(ctx context.Context, namespace string) ([]Fact, error) {
	var namespaces []string
	if namespace != "" {
		namespaces = []string{namespace}
	}

	records, err := b.store.List(ctx, store.Filter{
		Namespaces: namespaces,
		Where: &store.Predicate{
			Field: "metadata._memory.type",
			Op:    store.OpEq,
			Value: typeFact,
		},
		Limit: 10000,
	})
	if err != nil {
		return nil, err
	}

	facts := make([]Fact, 0, len(records))
	for _, r := range records {
		f, err := factFromRecord(r)
		if err != nil {
			continue
		}
		facts = append(facts, *f)
	}

	return facts, nil
}

// queryRelationships returns all relationships in a namespace.
func (b *Brain) queryRelationships(ctx context.Context, namespace string) ([]Relationship, error) {
	var namespaces []string
	if namespace != "" {
		namespaces = []string{namespace}
	}

	records, err := b.store.List(ctx, store.Filter{
		Namespaces: namespaces,
		Where: &store.Predicate{
			Field: "metadata._memory.type",
			Op:    store.OpEq,
			Value: typeRelationship,
		},
		Limit: 10000,
	})
	if err != nil {
		return nil, err
	}

	relationships := make([]Relationship, 0, len(records))
	for _, r := range records {
		rel, err := relationshipFromRecord(r)
		if err != nil {
			continue
		}
		relationships = append(relationships, *rel)
	}

	return relationships, nil
}

// queryEvents returns all events in a namespace.
func (b *Brain) queryEvents(ctx context.Context, namespace string) ([]store.Record, error) {
	var namespaces []string
	if namespace != "" {
		namespaces = []string{namespace}
	}

	return b.store.List(ctx, store.Filter{
		Namespaces: namespaces,
		Where: &store.Predicate{
			Field: "metadata._memory.type",
			Op:    store.OpEq,
			Value: typeEvent,
		},
		Limit: 10000,
	})
}
