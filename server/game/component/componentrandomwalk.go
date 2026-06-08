package component

import "webscape/server/math"

const ComponentIdRandomWalk = ComponentId("randomwalk")

type CRandomWalk struct {
	walkTimer   int
	maxDistance int
	origin      math.Vec2
	hasOrigin   bool
}

func NewCRandomWalk(walkTimer int, maxDistance int) *CRandomWalk {
	if maxDistance < 0 {
		maxDistance = 0
	}
	return &CRandomWalk{
		walkTimer:   walkTimer,
		maxDistance: maxDistance,
	}
}

func (c *CRandomWalk) GetId() ComponentId {
	return ComponentIdRandomWalk
}

func (c *CRandomWalk) GetWalkTimer() int {
	return c.walkTimer
}

func (c *CRandomWalk) SetWalkTimer(walkTimer int) {
	c.walkTimer = walkTimer
}

func (c *CRandomWalk) GetMaxDistance() int {
	return c.maxDistance
}

func (c *CRandomWalk) SetMaxDistance(maxDistance int) {
	if maxDistance < 0 {
		maxDistance = 0
	}
	c.maxDistance = maxDistance
}

func (c *CRandomWalk) HasOrigin() bool {
	return c.hasOrigin
}

func (c *CRandomWalk) GetOrigin() math.Vec2 {
	return c.origin
}

func (c *CRandomWalk) SetOrigin(origin math.Vec2) {
	c.origin = origin
	c.hasOrigin = true
}
