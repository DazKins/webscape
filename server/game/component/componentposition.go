package component

import (
	"webscape/server/math"
	"webscape/server/util"
)

const ComponentIdPosition = ComponentId("position")

type CPosition struct {
	position math.Vec2
}

func NewCPosition(position math.Vec2) *CPosition {
	return &CPosition{
		position: position,
	}
}

func (c *CPosition) GetId() ComponentId {
	return ComponentIdPosition
}

func (c *CPosition) Serialize() util.Json {
	return util.JObject(map[string]util.Json{
		"x": util.JNumber(c.position.X),
		"y": util.JNumber(c.position.Y),
	})
}

func (c *CPosition) GetPosition() math.Vec2 {
	return c.position
}

func (c *CPosition) SetPosition(position math.Vec2) {
	c.position = position
}
