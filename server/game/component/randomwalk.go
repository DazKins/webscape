package component

const ComponentIdRandomWalk = ComponentId("randomwalk")

type CRandomWalk struct {
	WalkTimer int
}

func (c *CRandomWalk) GetId() ComponentId {
	return ComponentIdRandomWalk
}
