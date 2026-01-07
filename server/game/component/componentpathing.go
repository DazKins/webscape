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
	target PathingTarget
	path   *util.Path
}

func NewCPathing(target PathingTarget) *CPathing {
	return &CPathing{
		target: target,
		path:   nil,
	}
}

func (c *CPathing) GetId() ComponentId {
	return ComponentIdPathing
}

func (c *CPathing) GetTarget() PathingTarget {
	return c.target
}

func (c *CPathing) SetTarget(target PathingTarget) {
	c.target = target
}

func (c *CPathing) GetPath() *util.Path {
	return c.path
}

func (c *CPathing) SetPath(path *util.Path) {
	c.path = path
}
