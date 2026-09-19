package game

import (
	"encoding/json"
	"fmt"

	"webscape/server/game/component"
	"webscape/server/game/model"
	"webscape/server/snapshot"
)

// gameSnapshot is a storage-independent, versioned save contract. It intentionally
// contains neither client replication baselines nor queued events or connections.
type gameSnapshot = snapshot.State

// RetainOfflinePlayers is configured before the game starts. Offline components
// remain outside the ECS, so systems cannot move, damage or reward absent players.
func (g *Game) RetainOfflinePlayers() {
	g.offlinePlayers = make(map[model.EntityId][]component.Component)
}

// Snapshot copies a consistent state under the game mutex. Database I/O happens
// later, outside this boundary, through a runtime-owned storage adapter.
func (g *Game) Snapshot() ([]byte, error) {
	g.stateMutex.Lock()
	defer g.stateMutex.Unlock()
	s := gameSnapshot{Version: 2, Players: map[string]map[string]component.SavedComponent{}}
	add := func(id model.EntityId, c component.Component) error {
		saved, err := component.SaveComponent(c)
		if err != nil {
			return fmt.Errorf("save entity %s: %w", id.String(), err)
		}
		if saved.Version == 0 {
			return nil
		}
		key := id.String()
		if s.Players[key] == nil {
			s.Players[key] = map[string]component.SavedComponent{}
		}
		s.Players[key][string(c.GetId())] = saved
		return nil
	}
	for id := range g.componentManager.GetComponent(component.ComponentIdPlayer) {
		for _, entities := range g.componentManager.GetAllComponents() {
			if c := entities[id]; c != nil {
				if err := add(id, c); err != nil {
					return nil, err
				}
			}
		}
	}

	for id, components := range g.offlinePlayers {
		for _, c := range components {
			if err := add(id, c); err != nil {
				return nil, err
			}
		}
	}
	return json.Marshal(s)
}

// RestoreSnapshot validates every player before installing any. Authored
// world entities remain freshly authored. Only offline players are installed.
// Call only during startup, before installing clients or starting the update loop.
func (g *Game) RestoreSnapshot(data []byte) error {
	g.stateMutex.Lock()
	defer g.stateMutex.Unlock()
	if len(g.clients) != 0 {
		return fmt.Errorf("cannot restore a running game")
	}
	s, err := snapshot.Decode(data)
	if err != nil {
		return err
	}
	if s.Version != 2 {
		return fmt.Errorf("unsupported snapshot version %d", s.Version)
	}
	entities, err := g.decodeSnapshotPlayers(s)
	if err != nil {
		return err
	}
	g.offlinePlayers = entities
	return nil
}
