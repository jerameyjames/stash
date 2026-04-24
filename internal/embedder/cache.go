package embedder

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"sync"
)

// Cached wraps an embedder with an in-memory LRU cache.
type Cached struct {
	embedder Embedder
	cache    map[string][]float32
	mu       sync.RWMutex
	maxSize  int
}

// NewCached creates a cached embedder with the given max cache size.
func NewCached(e Embedder, maxSize int) *Cached {
	if maxSize <= 0 {
		maxSize = 10000 // Default: 10k embeddings
	}
	return &Cached{
		embedder: e,
		cache:    make(map[string][]float32, maxSize),
		maxSize:  maxSize,
	}
}

// Embed returns cached embedding or calls the underlying embedder.
func (c *Cached) Embed(ctx context.Context, text string) ([]float32, error) {
	key := cacheKey(text)

	// Try cache first
	c.mu.RLock()
	if vec, ok := c.cache[key]; ok {
		c.mu.RUnlock()
		return vec, nil
	}
	c.mu.RUnlock()

	// Cache miss — call underlying embedder
	vec, err := c.embedder.Embed(ctx, text)
	if err != nil {
		return nil, err
	}

	// Store in cache
	c.mu.Lock()
	if len(c.cache) >= c.maxSize {
		// Simple eviction: random delete (could be LRU)
		for k := range c.cache {
			delete(c.cache, k)
			break
		}
	}
	c.cache[key] = vec
	c.mu.Unlock()

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
