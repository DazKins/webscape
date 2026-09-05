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
	target        PathingTarget
	path          *util.Path
	plannedTarget math.Vec2
	plannedRange  int
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
	c.path = nil
}

func (c *CPathing) GetPath() *util.Path {
	return c.path
}

func (c *CPathing) SetPath(path *util.Path) {
	c.path = path
}

func (c *CPathing) HasPlan(target math.Vec2, distance int) bool {
	return c.path != nil && c.path.Size() > 0 && c.plannedTarget == target && c.plannedRange == distance
}
func (c *CPathing) SetPlan(path *util.Path, target math.Vec2, distance int) {
	c.path, c.plannedTarget, c.plannedRange = path, target, distance
}
