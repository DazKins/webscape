package system

import (
	"webscape/server/game/component"
)

type PathingSystem struct {
	SystemBase
}

func (s *PathingSystem) Update() {
	entities := s.entityGetter.GetEntitiesWithComponents(component.ComponentIdPosition, component.ComponentIdPathing)

	for _, entity := range entities {
		positionComponent := entity.GetComponent(component.ComponentIdPosition).(*component.CPosition)
		pathingComponent := entity.GetComponent(component.ComponentIdPathing).(*component.CPathing)

		path := pathingComponent.GetPath()

		if path.Size() == 0 {
			entity.RemoveComponent(component.ComponentIdPathing)
			continue
		}

		nextPosition := path.Pop()
		positionComponent.SetPosition(*nextPosition)
	}
}
