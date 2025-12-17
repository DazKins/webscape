package system

import (
	"log"
	"webscape/server/game/component"
	"webscape/server/game/world"
	"webscape/server/math"
)

type PathingSystem struct {
	SystemBase
	World *world.World
}

func (s *PathingSystem) Update() {
	entityIds := s.ComponentManager.GetEntitiesWithComponents(component.ComponentIdPosition, component.ComponentIdPathing)

	for _, entityId := range entityIds {
		positionComponent := s.ComponentManager.GetEntityComponent(component.ComponentIdPosition, entityId).(*component.CPosition)
		pathingComponent := s.ComponentManager.GetEntityComponent(component.ComponentIdPathing, entityId).(*component.CPathing)

		target := pathingComponent.Target
		path := pathingComponent.Path

		if path != nil && path.Size() == 0 {
			s.ComponentManager.RemoveComponent(component.ComponentIdPathing, entityId)
			continue
		}

		pathToPosition := math.Vec2{}

		if target.Position.IsPresent() {
			pathToPosition = target.Position.Unwrap()
		} else if target.EntityId.IsPresent() {
			targetEntityId := target.EntityId.Unwrap()
			targetEntityPosition := s.ComponentManager.GetEntityComponent(component.ComponentIdPosition, targetEntityId).(*component.CPosition)
			pathToPosition = targetEntityPosition.Position
		}

		if pathToPosition.Eq(positionComponent.Position) {
			s.ComponentManager.RemoveComponent(component.ComponentIdPathing, entityId)
			continue
		}

		// It's easiest just to recompute the path every time for now...
		newPath, err := s.World.GetPath(positionComponent.Position, pathToPosition)
		if err != nil {
			log.Printf("failed to get path: %v\n", err)
			continue
		}
		pathingComponent.Path = &newPath
		path = &newPath

		nextPosition := path.Pop()
		positionComponent.Position = *nextPosition
	}
}
