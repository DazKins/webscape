package system

import (
	"webscape/server/game/collision"
	"webscape/server/game/component"
	"webscape/server/game/entity"
	"webscape/server/game/model"
	"webscape/server/game/world"
	"webscape/server/math"
)

const pathNotFoundMessage = "I can't find a way there!"

type PathingSystem struct {
	SystemBase
	World *world.World
}

func (s *PathingSystem) Update() {
	entityIds := s.ComponentManager.GetEntitiesWithComponents(component.ComponentIdPosition, component.ComponentIdPathing)

	for _, entityId := range entityIds {
		positionComponent := s.ComponentManager.GetEntityComponent(component.ComponentIdPosition, entityId).(*component.CPosition)
		pathingComponent := s.ComponentManager.GetEntityComponent(component.ComponentIdPathing, entityId).(*component.CPathing)

		target := pathingComponent.GetTarget()
		path := pathingComponent.GetPath()

		if path != nil && path.Size() == 0 {
			s.ComponentManager.RemoveComponent(component.ComponentIdPathing, entityId)
			continue
		}

		pathToPosition := math.Vec2{}
		isEntityTarget := false
		inCombat := s.ComponentManager.GetEntityComponent(component.ComponentIdCombatState, entityId) != nil

		if target.Position.IsPresent() {
			pathToPosition = target.Position.Unwrap()
		} else if target.EntityId.IsPresent() {
			targetEntityId := target.EntityId.Unwrap()
			targetEntityPosition := s.ComponentManager.GetEntityComponent(component.ComponentIdPosition, targetEntityId).(*component.CPosition)
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
			if inCombat {
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

		if isEntityTarget {
			stopPosition, ok := s.findEntityStopPosition(positionPos, pathToPosition, stopDistance)
			if ok {
				pathToPosition = stopPosition
			}
		}

		// It's easiest just to recompute the path every time for now...
		newPath, err := s.collision().GetPath(positionPos, pathToPosition)
		if err != nil {
			s.rejectPathing(entityId)
			continue
		}
		pathingComponent.SetPath(&newPath)
		path = &newPath

		nextPosition := path.Pop()
		if nextPosition != nil {
			positionComponent.SetPosition(*nextPosition)
		}
	}
}

func (s *PathingSystem) rejectPathing(entityId model.EntityId) {
	s.ComponentManager.RemoveComponent(component.ComponentIdPathing, entityId)
	s.ComponentManager.RemoveComponent(component.ComponentIdInteracting, entityId)
	s.ComponentManager.RemoveComponent(component.ComponentIdCombatState, entityId)
	s.sendChatMessage(entityId, pathNotFoundMessage)
}

func (s *PathingSystem) sendChatMessage(fromEntityId model.EntityId, message string) {
	chatMessageEntities := s.ComponentManager.GetComponent(component.ComponentIdChatMessage)
	for existingEntityId, comp := range chatMessageEntities {
		chatMessageComp := comp.(*component.CChatMessage)
		if chatMessageComp.GetFromEntityId() == fromEntityId {
			s.ComponentManager.RemoveEntity(existingEntityId)
		}
	}

	s.ComponentManager.CreateNewEntity(entity.CreateChatMessageEntity(fromEntityId, message)...)
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
		positionComponent.SetPosition(candidate)
		return
	}
}

func (s *PathingSystem) findEntityStopPosition(
	currentPosition math.Vec2,
	targetPosition math.Vec2,
	stopDistance int,
) (math.Vec2, bool) {
	bestPos := math.Vec2{}
	bestLen := -1

	for dx := -stopDistance; dx <= stopDistance; dx++ {
		for dy := -stopDistance; dy <= stopDistance; dy++ {
			if absInt(dx)+absInt(dy) > stopDistance {
				continue
			}
			if dx == 0 && dy == 0 {
				continue
			}
			candidate := math.Vec2{X: targetPosition.X + dx, Y: targetPosition.Y + dy}
			if s.collision().IsBlocked(candidate.X, candidate.Y) {
				continue
			}
			path, err := s.collision().GetPath(currentPosition, candidate)
			if err != nil {
				continue
			}
			if bestLen == -1 || path.Size() < bestLen {
				bestLen = path.Size()
				bestPos = candidate
			}
		}
	}

	if bestLen == -1 {
		return math.Vec2{}, false
	}
	return bestPos, true
}

func (s *PathingSystem) collision() collision.Checker {
	return collision.Checker{
		World:            s.World,
		ComponentManager: s.ComponentManager,
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
