package system

import (
	"webscape/server/game/component"
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
			// Remove all components from the entity (this effectively removes it)
			s.ComponentManager.RemoveEntity(entityId)
		}
	}
}
