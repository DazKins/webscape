package component

import (
	"webscape/server/util"
)

const ComponentIdHealth = ComponentId("health")

type CHealth struct {
	MaxHealth     int
	CurrentHealth int
}

func (c *CHealth) GetId() ComponentId {
	return ComponentIdHealth
}

func (c *CHealth) Serialize() util.Json {
	return util.JObject(map[string]util.Json{
		"maxHealth":     util.JNumber(c.MaxHealth),
		"currentHealth": util.JNumber(c.CurrentHealth),
	})
}
