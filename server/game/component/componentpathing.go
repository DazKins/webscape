package component

import (
	"webscape/server/game/model"
	"webscape/server/math"
	"webscape/server/util"
)

const ComponentIdPathing = ComponentId("pathing")

type PathingTarget struct {
	EntityId util.Optional[model.EntityId]
	Position util.Optional[math.Vec2]
}

type CPathing struct {
	Target PathingTarget
	Path   *util.Path
}

func (c *CPathing) GetId() ComponentId {
	return ComponentIdPathing
}
