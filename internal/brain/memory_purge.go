package brain

import "context"

func (b *Brain) Purge(ctx context.Context, id string) error {
	return b.store.Purge(ctx, id)
}
