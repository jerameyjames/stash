package brain

import (
	"context"
	"fmt"
)

func (b *Brain) Forget(ctx context.Context, namespace, query string) error {
	memories, err := b.Recall(ctx, namespace, query, 1)
	if err != nil {
		return err
	}
	if len(memories) == 0 {
		return fmt.Errorf("forget: %w", ErrNotFound)
	}

	return b.store.Delete(ctx, memories[0].ID)
}
