package component

import (
	"webscape/server/math"
	"webscape/server/util"
)

const ComponentIdPosition = ComponentId("position")

type CPosition struct {
	Position math.Vec2
}

func (c *CPosition) GetId() ComponentId {
	return ComponentIdPosition
}

func (c *CPosition) Serialize() util.Json {
	return util.JObject(map[string]util.Json{
		"x": util.JNumber(c.Position.X),
		"y": util.JNumber(c.Position.Y),
	})
}
