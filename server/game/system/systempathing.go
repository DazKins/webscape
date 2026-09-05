package system

import (
	"webscape/server/game/collision"
	"webscape/server/game/component"
	"webscape/server/game/gameevent"
	"webscape/server/game/model"
	"webscape/server/game/world"
	"webscape/server/math"
)

const pathNotFoundMessage = "I can't find a way there!"

type PathingSystem struct {
	SystemBase
	World        *world.World
	SpatialIndex SpatialCandidates
	EventEmitter GameEventEmitter
	TickSource   TickSource
}

func (s *PathingSystem) Update() {
	entityIds := s.ComponentManager.GetEntitiesWithComponents(component.ComponentIdPathing, component.ComponentIdPosition)

	for _, entityId := range entityIds {
		positionComponent := s.ComponentManager.GetEntityComponent(component.ComponentIdPosition, entityId).(*component.CPosition)
		pathingComponent := s.ComponentManager.GetEntityComponent(component.ComponentIdPathing, entityId).(*component.CPathing)

		target := pathingComponent.GetTarget()
		path := pathingComponent.GetPath()

		pathToPosition := math.Vec2{}
		isEntityTarget := false
		inCombat := s.ComponentManager.GetEntityComponent(component.ComponentIdCombatState, entityId) != nil
		isAttackInteraction := false
		if interacting := s.ComponentManager.GetEntityComponent(component.ComponentIdInteracting, entityId); interacting != nil {
			isAttackInteraction = interacting.(*component.CInteracting).GetOption() == component.InteractionOptionAttack
		}

		if target.Position.IsPresent() {
			pathToPosition = target.Position.Unwrap()
		} else if target.EntityId.IsPresent() {
			targetEntityId := target.EntityId.Unwrap()
			targetPositionValue := s.ComponentManager.GetEntityComponent(component.ComponentIdPosition, targetEntityId)
			if targetPositionValue == nil {
				s.entityStateTransitions(s.TickSource).RejectPathing(entityId)
				continue
			}
			targetEntityPosition := targetPositionValue.(*component.CPosition)
			pathToPosition = targetEntityPosition.GetPosition()
			isEntityTarget = true
		}

		// Calculate Manhattan distance
		positionPos := positionComponent.GetPosition()
		dx := pathToPosition.X - positionPos.X
		if dx < 0 {
			dx = -dx
		}
		dy := pathToPosition.Y - positionPos.Y
		if dy < 0 {
			dy = -dy
		}
		distance := dx + dy

		if isEntityTarget && distance == 0 {
			s.resolveOverlap(entityId, pathToPosition)
			s.ComponentManager.RemoveComponent(component.ComponentIdPathing, entityId)
			continue
		}

		// For entity targets, stop at combat range if in combat, otherwise stop when adjacent.
		// For position targets, stop when at exact position.
		shouldStop := false
		stopDistance := 1
		if isEntityTarget {
			if inCombat || isAttackInteraction {
				if combatStatsComponent := s.ComponentManager.GetEntityComponent(component.ComponentIdCombatStats, entityId); combatStatsComponent != nil {
					attackRange := combatStatsComponent.(*component.CCombatStats).GetAttackRange()
					if attackRange > stopDistance {
						stopDistance = attackRange
					}
				}
			}
			shouldStop = distance <= stopDistance
		} else {
			shouldStop = distance == 0
		}

		if shouldStop {
			s.ComponentManager.RemoveComponent(component.ComponentIdPathing, entityId)
			continue
		}

		if !isEntityTarget {
			stopDistance = 0
		}
		checker := s.collision()
		if !pathingComponent.HasPlan(pathToPosition, stopDistance) || !checker.CanStep(positionPos, *path.Peek()) {
			newPath, err := checker.GetPathWithinRange(positionPos, pathToPosition, stopDistance)
			if err != nil {
				s.rejectPathing(entityId)
				continue
			}
			pathingComponent.SetPlan(&newPath, pathToPosition, stopDistance)
			path = &newPath
		}

		nextPosition := path.Pop()
		if nextPosition != nil {
			s.entityStateTransitions(s.TickSource).BeginMoving(entityId)
			positionComponent.SetPosition(*nextPosition)
			s.ComponentManager.SetEntityComponent(entityId, positionComponent)
		}
	}
}

func (s *PathingSystem) rejectPathing(entityId model.EntityId) {
	s.entityStateTransitions(s.TickSource).RejectPathing(entityId)
	s.sendChatMessage(entityId, pathNotFoundMessage)
}

func (s *PathingSystem) sendChatMessage(fromEntityId model.EntityId, message string) {
	if s.EventEmitter == nil {
		return
	}
	s.EventEmitter.EmitGameEvent(gameevent.NewChatSpoken(fromEntityId, message))
}

func (s *PathingSystem) resolveOverlap(
	entityId model.EntityId,
	targetPosition math.Vec2,
) {
	directions := []math.Vec2{
		{X: 1, Y: 0},
		{X: -1, Y: 0},
		{X: 0, Y: 1},
		{X: 0, Y: -1},
	}

	for _, direction := range directions {
		candidate := targetPosition.Add(direction)
		if s.collision().IsBlocked(candidate.X, candidate.Y) {
			continue
		}
		positionComponent := s.ComponentManager.GetEntityComponent(component.ComponentIdPosition, entityId).(*component.CPosition)
		s.entityStateTransitions(s.TickSource).BeginMoving(entityId)
		positionComponent.SetPosition(candidate)
		s.ComponentManager.SetEntityComponent(entityId, positionComponent)
		return
	}
}

func (s *PathingSystem) collision() collision.Checker {
	return collision.Checker{
		World:            s.World,
		ComponentManager: s.ComponentManager,
		SpatialIndex:     s.SpatialIndex,
	}
}

func manhattanDistance(a math.Vec2, b math.Vec2) int {
	dx := a.X - b.X
	if dx < 0 {
		dx = -dx
	}
	dy := b.Y - a.Y
	if dy < 0 {
		dy = -dy
	}
	return dx + dy
}

func absInt(value int) int {
	if value < 0 {
		return -value
	}
	return value
}
