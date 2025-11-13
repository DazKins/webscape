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
	InteractionOptions []InteractionOption
}

func (c *CInteractable) GetId() ComponentId {
	return ComponentIdInteractable
}

func (c *CInteractable) Serialize() util.Json {
	return util.JObject(map[string]util.Json{
		"interactionOptions": util.JArrayFrom(c.InteractionOptions, func(option InteractionOption) util.Json {
			return util.JString(option)
		}),
	})
}
