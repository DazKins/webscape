package component

import (
	"webscape/server/game/model"
	"webscape/server/math"
	"webscape/server/util"
)

const ComponentIdFishing = ComponentId("fishing")

type FishingPhase string

const (
	FishingPhaseCasting FishingPhase = "casting"
	FishingPhaseWaiting FishingPhase = "waiting"
	FishingPhaseReeling FishingPhase = "reeling"
)

type CFishing struct {
	targetEntityId   model.EntityId
	phase            FishingPhase
	phaseStartedTick uint64
	originPosition   math.Vec2
}

func NewCFishing(
	targetEntityId model.EntityId,
	phaseStartedTick uint64,
	originPosition math.Vec2,
) *CFishing {
	return &CFishing{
		targetEntityId:   targetEntityId,
		phase:            FishingPhaseCasting,
		phaseStartedTick: phaseStartedTick,
		originPosition:   originPosition,
	}
}

func (c *CFishing) GetId() ComponentId {
	return ComponentIdFishing
}

func (c *CFishing) Serialize() util.Json {
	return util.JObject(map[string]util.Json{
		"targetEntityId":   util.JString(c.targetEntityId.String()),
		"phase":            util.JString(c.phase),
		"phaseStartedTick": util.JNumber(c.phaseStartedTick),
	})
}

func (c *CFishing) GetTargetEntityId() model.EntityId {
	return c.targetEntityId
}

func (c *CFishing) GetPhase() FishingPhase {
	return c.phase
}

func (c *CFishing) GetPhaseStartedTick() uint64 {
	return c.phaseStartedTick
}

func (c *CFishing) GetOriginPosition() math.Vec2 {
	return c.originPosition
}

func (c *CFishing) StartPhase(phase FishingPhase, tick uint64) {
	c.phase = phase
	c.phaseStartedTick = tick
}
