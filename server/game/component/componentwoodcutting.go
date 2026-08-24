package component

import (
	"webscape/server/game/model"
	"webscape/server/math"
)

const ComponentIdWoodcutting = ComponentId("woodcutting")

type CWoodcutting struct {
	targetEntityId    model.EntityId
	startPosition     math.Vec2
	cooldownRemaining int
}

func NewCWoodcutting(targetEntityId model.EntityId, startPosition math.Vec2) *CWoodcutting {
	return &CWoodcutting{
		targetEntityId: targetEntityId,
		startPosition:  startPosition,
	}
}

func (c *CWoodcutting) GetId() ComponentId {
	return ComponentIdWoodcutting
}

func (c *CWoodcutting) GetTargetEntityId() model.EntityId {
	return c.targetEntityId
}

func (c *CWoodcutting) GetStartPosition() math.Vec2 {
	return c.startPosition
}

func (c *CWoodcutting) GetCooldownRemaining() int {
	return c.cooldownRemaining
}

func (c *CWoodcutting) SetCooldownRemaining(ticks int) {
	if ticks < 0 {
		ticks = 0
	}
	c.cooldownRemaining = ticks
}
