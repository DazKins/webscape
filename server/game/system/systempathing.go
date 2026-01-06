package system

import (
	"log"
	"webscape/server/game/component"
	"webscape/server/game/entity"
	"webscape/server/game/model"
	"webscape/server/game/world"
	"webscape/server/math"
)

type PathingSystem struct {
	SystemBase
	World *world.World
}

func (s *PathingSystem) handleInteractionCompletion(entityId model.EntityId) {
	// Check if this entity was interacting with another entity
	interactingComponent := s.ComponentManager.GetEntityComponent(component.ComponentIdInteracting, entityId)
	if interactingComponent != nil {
		interacting := interactingComponent.(*component.CInteracting)

		// Handle the interaction based on the option
		if interacting.Option == component.InteractionOptionTalk {
			// The target entity says "Hello!"
			chatMessageComponents := entity.CreateChatMessageEntity(interacting.TargetEntityId, "Hello!")
			s.ComponentManager.CreateNewEntity(chatMessageComponents...)
		}

		// Remove the interacting component after handling
		s.ComponentManager.RemoveComponent(component.ComponentIdInteracting, entityId)
	}
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
			s.handleInteractionCompletion(entityId)
			continue
		}

		pathToPosition := math.Vec2{}
		isEntityTarget := false

		if target.Position.IsPresent() {
			pathToPosition = target.Position.Unwrap()
		} else if target.EntityId.IsPresent() {
			targetEntityId := target.EntityId.Unwrap()
			targetEntityPosition := s.ComponentManager.GetEntityComponent(component.ComponentIdPosition, targetEntityId).(*component.CPosition)
			pathToPosition = targetEntityPosition.Position
			isEntityTarget = true
		}

		// Calculate Manhattan distance
		dx := pathToPosition.X - positionComponent.Position.X
		if dx < 0 {
			dx = -dx
		}
		dy := pathToPosition.Y - positionComponent.Position.Y
		if dy < 0 {
			dy = -dy
		}
		distance := dx + dy

		// For entity targets, stop when 1 tile away (adjacent). For position targets, stop when at exact position.
		shouldStop := false
		if isEntityTarget {
			shouldStop = distance <= 1
		} else {
			shouldStop = distance == 0
		}

		if shouldStop {
			s.ComponentManager.RemoveComponent(component.ComponentIdPathing, entityId)
			s.handleInteractionCompletion(entityId)
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
