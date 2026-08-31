package component

import "webscape/server/util"

const ComponentIdLocomotion = ComponentId("locomotion")

type LocomotionPhase string

const (
	LocomotionPhaseIdle   LocomotionPhase = "idle"
	LocomotionPhaseMoving LocomotionPhase = "moving"
)

type CLocomotion struct {
	phase            LocomotionPhase
	phaseStartedTick uint64
	lastMovementTick uint64
}

func NewCLocomotion(phase LocomotionPhase, phaseStartedTick uint64) *CLocomotion {
	return &CLocomotion{
		phase:            phase,
		phaseStartedTick: phaseStartedTick,
	}
}

func (c *CLocomotion) GetId() ComponentId {
	return ComponentIdLocomotion
}

func (c *CLocomotion) Serialize() util.Json {
	return util.JObject{
		"phase":            util.JString(c.phase),
		"phaseStartedTick": util.JNumber(c.phaseStartedTick),
	}
}

func (c *CLocomotion) GetPhase() LocomotionPhase {
	return c.phase
}

func (c *CLocomotion) GetPhaseStartedTick() uint64 {
	return c.phaseStartedTick
}

func (c *CLocomotion) GetLastMovementTick() uint64 {
	return c.lastMovementTick
}

func (c *CLocomotion) MarkMoving(tick uint64) bool {
	changed := c.phase != LocomotionPhaseMoving
	if changed {
		c.phase = LocomotionPhaseMoving
		c.phaseStartedTick = tick
	}
	c.lastMovementTick = tick
	return changed
}

func (c *CLocomotion) SetIdle(tick uint64) bool {
	if c.phase == LocomotionPhaseIdle {
		return false
	}
	c.phase = LocomotionPhaseIdle
	c.phaseStartedTick = tick
	return true
}
