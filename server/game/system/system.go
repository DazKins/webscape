package system

import (
	"webscape/server/game/component"
	"webscape/server/game/entity"
)

type System interface {
	SetEntityGetter(entityGetter entityGetter)
	Update()
}

type entityGetter interface {
	GetEntitiesWithComponents(componentIds ...component.ComponentId) []*entity.Entity
}

type SystemBase struct {
	entityGetter entityGetter
}

func (s *SystemBase) SetEntityGetter(entityGetter entityGetter) {
	s.entityGetter = entityGetter
}
