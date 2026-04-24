package brain

import (
	"context"
	"fmt"
	"time"

	"github.com/alash3al/stash/internal/brain/store"
	"github.com/google/uuid"
)

type Memory struct {
	ID        string    
	Namespace string    
	Content   string    
	Score     float32   
	CreatedAt time.Time 
}

const typeEvent = "event"

func (b *Brain) Remember(ctx context.Context, namespace, content string, metadata map[string]any) (string, error) {
	if content == "" {
		return "", ErrEmptyContent
	}
	if err := validateMetadata(metadata); err != nil {
		return "", err
	}

	vec, err := b.embedder.Embed(ctx, content)
	if err != nil {
		return "", err
	}

	eventID := uuid.New().String()
	now := time.Now().UTC()

	memMeta := map[string]any{
		"type":      typeEvent,
		"content":   content,
		"timestamp": now.Format(time.RFC3339),
	}

	recordMeta := map[string]any{
		"_memory": memMeta,
	}
	for k, v := range metadata {
		recordMeta[k] = v
	}

	record := store.Record{
		ID:        eventID,
		Namespace: namespace,
		Content:   content,
		Vectors: map[string]store.Vector{
			b.embedder.Model(): {
				Values: vec,
				Model:  b.embedder.Model(),
			},
		},
		Metadata: recordMeta,
	}

	if err := b.store.Put(ctx, record); err != nil {
		return "", err
	}

	select {
	case b.pipelineCh <- namespace:
	default:
	}

	return eventID, nil
}

func validateMetadata(metadata map[string]any) error {
	for k := range metadata {
		if len(k) > 0 && k[0] == '_' {
			return fmt.Errorf("metadata key %q: must not start with underscore", k)
		}
	}
	return nil
}
