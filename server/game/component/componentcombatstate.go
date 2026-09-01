package component

import (
	"webscape/server/game/model"
	"webscape/server/util"
)

const ComponentIdCombatState = ComponentId("combatstate")

type CCombatState struct {
	targetId         model.EntityId
	phase            CombatPhase
	phaseStartedTick uint64
	attackMethod     model.AttackMethod
	windUpTicks      int
	nextAttackTick   uint64
}

type CombatPhase string

const (
	CombatPhaseApproaching CombatPhase = "approaching"
	CombatPhaseCasting     CombatPhase = "casting"
	CombatPhaseRecovering  CombatPhase = "recovering"
)

func NewCCombatState(targetId model.EntityId) *CCombatState {
	return &CCombatState{
		targetId:     targetId,
		phase:        CombatPhaseApproaching,
		attackMethod: model.AttackMethodMelee,
	}
}

func (c *CCombatState) GetId() ComponentId {
	return ComponentIdCombatState
}

func (c *CCombatState) Serialize() util.Json {
	return util.JObject(map[string]util.Json{
		"targetEntityId":   util.JString(c.targetId.String()),
		"phase":            util.JString(string(c.phase)),
		"phaseStartedTick": util.JNumber(c.phaseStartedTick),
		"attackMethod":     util.JString(string(c.attackMethod)),
		"windUpTicks":      util.JNumber(c.windUpTicks),
	})
}

func (c *CCombatState) GetTargetId() model.EntityId {
	return c.targetId
}

func (c *CCombatState) SetTargetId(targetId model.EntityId) {
	c.targetId = targetId
}

func (c *CCombatState) GetPhase() CombatPhase { return c.phase }

func (c *CCombatState) GetPhaseStartedTick() uint64 { return c.phaseStartedTick }

func (c *CCombatState) GetAttackMethod() model.AttackMethod { return c.attackMethod }

func (c *CCombatState) GetWindUpTicks() int { return c.windUpTicks }

func (c *CCombatState) GetNextAttackTick() uint64 { return c.nextAttackTick }

func (c *CCombatState) ProfileMatches(stats *CCombatStats) bool {
	return c.attackMethod == stats.GetAttackMethod() && c.windUpTicks == stats.GetWindUpTicks()
}

func (c *CCombatState) Restart(stats *CCombatStats, tick uint64) {
	c.phase = CombatPhaseApproaching
	c.phaseStartedTick = tick
	c.attackMethod = stats.GetAttackMethod()
	c.windUpTicks = stats.GetWindUpTicks()
	c.nextAttackTick = tick
}

func (c *CCombatState) BeginCasting(tick uint64, nextAttackTick uint64) {
	c.phase = CombatPhaseCasting
	c.phaseStartedTick = tick
	c.nextAttackTick = nextAttackTick
}

func (c *CCombatState) BeginRecovering(tick uint64, nextAttackTick uint64) {
	c.phase = CombatPhaseRecovering
	c.phaseStartedTick = tick
	c.nextAttackTick = nextAttackTick
}

func (c *CCombatState) BeginApproaching(tick uint64) {
	c.phase = CombatPhaseApproaching
	c.phaseStartedTick = tick
}
