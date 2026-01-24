package system

import (
	"webscape/server/game/component"
	"webscape/server/math"
)

type HealthSystem struct {
	SystemBase
}

func (s *HealthSystem) Update() {
	entityIds := s.ComponentManager.GetEntitiesWithComponents(component.ComponentIdHealth)

	for _, entityId := range entityIds {
		healthComponent := s.ComponentManager.GetEntityComponent(component.ComponentIdHealth, entityId).(*component.CHealth)

		// If health is zero or below, mark entity for removal
		if healthComponent.GetCurrentHealth() <= 0 {
			if s.ComponentManager.GetEntityComponent(component.ComponentIdPlayer, entityId) != nil {
				healthComponent.SetCurrentHealth(healthComponent.GetMaxHealth())
				s.ComponentManager.SetEntityComponent(entityId, healthComponent)

				positionComponent := s.ComponentManager.GetEntityComponent(component.ComponentIdPosition, entityId)
				if positionComponent != nil {
					position := positionComponent.(*component.CPosition)
					position.SetPosition(math.Vec2{X: 0, Y: 0})
					s.ComponentManager.SetEntityComponent(entityId, position)
				}

				s.ComponentManager.RemoveComponent(component.ComponentIdCombatState, entityId)
				s.ComponentManager.RemoveComponent(component.ComponentIdPathing, entityId)
				s.ComponentManager.RemoveComponent(component.ComponentIdInteracting, entityId)

				combatLog := s.ComponentManager.GetEntityComponent(component.ComponentIdCombatLog, entityId)
				if combatLog != nil {
					combatLog.(*component.CCombatLog).Clear()
				}
			} else {
				// Remove all components from the entity (this effectively removes it)
				s.ComponentManager.RemoveEntity(entityId)
			}
		}
	}
}
