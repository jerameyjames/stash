package brain

import (
	"context"
	"time"
)

func (b *Brain) Run(ctx context.Context) error {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case ns := <-b.pipelineCh:
			select {
			case <-b.pipelineCh:
			case <-time.After(10 * time.Second):
			}
			b.consolidate(ctx, ns)
		case <-ticker.C:
			b.purgeExpired(ctx)
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (b *Brain) consolidate(ctx context.Context, namespace string) {
}

func (b *Brain) purgeExpired(ctx context.Context) {
}
