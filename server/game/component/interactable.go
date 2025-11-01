package component

import "webscape/server/game/componentheader"

const ComponentIdInteractable = ComponentId("interactable")

type InteractionOption string

const (
	InteractionOptionTalk   = "talk"
	InteractionOptionTrade  = "trade"
	InteractionOptionAttack = "attack"
)

type CInteractable struct {
	componentheader.ComponentHeader
	interactionOptions []InteractionOption
}

func (c *CInteractable) GetId() ComponentId {
	return ComponentIdInteractable
}

func (c *CInteractable) Serialize() map[string]any {
	return map[string]any{
		"interactionOptions": c.interactionOptions,
	}
}

func (c *CInteractable) GetInteractionOptions() []InteractionOption {
	return c.interactionOptions
}

func (c *CInteractable) SetInteractionOptions(interactionOptions []InteractionOption) {
	c.interactionOptions = interactionOptions
	c.Update()
}
