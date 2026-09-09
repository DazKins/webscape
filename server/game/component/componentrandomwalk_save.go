package component

import (
	"webscape/server/math"
)

type savedRandomWalk struct {
	WalkTimer   int       `json:"walkTimer"`
	MaxDistance int       `json:"maxDistance"`
	Origin      math.Vec2 `json:"origin"`
	HasOrigin   bool      `json:"hasOrigin"`
}

func (c *CRandomWalk) Save() (SavedComponent, error) {
	return marshalSaved(savedRandomWalk{WalkTimer: c.walkTimer, MaxDistance: c.maxDistance, Origin: c.origin, HasOrigin: c.hasOrigin})
}
func init() { registerComponentRestore(ComponentIdRandomWalk, 1, restoreRandomWalk) }
func restoreRandomWalk(saved SavedComponent) (Component, error) {
	var s savedRandomWalk
	if err := decodeSaved(saved.Data, &s); err != nil {
		return nil, err
	}
	return &CRandomWalk{walkTimer: s.WalkTimer, maxDistance: s.MaxDistance, origin: s.Origin, hasOrigin: s.HasOrigin}, nil
}
