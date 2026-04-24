package brain

import (
	"context"
	"sort"

	"github.com/alash3al/stash/internal/brain/store"
)

// Recall retrieves memories relevant to a query via semantic search.
// Searches both events and facts, returns unified results ranked by relevance.
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

	// Search facts first (consolidated, higher quality)
	factResults, err := b.store.Search(ctx, store.Query{
		Namespaces: namespaces,
		Vector:     vec,
		VectorName: b.vectorKey(),
		TopK:       limit,
		Filter: &store.Predicate{
			Field: "metadata._memory.type",
			Op:    store.OpEq,
			Value: typeFact,
		},
	})
	if err != nil {
		return nil, err
	}

	// If not enough facts, search events
	remaining := limit - len(factResults)
	var eventResults []store.SearchResult
	if remaining > 0 {
		eventResults, err = b.store.Search(ctx, store.Query{
			Namespaces: namespaces,
			Vector:     vec,
			VectorName: b.vectorKey(),
			TopK:       remaining,
			Filter: &store.Predicate{
				Field: "metadata._memory.type",
				Op:    store.OpEq,
				Value: typeEvent,
			},
		})
		if err != nil {
			return nil, err
		}
	}

	// Combine and sort by score
	allResults := append(factResults, eventResults...)
	sort.Slice(allResults, func(i, j int) bool {
		return allResults[i].Score > allResults[j].Score
	})

	memories := make([]Memory, 0, len(allResults))
	for _, result := range allResults {
		memories = append(memories, memoryFromRecord(result.Record, result.Score))
	}

	return memories, nil
}
