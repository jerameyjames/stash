package embedder

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
)

// Cached stores embeddings in a generic store using a special namespace.
// Composes with store, doesn't extend it.
type Cached struct {
	embedder  Embedder
	getRecord func(ctx context.Context, id string) (map[string][]float32, error)
	putRecord func(ctx context.Context, id string, text string, vector []float32, model string) error
}

// NewCached creates a cached embedder using store operations.
func NewCached(
	e Embedder,
	getRecord func(ctx context.Context, id string) (map[string][]float32, error),
	putRecord func(ctx context.Context, id string, text string, vector []float32, model string) error,
) *Cached {
	return &Cached{
		embedder:  e,
		getRecord: getRecord,
		putRecord: putRecord,
	}
}

// Embed returns cached embedding or calls the underlying embedder.
func (c *Cached) Embed(ctx context.Context, text string) ([]float32, error) {
	hash := cacheKey(text)

	// Try cache first
	vectors, err := c.getRecord(ctx, hash)
	if err == nil && vectors != nil {
		if vec, ok := vectors[c.embedder.Model()]; ok && len(vec) > 0 {
			return vec, nil
		}
	}

	// Cache miss — call underlying embedder
	vec, err := c.embedder.Embed(ctx, text)
	if err != nil {
		return nil, err
	}

	// Store in cache (fire and forget)
	go func() {
		c.putRecord(context.Background(), hash, text, vec, c.embedder.Model())
	}()

	return vec, nil
}

// Model returns the underlying embedder's model.
func (c *Cached) Model() string {
	return c.embedder.Model()
}

// Dims returns the underlying embedder's dimensions.
func (c *Cached) Dims() int {
	return c.embedder.Dims()
}

func cacheKey(text string) string {
	hash := sha256.Sum256([]byte(text))
	return hex.EncodeToString(hash[:])
}
