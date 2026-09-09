// Package persistence stores opaque snapshots. It imports no game, component,
// command or transport packages; save format ownership remains with the game.
package persistence

import (
	"context"
	"sync"
)

type Store interface {
	// Load returns nil when this world has never been saved.
	Load(context.Context) ([]byte, error)
	Save(context.Context, []byte) error
	Close(context.Context) error
}

type Source interface{ Snapshot() ([]byte, error) }

// Coordinator serializes capture AND save, so concurrent tick/disconnect saves
// cannot commit an older snapshot over a newer one. No game lock covers DB I/O.
type Coordinator struct {
	mu     sync.Mutex
	source Source
	store  Store
}

func NewCoordinator(source Source, store Store) *Coordinator {
	return &Coordinator{source: source, store: store}
}

func (c *Coordinator) Save(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	data, err := c.source.Snapshot()
	if err != nil {
		return err
	}
	return c.store.Save(ctx, data)
}
