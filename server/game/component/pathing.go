package component

import (
	"webscape/server/game/componentheader"
	"webscape/server/util"
)

const ComponentIdPathing = ComponentId("pathing")

type CPathing struct {
	componentheader.ComponentHeader
	path *util.Path
}

func (c *CPathing) GetId() ComponentId {
	return ComponentIdPathing
}

func (c *CPathing) GetPath() *util.Path {
	return c.path
}

func (c *CPathing) SetPath(path *util.Path) {
	c.path = path
	c.Update()
}
