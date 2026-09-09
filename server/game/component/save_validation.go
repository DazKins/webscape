package component

import (
	"fmt"
	"webscape/server/game/model"
	"webscape/server/math"
)

// SaveContext supplies read-only world/entity queries and an item-ownership
// check without coupling component invariants to Game or a database adapter.
type SaveContext interface {
	Component(ComponentId) Component
	ClaimItem(*model.Item) error
	ValidPosition(math.Vec2) bool
	ChunkSize() math.Vec2
	HasConversation(string) bool
	HasQuest(string) bool
	ValidQuestStep(string, int, string) bool
}

type SavedValidator interface{ ValidateSaved(SaveContext) error }

func ValidateSavedEntity(components []Component, ctx SaveContext) error {
	if ctx.Component(ComponentIdPosition) == nil {
		return fmt.Errorf("entity has no position")
	}
	for _, c := range components {
		if validator, ok := c.(SavedValidator); ok {
			if err := validator.ValidateSaved(ctx); err != nil {
				return fmt.Errorf("component %s: %w", c.GetId(), err)
			}
		}
	}
	return nil
}
