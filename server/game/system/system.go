package system

import (
	"webscape/server/game/component"
	"webscape/server/game/gameevent"
	"webscape/server/game/model"
)

type GameEventEmitter interface {
	EmitGameEvent(event gameevent.Event)
}

type TickSource interface {
	CurrentTick() uint64
}

type SpatialCandidates interface {
	EntitiesAt(x int, y int) []model.EntityId
}

type System interface {
	Update()
}

type SystemBase struct {
	ComponentManager *component.ComponentManager
	StateTransitions *EntityStateTransitions
}

func (s SystemBase) entityStateTransitions(tickSource TickSource) *EntityStateTransitions {
	if s.StateTransitions != nil {
		return s.StateTransitions
	}
	return NewEntityStateTransitions(s.ComponentManager, tickSource)
}
