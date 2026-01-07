package system

import (
	"webscape/server/game/component"
	"webscape/server/game/model"
)

type ChatMessageSender interface {
	SendChatMessageEntityFor(fromEntityId model.EntityId, message string) model.EntityId
}

type InteractionSystem struct {
	SystemBase
	ChatMessageSender ChatMessageSender
}

func (s *InteractionSystem) processInteraction(
	interacting *component.CInteracting,
) {
	switch interacting.GetOption() {
	case component.InteractionOptionTalk:
		// Handle talk interaction - target entity says "Hello!"
		s.ChatMessageSender.SendChatMessageEntityFor(
			interacting.GetTargetEntityId(), "Hello!")

	case component.InteractionOptionAttack:
		// Handle attack interaction - reduce target's health
		healthComponent := s.ComponentManager.GetEntityComponent(
			component.ComponentIdHealth, interacting.GetTargetEntityId())
		if healthComponent != nil {
			health := healthComponent.(*component.CHealth)
			// Deal 10 damage per attack
			newHealth := health.GetCurrentHealth() - 10
			if newHealth < 0 {
				newHealth = 0
			}
			health.SetCurrentHealth(newHealth)
			// Update the health component
			s.ComponentManager.SetEntityComponent(interacting.GetTargetEntityId(), health)

			// The target entity says "Ow!" when hit
			s.ChatMessageSender.SendChatMessageEntityFor(
				interacting.GetTargetEntityId(), "Ow!")
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
			component.ComponentIdPosition, interacting.GetTargetEntityId())

		if targetPositionComponent == nil {
			// Target no longer exists, cancel interaction
			s.ComponentManager.RemoveComponent(component.ComponentIdInteracting, entityId)
			continue
		}

		targetPosition := targetPositionComponent.(*component.CPosition)

		// Calculate Manhattan distance
		positionPos := position.GetPosition()
		targetPos := targetPosition.GetPosition()
		dx := targetPos.X - positionPos.X
		if dx < 0 {
			dx = -dx
		}
		dy := targetPos.Y - positionPos.Y
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
