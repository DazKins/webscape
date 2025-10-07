package component

import "webscape/server/math"

const ComponentIdPosition ComponentId = "position"

type CPosition struct {
	Position math.Vec2
}

func (c *CPosition) Update() bool {
	return false
}

func (c *CPosition) GetId() ComponentId {
	return ComponentIdPosition
}

func (c *CPosition) Serialize() map[string]any {
	return map[string]any{
		"x": c.Position.X,
		"y": c.Position.Y,
	}
}
