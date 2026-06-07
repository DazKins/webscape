package system

import (
	"webscape/server/game/component"
	"webscape/server/game/gameevent"
	"webscape/server/game/model"
	"webscape/server/util"
)

type ConversationStarter interface {
	StartConversationFor(playerEntityId model.EntityId, targetEntityId model.EntityId)
}

type GameEventEmitter interface {
	EmitGameEvent(event gameevent.Event)
}

type LootHandler interface {
	LootEntityFor(playerEntityId model.EntityId, targetEntityId model.EntityId)
}

type InteractionSystem struct {
	SystemBase
	ConversationStarter ConversationStarter
	EventEmitter        GameEventEmitter
	LootHandler         LootHandler
}

func (s *InteractionSystem) processInteraction(
	entityId model.EntityId,
	interacting *component.CInteracting,
) {
	switch interacting.GetOption() {
	case component.InteractionOptionTalk:
		s.ConversationStarter.StartConversationFor(
			entityId, interacting.GetTargetEntityId())

	case component.InteractionOptionAttack:
		// Start combat with the target entity
		combatState := component.NewCCombatState(interacting.GetTargetEntityId())
		s.ComponentManager.SetEntityComponent(entityId, combatState)

	case component.InteractionOptionLoot:
		if s.LootHandler != nil {
			s.LootHandler.LootEntityFor(entityId, interacting.GetTargetEntityId())
		}

	case component.InteractionOptionOpen:
		if !s.setOpenableState(interacting.GetTargetEntityId(), true) {
			return
		}

	case component.InteractionOptionClose:
		if !s.setOpenableState(interacting.GetTargetEntityId(), false) {
			return
		}
	}

	s.emitInteractEvent(entityId, interacting)
}

func (s *InteractionSystem) setOpenableState(targetEntityId model.EntityId, isOpen bool) bool {
	openable := s.ComponentManager.GetEntityComponent(component.ComponentIdOpenable, targetEntityId)
	if openable == nil {
		return false
	}
	openableComponent := openable.(*component.COpenable)
	openableComponent.SetOpen(isOpen)
	s.ComponentManager.SetEntityComponent(targetEntityId, openableComponent)
	return true
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
			s.processInteraction(entityId, interacting)

			// Remove the interacting component after processing
			s.ComponentManager.RemoveComponent(component.ComponentIdInteracting, entityId)
		}
		// If not in range, keep the component and let pathing system handle movement
	}
}

func (s *InteractionSystem) emitInteractEvent(entityId model.EntityId, interacting *component.CInteracting) {
	if s.EventEmitter == nil {
		return
	}
	metadata := s.ComponentManager.GetEntityComponent(component.ComponentIdMetadata, interacting.GetTargetEntityId())
	if metadata == nil {
		return
	}
	metadataObject, ok := metadata.(*component.CMetadata).GetMetadata().(util.JObject)
	if !ok {
		return
	}
	objectId, ok := metadataObject["objectId"].(util.JString)
	if !ok || objectId == "" {
		return
	}

	event := gameevent.New(
		"interact:object:"+gameevent.NormalizeToken(string(objectId))+":"+gameevent.NormalizeToken(string(interacting.GetOption())),
		entityId,
	)
	event.TargetEntityId = interacting.GetTargetEntityId()
	s.EventEmitter.EmitGameEvent(event)
}
