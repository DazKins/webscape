package system

import (
	"webscape/server/game/component"
	"webscape/server/game/gameevent"
	"webscape/server/game/model"
)

type GameEventEmitter interface {
	EmitGameEvent(event gameevent.Event)
}

type SpatialCandidates interface {
	EntitiesAt(x int, y int) []model.EntityId
}

type System interface {
	Update()
}

type SystemBase struct {
	ComponentManager *component.ComponentManager
}
