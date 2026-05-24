package component

import "webscape/server/game/model"

const ComponentIdCombatState = ComponentId("combatstate")

type CCombatState struct {
	targetId          model.EntityId
	cooldownRemaining int
}

func NewCCombatState(targetId model.EntityId) *CCombatState {
	return &CCombatState{
		targetId:          targetId,
		cooldownRemaining: 0,
	}
}

func (c *CCombatState) GetId() ComponentId {
	return ComponentIdCombatState
}

func (c *CCombatState) GetTargetId() model.EntityId {
	return c.targetId
}

func (c *CCombatState) SetTargetId(targetId model.EntityId) {
	c.targetId = targetId
}

func (c *CCombatState) GetCooldownRemaining() int {
	return c.cooldownRemaining
}

func (c *CCombatState) SetCooldownRemaining(cooldownRemaining int) {
	c.cooldownRemaining = cooldownRemaining
}
