package component

const ComponentIdRandomWalk = ComponentId("randomwalk")

type CRandomWalk struct {
	walkTimer int
}

func NewCRandomWalk(walkTimer int) *CRandomWalk {
	return &CRandomWalk{
		walkTimer: walkTimer,
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
