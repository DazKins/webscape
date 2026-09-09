package game

import (
	"fmt"
	"webscape/server/game/component"
	"webscape/server/game/model"
	"webscape/server/game/world"
	"webscape/server/math"

	"github.com/google/uuid"
)

// restoredEntities is built and validated without changing the live manager.
// Installation is a separate final step, so invalid saves cannot partially load.
type restoredEntities struct {
	active  map[model.EntityId][]component.Component
	offline map[model.EntityId][]component.Component
}

func decodeSnapshotEntities(s gameSnapshot, w *world.World) (restoredEntities, error) {
	result := restoredEntities{active: map[model.EntityId][]component.Component{}, offline: map[model.EntityId][]component.Component{}}
	claims := map[model.ItemId]bool{}
	for key, saved := range s.Entities {
		parsed, err := uuid.Parse(key)
		if err != nil || parsed == uuid.Nil || parsed.String() != key {
			return result, fmt.Errorf("invalid saved entity id %q", key)
		}
		id := model.EntityId(parsed)
		components, err := decodeEntityComponents(saved)
		if err != nil {
			return result, fmt.Errorf("restore entity %s: %w", key, err)
		}
		ctx := newSaveContext(w, components, claims)
		if err := component.ValidateSavedEntity(components, ctx); err != nil {
			return result, fmt.Errorf("restore entity %s: %w", key, err)
		}
		if ctx.Component(component.ComponentIdPlayer) != nil {
			result.offline[id] = components
		} else {
			result.active[id] = components
		}
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
	return result, nil
}

func (g *Game) installSnapshot(s gameSnapshot, entities restoredEntities) {
	ids := map[model.EntityId]bool{}
	for _, components := range g.componentManager.GetAllComponents() {
		for id := range components {
			ids[id] = true
		}
	}
	for id := range ids {
		g.componentManager.RemoveEntity(id)
	}
	for id, components := range entities.active {
		g.componentManager.SetEntityComponents(id, components...)
	}
	g.offlinePlayers, g.currentTick = entities.offline, s.Tick
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
func (c *saveContext) HasConversation(id string) bool {
	_, ok := c.world.GetConversation(id)
	return ok
}
func (c *saveContext) HasQuest(id string) bool { _, ok := c.world.GetQuest(id); return ok }
func (c *saveContext) ValidQuestStep(id string, index int, step string) bool {
	q, ok := c.world.GetQuest(id)
	return ok && index >= 0 && index < len(q.Steps) && q.Steps[index].Id == step
}
