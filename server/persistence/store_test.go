package persistence

import (
	"context"
	"errors"
	"sync"
	"testing"
)

type testSource struct {
	value byte
	err   error
}

func (s *testSource) Snapshot() ([]byte, error) { s.value++; return []byte{s.value}, s.err }

type testStore struct {
	values []byte
	err    error
}

func (s *testStore) Load(context.Context) ([]byte, error) { return nil, nil }
func (s *testStore) Close(context.Context) error          { return nil }
func (s *testStore) Save(_ context.Context, data []byte) error {
	s.values = append(s.values, data[0])
	return s.err
}

func TestCoordinatorSerializesCaptureAndCommit(t *testing.T) {
	source := &testSource{}
	store := &testStore{}
	c := NewCoordinator(source, store)
	var wg sync.WaitGroup
	for range 30 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := c.Save(context.Background()); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	for i, v := range store.values {
		if v != byte(i+1) {
			t.Fatal("older snapshot committed after newer snapshot")
		}
	}
	source.err = errors.New("codec failure")
	if err := c.Save(context.Background()); err == nil || len(store.values) != 30 {
		t.Fatal("snapshot failure wrote to store")
	}
	source.err = nil
	store.err = errors.New("database failure")
	if err := c.Save(context.Background()); err == nil {
		t.Fatal("storage failure swallowed")
	}
}
