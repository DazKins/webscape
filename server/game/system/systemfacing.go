package system

import (
	"webscape/server/game/component"
	"webscape/server/game/model"
)

const (
	facingPriorityOutgoingCombat = iota
	facingPriorityIncomingCombat
	facingPriorityOutgoingWoodcutting
	facingPriorityOutgoingConversation
	facingPriorityIncomingConversation
)

type facingCandidate struct {
	targetEntityId model.EntityId
	priority       int
	distance       int
}

type FacingSystem struct {
	SystemBase
}

func (s *FacingSystem) Update() {
	desired := make(map[model.EntityId]facingCandidate)

	for entityId, value := range s.ComponentManager.GetComponent(component.ComponentIdCombatState) {
		targetEntityId := value.(*component.CCombatState).GetTargetId()
		s.addPair(
			desired,
			entityId,
			targetEntityId,
			facingPriorityOutgoingCombat,
			facingPriorityIncomingCombat,
		)
	}

	for entityId, value := range s.ComponentManager.GetComponent(component.ComponentIdActiveConversation) {
		targetEntityId := value.(*component.CActiveConversation).GetTargetEntityId()
		s.addPair(
			desired,
			entityId,
			targetEntityId,
			facingPriorityOutgoingConversation,
			facingPriorityIncomingConversation,
		)
	}

	for entityId, value := range s.ComponentManager.GetComponent(component.ComponentIdWoodcutting) {
		targetEntityId := value.(*component.CWoodcutting).GetTargetEntityId()
		entityPosition := s.position(entityId)
		targetPosition := s.position(targetEntityId)
		if entityPosition == nil || targetPosition == nil {
			continue
		}
		s.addCandidate(
			desired,
			entityId,
			targetEntityId,
			facingPriorityOutgoingWoodcutting,
			manhattanDistance(entityPosition.GetPosition(), targetPosition.GetPosition()),
		)
	}

	for entityId, value := range s.ComponentManager.GetComponent(component.ComponentIdFacing) {
		candidate, ok := desired[entityId]
		if !ok {
			s.ComponentManager.RemoveComponent(component.ComponentIdFacing, entityId)
			continue
		}

		facing := value.(*component.CFacing)
		if facing.GetTargetEntityId() != candidate.targetEntityId {
			s.ComponentManager.SetEntityComponent(entityId, component.NewCFacing(candidate.targetEntityId))
		}
		delete(desired, entityId)
	}

	for entityId, candidate := range desired {
		s.ComponentManager.SetEntityComponent(entityId, component.NewCFacing(candidate.targetEntityId))
	}
}

func (s *FacingSystem) addPair(
	desired map[model.EntityId]facingCandidate,
	entityId model.EntityId,
	targetEntityId model.EntityId,
	entityPriority int,
	targetPriority int,
) {
	entityPosition := s.position(entityId)
	targetPosition := s.position(targetEntityId)
	if entityPosition == nil || targetPosition == nil {
		return
	}

	distance := manhattanDistance(entityPosition.GetPosition(), targetPosition.GetPosition())
	s.addCandidate(desired, entityId, targetEntityId, entityPriority, distance)
	s.addCandidate(desired, targetEntityId, entityId, targetPriority, distance)
}

func (s *FacingSystem) position(entityId model.EntityId) *component.CPosition {
	value := s.ComponentManager.GetEntityComponent(component.ComponentIdPosition, entityId)
	if value == nil {
		return nil
	}
	return value.(*component.CPosition)
}

func (s *FacingSystem) addCandidate(
	desired map[model.EntityId]facingCandidate,
	entityId model.EntityId,
	targetEntityId model.EntityId,
	priority int,
	distance int,
) {
	candidate := facingCandidate{
		targetEntityId: targetEntityId,
		priority:       priority,
		distance:       distance,
	}
	existing, ok := desired[entityId]
	if !ok || candidate.isPreferredTo(existing) {
		desired[entityId] = candidate
	}
}

func (c facingCandidate) isPreferredTo(other facingCandidate) bool {
	if c.priority != other.priority {
		return c.priority < other.priority
	}
	if c.distance != other.distance {
		return c.distance < other.distance
	}
	return c.targetEntityId.String() < other.targetEntityId.String()
}
