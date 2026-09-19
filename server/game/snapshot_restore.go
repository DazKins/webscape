package game

import (
	"fmt"
	"webscape/server/game/collision"
	"webscape/server/game/component"
	"webscape/server/game/model"
	"webscape/server/game/world"
	"webscape/server/math"

	"github.com/google/uuid"
)

func (g *Game) decodeSnapshotPlayers(s gameSnapshot) (map[model.EntityId][]component.Component, error) {
	w := g.world
	checker := collision.Checker{World: w, ComponentManager: g.componentManager}
	result := map[model.EntityId][]component.Component{}
	claims := map[model.ItemId]bool{}
	for key, saved := range s.Players {
		parsed, err := uuid.Parse(key)
		if err != nil || parsed == uuid.Nil || parsed.String() != key {
			return result, fmt.Errorf("invalid saved entity id %q", key)
		}
		id := model.EntityId(parsed)
		components, err := decodeEntityComponents(saved)
		if err != nil {
			return result, fmt.Errorf("restore entity %s: %w", key, err)
		}
		// Older saves predate mana. Migrate players before validating their
		// required components, preserving mana already present in newer saves.
		if _, player := saved[string(component.ComponentIdPlayer)]; player {
			if _, hasMana := saved[string(component.ComponentIdMana)]; !hasMana {
				maximum := w.GetManaSettings().MaxMana
				components = append(components, component.NewCMana(maximum, maximum))
			}
		}
		components = component.IdleComponents(components, 0)
		for _, c := range components {
			if position, ok := c.(*component.CPosition); ok {
				pos := position.GetPosition()
				if checker.IsBlocked(pos.X, pos.Y) {
					position.SetPosition(w.GetPlayerSpawn())
				}
			}
		}
		ctx := newSaveContext(w, components, claims)
		if err := component.ValidateSavedEntity(components, ctx); err != nil {
			return result, fmt.Errorf("restore entity %s: %w", key, err)
		}
		result[id] = components
	}
	return result, nil
}

func decodeEntityComponents(saved map[string]component.SavedComponent) ([]component.Component, error) {
	result := make([]component.Component, 0, len(saved))
	for id, record := range saved {
		c, err := component.RestoreComponent(component.ComponentId(id), record)
		if err != nil {
			return nil, fmt.Errorf("component %s: %w", id, err)
		}
		result = append(result, c)
	}
	// Older player saves predate banking.
	if _, player := saved[string(component.ComponentIdPlayer)]; player {
		if _, bank := saved[string(component.ComponentIdBank)]; !bank {
			result = append(result, component.NewCBank())
		}
	}
	return result, nil
}

// saveContext bridges component-owned validation to authored registries. Only
// cross-entity item uniqueness belongs here; each component validates its fields.
type saveContext struct {
	world      *world.World
	components map[component.ComponentId]component.Component
	itemClaims map[model.ItemId]bool
}

func newSaveContext(w *world.World, components []component.Component, claims map[model.ItemId]bool) *saveContext {
	c := &saveContext{world: w, components: map[component.ComponentId]component.Component{}, itemClaims: claims}
	for _, value := range components {
		c.components[value.GetId()] = value
	}
	return c
}
func (c *saveContext) Component(id component.ComponentId) component.Component {
	return c.components[id]
}
func (c *saveContext) ClaimItem(item *model.Item) error {
	if err := item.ValidateSaved(); err != nil {
		return err
	}
	if c.itemClaims[item.Id] {
		return fmt.Errorf("saved item has multiple owners")
	}
	c.itemClaims[item.Id] = true
	return nil
}
func (c *saveContext) ValidPosition(pos math.Vec2) bool {
	coord, _ := c.world.GlobalToChunk(pos.X, pos.Y)
	return c.world.HasChunk(coord)
}
func (c *saveContext) ChunkSize() math.Vec2 {
	size := c.world.GetChunkSize()
	return math.Vec2{X: size.X, Y: size.Y}
}
func (c *saveContext) HasQuest(id string) bool { _, ok := c.world.GetQuest(id); return ok }
func (c *saveContext) ValidQuestStep(id string, index int, step string) bool {
	q, ok := c.world.GetQuest(id)
	return ok && index >= 0 && index < len(q.Steps) && q.Steps[index].Id == step
}
