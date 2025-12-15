package system

import (
	"webscape/server/game/component"
)

type PathingSystem struct {
	SystemBase
}

func (s *PathingSystem) Update() {
	entityIds := s.ComponentManager.GetEntitiesWithComponents(component.ComponentIdPosition, component.ComponentIdPathing)

	for _, entityId := range entityIds {
		positionComponent := s.ComponentManager.GetEntityComponent(component.ComponentIdPosition, entityId).(*component.CPosition)
		pathingComponent := s.ComponentManager.GetEntityComponent(component.ComponentIdPathing, entityId).(*component.CPathing)

		path := pathingComponent.Path

		if path.Size() == 0 {
			s.ComponentManager.RemoveComponent(component.ComponentIdPathing, entityId)
			continue
		}

		nextPosition := path.Pop()
		positionComponent.Position = *nextPosition
	}
}
