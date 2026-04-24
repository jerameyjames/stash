package brain

import (
	"errors"

	"github.com/alash3al/stash/internal/brain/store"
	"github.com/alash3al/stash/internal/embedder"
	"github.com/alash3al/stash/internal/reasoner"
)

var (
	errMissingStore    = errors.New("brain: store is required")
	errMissingEmbedder = errors.New("brain: embedder is required")
	errMissingReasoner = errors.New("brain: reasoner is required")
)

type Brain struct {
	store      store.Store
	embedder   embedder.Embedder
	reasoner   reasoner.Reasoner
	pipelineCh chan string
}

func New(s store.Store, e embedder.Embedder, r reasoner.Reasoner) (*Brain, error) {
	if s == nil {
		return nil, errMissingStore
	}
	if e == nil {
		return nil, errMissingEmbedder
	}
	if r == nil {
		return nil, errMissingReasoner
	}
	return &Brain{
		store:      s,
		embedder:   e,
		reasoner:   r,
		pipelineCh: make(chan string, 100),
	}, nil
}

func (b *Brain) Close() error {
	return b.store.Close()
}
