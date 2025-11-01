package component

import "webscape/server/game/component/componentheader"

const ComponentIdRandomWalk = ComponentId("randomwalk")

type CRandomWalk struct {
	componentheader.ComponentHeader
	walkTimer int
}

func (c *CRandomWalk) GetId() ComponentId {
	return ComponentIdRandomWalk
}

func (c *CRandomWalk) GetWalkTimer() int {
	return c.walkTimer
}

func (c *CRandomWalk) SetWalkTimer(walkTimer int) {
	c.walkTimer = walkTimer
	c.Update()
}
