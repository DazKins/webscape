package component

import "webscape/server/util"

const ComponentIdMana = ComponentId("mana")

type CMana struct {
	maxMana     int
	currentMana int
}

func NewCMana(maxMana, currentMana int) *CMana {
	return &CMana{maxMana: maxMana, currentMana: max(0, min(currentMana, maxMana))}
}

func (c *CMana) GetId() ComponentId  { return ComponentIdMana }
func (c *CMana) GetMaxMana() int     { return c.maxMana }
func (c *CMana) GetCurrentMana() int { return c.currentMana }

func (c *CMana) Spend(amount int) bool {
	if amount < 0 || c.currentMana < amount {
		return false
	}
	c.currentMana -= amount
	return true
}

func (c *CMana) Restore(amount int) {
	// Bound the addition as well as its result to avoid integer overflow.
	c.currentMana += min(max(0, amount), c.maxMana-c.currentMana)
}

func (c *CMana) Serialize() util.Json {
	return util.JObject{
		"maxMana":     util.JNumber(c.maxMana),
		"currentMana": util.JNumber(c.currentMana),
	}
}
