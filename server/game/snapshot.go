package game

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"webscape/server/game/component"
	"webscape/server/game/model"
	"webscape/server/util"

	"github.com/google/uuid"
)

// gameSnapshot is a storage-independent, versioned save contract. It intentionally
// contains neither client replication baselines nor queued events or connections.
type gameSnapshot struct {
	Version     int                                                           `json:"version"`
	ContentHash string                                                        `json:"contentHash"`
	Tick        uint64                                                        `json:"tick"`
	Entities    map[string]map[component.ComponentId]component.SavedComponent `json:"entities"`
}

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
	s := gameSnapshot{Version: 1, ContentHash: g.world.ContentHash(), Tick: g.currentTick, Entities: map[string]map[component.ComponentId]component.SavedComponent{}}
	add := func(id model.EntityId, c component.Component) error {
		saved, err := component.SaveComponent(c)
		if err != nil {
			return fmt.Errorf("save entity %s: %w", id.String(), err)
		}
		if saved.Version == 0 {
			return nil
		}
		key := id.String()
		if s.Entities[key] == nil {
			s.Entities[key] = map[component.ComponentId]component.SavedComponent{}
		}
		s.Entities[key][c.GetId()] = saved
		return nil
	}
	for _, entities := range g.componentManager.GetAllComponents() {
		for id, c := range entities {
			if err := add(id, c); err != nil {
				return nil, err
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

// RestoreSnapshot validates the entire save before replacing any entities. Authored
// entities are replaced as a set, preserving deletions and preventing duplicate spawns.
// Call only during startup, before installing clients or starting the update loop.
func (g *Game) RestoreSnapshot(data []byte) error {
	g.stateMutex.Lock()
	defer g.stateMutex.Unlock()
	if len(g.clients) != 0 {
		return fmt.Errorf("cannot restore a running game")
	}
	var s gameSnapshot
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&s); err != nil {
		return fmt.Errorf("decode snapshot: %w", err)
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return fmt.Errorf("trailing snapshot data")
	}
	if s.Version != 1 {
		return fmt.Errorf("unsupported snapshot version %d", s.Version)
	}
	if s.ContentHash != g.world.ContentHash() {
		return fmt.Errorf("saved content hash differs from authored game; migrate the save or use a new persistence.worldKey")
	}
	if s.Entities == nil {
		return fmt.Errorf("missing snapshot entities")
	}
	active := map[model.EntityId][]component.Component{}
	offline := map[model.EntityId][]component.Component{}
	itemOwners := map[model.ItemId]model.EntityId{}
	for key, saved := range s.Entities {
		parsed, err := uuid.Parse(key)
		if err != nil || parsed == uuid.Nil || parsed.String() != key {
			return fmt.Errorf("invalid saved entity id %q", key)
		}
		id := model.EntityId(parsed)
		components := make([]component.Component, 0, len(saved))
		for cid, record := range saved {
			c, err := component.RestoreComponent(cid, record)
			if err != nil {
				return fmt.Errorf("restore entity %s component %s: %w", key, cid, err)
			}
			components = append(components, c)
		}
		if err := component.ValidateSavedEntity(components); err != nil {
			return fmt.Errorf("restore entity %s: %w", key, err)
		}
		for _, c := range components {
			items := []*model.Item{}
			switch c := c.(type) {
			case *component.CPosition:
				pos := c.GetPosition()
				coord, _ := g.world.GlobalToChunk(pos.X, pos.Y)
				if !g.world.HasChunk(coord) {
					return fmt.Errorf("saved entity position is outside the authored world")
				}
			case *component.CMetadata:
				metadata := c.GetMetadata().(util.JObject)
				for _, dimension := range []struct {
					name  string
					limit int
				}{{"width", g.world.GetChunkSize().X}, {"height", g.world.GetChunkSize().Y}} {
					if raw, exists := metadata[dimension.name]; exists {
						size, ok := raw.(util.JNumber)
						if !ok || size < 1 || size > util.JNumber(dimension.limit) || size != util.JNumber(int(size)) {
							return fmt.Errorf("invalid saved entity footprint")
						}
					}
				}
			case *component.CInventory:
				items = c.GetAllItems()
			case *component.CEquipped:
				for _, slot := range []model.EquipmentSlot{model.SlotHead, model.SlotChest, model.SlotLegs, model.SlotFeet, model.SlotWeapon, model.SlotOffhand} {
					if item := c.GetEquippedItem(slot); item != nil {
						items = append(items, item)
					}
				}
			case *component.CDroppedItem:
				items = append(items, c.Item)
			case *component.CConversation:
				if _, ok := g.world.GetConversation(c.GetConversationId()); !ok {
					return fmt.Errorf("unknown saved conversation")
				}
			case *component.CQuestLog:
				for _, p := range c.GetActiveProgress() {
					q, ok := g.world.GetQuest(p.QuestId)
					if !ok || p.CurrentStepIndex >= len(q.Steps) || q.Steps[p.CurrentStepIndex].Id != p.StepId {
						return fmt.Errorf("invalid saved quest step")
					}
				}
				for _, q := range c.GetCompletedQuests() {
					if _, ok := g.world.GetQuest(q.QuestId); !ok {
						return fmt.Errorf("unknown completed quest")
					}
				}
			}
			for _, item := range items {
				if _, exists := itemOwners[item.Id]; exists {
					return fmt.Errorf("saved item belongs to multiple entities")
				}
				itemOwners[item.Id] = id
			}
		}
		if _, player := saved[component.ComponentIdPlayer]; player {
			offline[id] = components
		} else {
			active[id] = components
		}
	}
	ids := map[model.EntityId]bool{}
	for _, entities := range g.componentManager.GetAllComponents() {
		for id := range entities {
			ids[id] = true
		}
	}
	for id := range ids {
		g.componentManager.RemoveEntity(id)
	}
	for id, components := range active {
		g.componentManager.SetEntityComponents(id, components...)
	}
	g.offlinePlayers, g.currentTick = offline, s.Tick
	return nil
}
