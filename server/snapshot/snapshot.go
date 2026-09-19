// Package snapshot defines the versioned player-save envelope shared by game and storage.
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
	Version int                             `json:"version"`
	Players map[string]map[string]Component `json:"players"`
}

// Decode accepts the current player-only save contract.
func Decode(data []byte) (State, error) {
	var state State
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&state); err != nil {
		return State{}, fmt.Errorf("decode player save: %w", err)
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return State{}, fmt.Errorf("trailing save data")
	}
	if state.Version != 2 {
		return State{}, fmt.Errorf("unsupported player save version %d", state.Version)
	}
	if state.Players == nil {
		return State{}, fmt.Errorf("invalid player save")
	}
	for id, components := range state.Players {
		parsed, err := uuid.Parse(id)
		if err != nil || parsed == uuid.Nil || parsed.String() != id {
			return State{}, fmt.Errorf("invalid player id %q", id)
		}
		if _, ok := components["player"]; !ok {
			return State{}, fmt.Errorf("saved record has no player component")
		}
		for key, c := range components {
			if key == "" || c.Version < 1 || !json.Valid(c.Data) || bytes.Equal(bytes.TrimSpace(c.Data), []byte("null")) {
				return State{}, fmt.Errorf("invalid saved component %q", key)
			}
		}
	}
	return state, nil
}
