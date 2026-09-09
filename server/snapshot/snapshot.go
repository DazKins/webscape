// Package snapshot defines the storage-neutral envelope shared by the game and
// adapters. Component payloads remain opaque; only components know their schema.
package snapshot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/google/uuid"
)

type Component struct {
	Version int             `json:"version"`
	Data    json.RawMessage `json:"data"`
}
type State struct {
	Version     int                             `json:"version"`
	ContentHash string                          `json:"contentHash"`
	Tick        uint64                          `json:"tick"`
	Entities    map[string]map[string]Component `json:"entities"`
}

func Decode(data []byte) (State, error) {
	var state State
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&state); err != nil {
		return state, fmt.Errorf("decode snapshot: %w", err)
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return state, fmt.Errorf("trailing snapshot data")
	}
	if state.Version < 1 || state.Entities == nil {
		return state, fmt.Errorf("invalid snapshot envelope")
	}
	for id, components := range state.Entities {
		parsed, err := uuid.Parse(id)
		if err != nil || parsed == uuid.Nil || parsed.String() != id || len(components) == 0 {
			return state, fmt.Errorf("invalid saved entity %q", id)
		}
		for key, c := range components {
			if key == "" || c.Version < 1 || !json.Valid(c.Data) || bytes.Equal(bytes.TrimSpace(c.Data), []byte("null")) {
				return state, fmt.Errorf("invalid saved component %q", key)
			}
		}
	}
	return state, nil
}
