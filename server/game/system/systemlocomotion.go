package system

import (
	"webscape/server/game/component"
)

type LocomotionSystem struct {
	SystemBase
	TickSource TickSource
}

func (s *LocomotionSystem) Update() {
	tick := currentTick(s.TickSource)
	for entityId, value := range s.ComponentManager.GetComponent(component.ComponentIdLocomotion) {
		locomotion := value.(*component.CLocomotion)
		if locomotion.GetPhase() != component.LocomotionPhaseMoving ||
			locomotion.GetLastMovementTick() >= tick {
			continue
		}
		if locomotion.SetIdle(tick) {
			s.ComponentManager.SetEntityComponent(entityId, locomotion)
		}
	}
}

func currentTick(source TickSource) uint64 {
	if source == nil {
		return 0
	}
	return source.CurrentTick()
}
