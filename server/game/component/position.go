package component

import (
	"webscape/server/game/componentheader"
	"webscape/server/math"
)

const ComponentIdPosition = ComponentId("position")

type CPosition struct {
	componentheader.ComponentHeader
	position math.Vec2
}

func (c *CPosition) GetId() ComponentId {
	return ComponentIdPosition
}

func (c *CPosition) Serialize() map[string]any {
	return map[string]any{
		"x": c.position.X,
		"y": c.position.Y,
	}
}

func (c *CPosition) GetPosition() math.Vec2 {
	return c.position
}

func (c *CPosition) SetPosition(position math.Vec2) {
	c.position = position
	c.Update()
}
