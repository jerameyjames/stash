package brain

import (
	"context"
	"time"

	"github.com/alash3al/stash/internal/brain/store"
)

type Report struct {
	Namespace     string    
	TotalMemories int       
	GeneratedAt   time.Time 
}

func (b *Brain) Reflect(ctx context.Context, namespace string) (*Report, error) {
	var namespaces []string
	if namespace != "" {
		namespaces = []string{namespace}
	}

	count, err := b.store.Count(ctx, namespaces, &store.Predicate{
		Field: "metadata._memory.type",
		Op:    store.OpEq,
		Value: typeEvent,
	})
	if err != nil {
		return nil, err
	}

	return &Report{
		Namespace:     namespace,
		TotalMemories: int(count),
		GeneratedAt:   time.Now().UTC(),
	}, nil
}
