package component

import (
	"webscape/server/util"
)

const ComponentIdHealth = ComponentId("health")

type CHealth struct {
	maxHealth     int
	currentHealth int
}

func NewCHealth(maxHealth int, currentHealth int) *CHealth {
	return &CHealth{
		maxHealth:     maxHealth,
		currentHealth: currentHealth,
	}
}

func (c *CHealth) GetId() ComponentId {
	return ComponentIdHealth
}

func (c *CHealth) Serialize() util.Json {
	return util.JObject(map[string]util.Json{
		"maxHealth":     util.JNumber(c.maxHealth),
		"currentHealth": util.JNumber(c.currentHealth),
	})
}

func (c *CHealth) GetMaxHealth() int {
	return c.maxHealth
}

func (c *CHealth) SetMaxHealth(maxHealth int) {
	c.maxHealth = maxHealth
}

func (c *CHealth) GetCurrentHealth() int {
	return c.currentHealth
}

func (c *CHealth) SetCurrentHealth(currentHealth int) {
	c.currentHealth = currentHealth
}
