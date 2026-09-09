package component

import (
	"fmt"
	"webscape/server/math"
)

type savedPosition struct {
	Position math.Vec2 `json:"position"`
}

func (c *CPosition) Save() (SavedComponent, error) {
	return marshalSaved(savedPosition{Position: c.position})
}
func init() { registerComponentRestore(ComponentIdPosition, 1, restorePosition) }
func restorePosition(saved SavedComponent) (Component, error) {
	var s savedPosition
	if err := decodeSaved(saved.Data, &s); err != nil {
		return nil, err
	}
	return &CPosition{position: s.Position}, nil
}

func (c *CPosition) ValidateSaved(ctx SaveContext) error {
	if !ctx.ValidPosition(c.position) {
		return fmt.Errorf("saved entity position is outside the authored world")
	}
	return nil
}
