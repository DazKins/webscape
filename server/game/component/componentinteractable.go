package component

import (
	"webscape/server/util"
)

const ComponentIdInteractable = ComponentId("interactable")

type InteractionOption string

const (
	InteractionOptionTalk   = "talk"
	InteractionOptionTrade  = "trade"
	InteractionOptionAttack = "attack"
)

type CInteractable struct {
	interactionOptions []InteractionOption
}

func NewCInteractable(interactionOptions []InteractionOption) *CInteractable {
	return &CInteractable{
		interactionOptions: interactionOptions,
	}
}

func (c *CInteractable) GetId() ComponentId {
	return ComponentIdInteractable
}

func (c *CInteractable) Serialize() util.Json {
	return util.JObject(map[string]util.Json{
		"interactionOptions": util.JArrayFrom(c.interactionOptions, func(option InteractionOption) util.Json {
			return util.JString(option)
		}),
	})
}

func (c *CInteractable) GetInteractionOptions() []InteractionOption {
	// Return a copy to prevent external modification
	result := make([]InteractionOption, len(c.interactionOptions))
	copy(result, c.interactionOptions)
	return result
}

func (c *CInteractable) SetInteractionOptions(interactionOptions []InteractionOption) {
	c.interactionOptions = interactionOptions
}
