package component

import "webscape/server/math"

const ComponentIdCombatAI = ComponentId("combatai")

type CCombatAI struct {
	aggroRadius       int
	leashRadius       int
	aggroTimeoutTicks int
	homePosition      math.Vec2
}

func NewCCombatAI(aggroRadius int, leashRadius int, aggroTimeoutTicks int, homePosition math.Vec2) *CCombatAI {
	return &CCombatAI{
		aggroRadius:       aggroRadius,
		leashRadius:       leashRadius,
		aggroTimeoutTicks: aggroTimeoutTicks,
		homePosition:      homePosition,
	}
}

func (c *CCombatAI) GetId() ComponentId {
	return ComponentIdCombatAI
}

func (c *CCombatAI) GetAggroRadius() int {
	return c.aggroRadius
}

func (c *CCombatAI) GetLeashRadius() int {
	return c.leashRadius
}

func (c *CCombatAI) GetAggroTimeoutTicks() int {
	return c.aggroTimeoutTicks
}

func (c *CCombatAI) GetHomePosition() math.Vec2 {
	return c.homePosition
}

func (c *CCombatAI) SetHomePosition(homePosition math.Vec2) {
	c.homePosition = homePosition
}
