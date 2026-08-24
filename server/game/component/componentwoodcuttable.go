package component

import (
	"webscape/server/game/model"
	"webscape/server/util"
)

const ComponentIdWoodcuttable = ComponentId("woodcuttable")

type CWoodcuttable struct {
	maxDurability         int
	currentDurability     int
	respawnTicks          int
	yield                 LootItem
	depleted              bool
	remainingRespawnTicks int
	lastFellerEntityId    model.EntityId
}

func NewCWoodcuttable(maxDurability int, respawnTicks int, yield LootItem) *CWoodcuttable {
	return &CWoodcuttable{
		maxDurability:     maxDurability,
		currentDurability: maxDurability,
		respawnTicks:      respawnTicks,
		yield:             yield,
	}
}

func (c *CWoodcuttable) GetId() ComponentId {
	return ComponentIdWoodcuttable
}

func (c *CWoodcuttable) Serialize() util.Json {
	return util.JObject(map[string]util.Json{
		"maxDurability":         util.JNumber(c.maxDurability),
		"currentDurability":     util.JNumber(c.currentDurability),
		"depleted":              util.JBool(c.depleted),
		"remainingRespawnTicks": util.JNumber(c.remainingRespawnTicks),
	})
}

func (c *CWoodcuttable) GetMaxDurability() int {
	return c.maxDurability
}

func (c *CWoodcuttable) GetCurrentDurability() int {
	return c.currentDurability
}

func (c *CWoodcuttable) SetCurrentDurability(durability int) {
	if durability < 0 {
		durability = 0
	}
	if durability > c.maxDurability {
		durability = c.maxDurability
	}
	c.currentDurability = durability
}

func (c *CWoodcuttable) GetRespawnTicks() int {
	return c.respawnTicks
}

func (c *CWoodcuttable) GetYield() LootItem {
	return c.yield
}

func (c *CWoodcuttable) IsDepleted() bool {
	return c.depleted
}

func (c *CWoodcuttable) SetDepleted(depleted bool) {
	c.depleted = depleted
}

func (c *CWoodcuttable) GetRemainingRespawnTicks() int {
	return c.remainingRespawnTicks
}

func (c *CWoodcuttable) SetRemainingRespawnTicks(ticks int) {
	if ticks < 0 {
		ticks = 0
	}
	c.remainingRespawnTicks = ticks
}

func (c *CWoodcuttable) GetLastFellerEntityId() model.EntityId {
	return c.lastFellerEntityId
}

func (c *CWoodcuttable) SetLastFellerEntityId(entityId model.EntityId) {
	c.lastFellerEntityId = entityId
}
