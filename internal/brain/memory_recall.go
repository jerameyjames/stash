package brain

import (
	"context"
	"time"

	"github.com/alash3al/stash/internal/brain/store"
)

func (b *Brain) Recall(ctx context.Context, namespace, query string, limit int) ([]Memory, error) {
	if limit == 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	vec, err := b.embedder.Embed(ctx, query)
	if err != nil {
		return nil, err
	}

	var namespaces []string
	if namespace != "" {
		namespaces = []string{namespace}
	}

	results, err := b.store.Search(ctx, store.Query{
		Namespaces: namespaces,
		Vector:     vec,
		VectorName: b.embedder.Model(),
		TopK:       limit,
		Filter: &store.Predicate{
			Field: "metadata._memory.type",
			Op:    store.OpEq,
			Value: typeEvent,
		},
	})
	if err != nil {
		return nil, err
	}

	memories := make([]Memory, 0, len(results))
	for _, result := range results {
		m, err := recordToMemory(result.Record, result.Score)
		if err != nil {
			continue
		}
		memories = append(memories, m)
	}

	return memories, nil
}

func recordToMemory(r store.Record, score float32) (Memory, error) {
	memMeta, ok := r.Metadata["_memory"].(map[string]any)
	if !ok {
		return Memory{}, nil
	}

	var timestamp time.Time
	if ts, ok := memMeta["timestamp"].(string); ok {
		if parsed, err := time.Parse(time.RFC3339, ts); err == nil {
			timestamp = parsed
		}
	}
	if timestamp.IsZero() {
		timestamp = r.CreatedAt
	}

	return Memory{
		ID:        r.ID,
		Namespace: r.Namespace,
		Content:   r.Content,
		Score:     score,
		CreatedAt: timestamp,
	}, nil
}
