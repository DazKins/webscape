package component

import (
	"webscape/server/game/model"
	"webscape/server/math"
	"webscape/server/util"
)

const ComponentIdWoodcutting = ComponentId("woodcutting")

type WoodcuttingPhase string

const (
	WoodcuttingPhaseSwinging   WoodcuttingPhase = "swinging"
	WoodcuttingPhaseRecovering WoodcuttingPhase = "recovering"
)

type CWoodcutting struct {
	targetEntityId   model.EntityId
	phase            WoodcuttingPhase
	phaseStartedTick uint64
	originPosition   math.Vec2
}

func NewCWoodcutting(
	targetEntityId model.EntityId,
	phaseStartedTick uint64,
	originPosition math.Vec2,
) *CWoodcutting {
	return &CWoodcutting{
		targetEntityId:   targetEntityId,
		phase:            WoodcuttingPhaseSwinging,
		phaseStartedTick: phaseStartedTick,
		originPosition:   originPosition,
	}
}

func (c *CWoodcutting) GetId() ComponentId {
	return ComponentIdWoodcutting
}

func (c *CWoodcutting) Serialize() util.Json {
	return util.JObject{
		"targetEntityId":   util.JString(c.targetEntityId.String()),
		"phase":            util.JString(c.phase),
		"phaseStartedTick": util.JNumber(c.phaseStartedTick),
	}
}

func (c *CWoodcutting) GetTargetEntityId() model.EntityId {
	return c.targetEntityId
}

func (c *CWoodcutting) GetPhase() WoodcuttingPhase {
	return c.phase
}

func (c *CWoodcutting) GetPhaseStartedTick() uint64 {
	return c.phaseStartedTick
}

func (c *CWoodcutting) GetOriginPosition() math.Vec2 {
	return c.originPosition
}

func (c *CWoodcutting) StartPhase(phase WoodcuttingPhase, tick uint64) {
	c.phase = phase
	c.phaseStartedTick = tick
}
