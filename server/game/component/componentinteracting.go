package component

import (
	"webscape/server/game/model"
)

const ComponentIdInteracting = ComponentId("interacting")

type CInteracting struct {
	targetEntityId model.EntityId
	option         InteractionOption
}

func NewCInteracting(targetEntityId model.EntityId, option InteractionOption) *CInteracting {
	return &CInteracting{
		targetEntityId: targetEntityId,
		option:         option,
	}
}

func (c *CInteracting) GetId() ComponentId {
	return ComponentIdInteracting
}

func (c *CInteracting) GetTargetEntityId() model.EntityId {
	return c.targetEntityId
}

func (c *CInteracting) SetTargetEntityId(targetEntityId model.EntityId) {
	c.targetEntityId = targetEntityId
}

func (c *CInteracting) GetOption() InteractionOption {
	return c.option
}

func (c *CInteracting) SetOption(option InteractionOption) {
	c.option = option
}
