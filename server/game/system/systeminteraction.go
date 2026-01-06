package system

import (
	"webscape/server/game/component"
	"webscape/server/game/entity"
)

type InteractionSystem struct {
	SystemBase
}

func (s *InteractionSystem) processInteraction(
	interacting *component.CInteracting,
) {
	switch interacting.Option {
	case component.InteractionOptionTalk:
		// Handle talk interaction - target entity says "Hello!"
		chatMessageComponents := entity.CreateChatMessageEntity(
			interacting.TargetEntityId, "Hello!")
		s.ComponentManager.CreateNewEntity(chatMessageComponents...)

	case component.InteractionOptionAttack:
		// Handle attack interaction - reduce target's health
		healthComponent := s.ComponentManager.GetEntityComponent(
			component.ComponentIdHealth, interacting.TargetEntityId)
		if healthComponent != nil {
			health := healthComponent.(*component.CHealth)
			// Deal 10 damage per attack
			health.CurrentHealth -= 10
			if health.CurrentHealth < 0 {
				health.CurrentHealth = 0
			}
			// Update the health component
			s.ComponentManager.SetEntityComponent(interacting.TargetEntityId, health)

			// The target entity says "Ow!" when hit
			chatMessageComponents := entity.CreateChatMessageEntity(
				interacting.TargetEntityId, "Ow!")
			s.ComponentManager.CreateNewEntity(chatMessageComponents...)
		}
	}
}

func (s *InteractionSystem) Update() {
	// Find all entities that are trying to interact
	entityIds := s.ComponentManager.GetEntitiesWithComponents(
		component.ComponentIdInteracting,
		component.ComponentIdPosition,
	)

	for _, entityId := range entityIds {
		interacting := s.ComponentManager.GetEntityComponent(
			component.ComponentIdInteracting, entityId).(*component.CInteracting)
		position := s.ComponentManager.GetEntityComponent(
			component.ComponentIdPosition, entityId).(*component.CPosition)

		// Check if target entity still exists and get its position
		targetPositionComponent := s.ComponentManager.GetEntityComponent(
			component.ComponentIdPosition, interacting.TargetEntityId)

		if targetPositionComponent == nil {
			// Target no longer exists, cancel interaction
			s.ComponentManager.RemoveComponent(component.ComponentIdInteracting, entityId)
			continue
		}

		targetPosition := targetPositionComponent.(*component.CPosition)

		// Calculate Manhattan distance
		dx := targetPosition.Position.X - position.Position.X
		if dx < 0 {
			dx = -dx
		}
		dy := targetPosition.Position.Y - position.Position.Y
		if dy < 0 {
			dy = -dy
		}
		distance := dx + dy

		// Check if in range (1 tile away for interactions)
		if distance <= 1 {
			// Ready to interact! Process the interaction
			s.processInteraction(interacting)

			// Remove the interacting component after processing
			s.ComponentManager.RemoveComponent(component.ComponentIdInteracting, entityId)
		}
		// If not in range, keep the component and let pathing system handle movement
	}
}
