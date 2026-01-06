package component

import (
	"webscape/server/game/model"
)

const ComponentIdInteracting = ComponentId("interacting")

type CInteracting struct {
	TargetEntityId model.EntityId
	Option         InteractionOption
}

func (c *CInteracting) GetId() ComponentId {
	return ComponentIdInteracting
}
