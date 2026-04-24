package brain

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Contradict finds potential contradictions in facts.
// Compares facts with the same entity+property but different values.
func (b *Brain) Contradict(ctx context.Context, namespace string) ([]Contradiction, error) {
	facts, err := b.queryFacts(ctx, namespace)
	if err != nil {
		return nil, fmt.Errorf("query facts: %w", err)
	}

	// Group facts by entity+property
	groups := make(map[string][]Fact)
	for _, fact := range facts {
		entity, _ := fact.Metadata["entity"].(string)
		property, _ := fact.Metadata["property"].(string)
		if entity == "" || property == "" {
			continue
		}
		key := entity + "|" + property
		groups[key] = append(groups[key], fact)
	}

	var contradictions []Contradiction
	for _, group := range groups {
		if len(group) < 2 {
			continue
		}

		entity, _ := group[0].Metadata["entity"].(string)
		property, _ := group[0].Metadata["property"].(string)

		// Compare each pair
		for i := 0; i < len(group); i++ {
			for j := i + 1; j < len(group); j++ {
				value1, _ := group[i].Metadata["value"].(string)
				value2, _ := group[j].Metadata["value"].(string)

				if value1 == "" || value2 == "" || value1 == value2 {
					continue
				}

				// Determine status: evolution (sequential) vs conflict (overlapping)
				status := "evolution"
				if timeRangesOverlap(
					group[i].ValidFrom, nil,
					group[j].ValidFrom, nil,
					time.Now(),
				) {
					status = "conflict"
				}

				contradictions = append(contradictions, Contradiction{
					ID:           uuid.New().String(),
					FactID1:      group[i].ID,
					FactID2:      group[j].ID,
					Entity:       entity,
					Property:     property,
					Value1:       value1,
					Value2:       value2,
					Status:       status,
					DiscoveredAt: time.Now().UTC(),
				})
			}
		}
	}

	return contradictions, nil
}

// timeRangesOverlap checks if two temporal ranges overlap.
func timeRangesOverlap(from1 time.Time, until1 *time.Time, from2 time.Time, until2 *time.Time, now time.Time) bool {
	effectiveUntil1 := until1
	if effectiveUntil1 == nil {
		effectiveUntil1 = &now
	}

	effectiveUntil2 := until2
	if effectiveUntil2 == nil {
		effectiveUntil2 = &now
	}

	return from1.Before(*effectiveUntil2) && from2.Before(*effectiveUntil1)
}
