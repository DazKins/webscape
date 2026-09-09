package component

import (
	"fmt"
	"webscape/server/game/model"
	"webscape/server/math"
)

type savedSpawn struct {
	Position       math.Vec2       `json:"position"`
	RespawnTicks   int             `json:"respawnTicks"`
	RemainingTicks int             `json:"remainingTicks"`
	TemplateID     string          `json:"templateId"`
	Template       map[string]any  `json:"template"`
	ChildID        *model.EntityId `json:"childId"`
	HasSpawned     bool            `json:"hasSpawned"`
}

func (c *CSpawn) Save() (SavedComponent, error) {
	s := savedSpawn{Position: c.spawnPosition, RespawnTicks: c.respawnTicks, RemainingTicks: c.remainingRespawnTicks, TemplateID: c.childTemplateEntityId, Template: c.childTemplateComponents, HasSpawned: c.hasSpawned}
	if c.childEntityId.IsPresent() {
		id := c.childEntityId.Unwrap()
		s.ChildID = &id
	}
	return marshalSaved(s)
}
func init() { registerComponentRestore(ComponentIdSpawn, 1, restoreSpawn) }
func restoreSpawn(saved SavedComponent) (Component, error) {
	var s savedSpawn
	if err := decodeSaved(saved.Data, &s); err != nil {
		return nil, err
	}
	c := NewCSpawn(s.Position, s.RespawnTicks, s.TemplateID, s.Template)
	c.remainingRespawnTicks, c.hasSpawned = s.RemainingTicks, s.HasSpawned
	if s.ChildID != nil {
		c.SetChildEntityId(*s.ChildID)
	}
	return c, nil
}

func (c *CSpawn) ValidateSaved(ctx SaveContext) error {
	if c.respawnTicks < 0 || c.remainingRespawnTicks < 0 || c.childTemplateComponents == nil {
		return fmt.Errorf("invalid spawn state")
	}
	return nil
}
