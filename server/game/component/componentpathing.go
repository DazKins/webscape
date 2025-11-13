package component

import (
	"webscape/server/util"
)

const ComponentIdPathing = ComponentId("pathing")

type CPathing struct {
	Path *util.Path
}

func (c *CPathing) GetId() ComponentId {
	return ComponentIdPathing
}
