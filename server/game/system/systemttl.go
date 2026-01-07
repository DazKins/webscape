package system

import (
	"webscape/server/game/component"
)

type TtlSystem struct {
	SystemBase
}

func (s *TtlSystem) Update() {
	entityIds := s.ComponentManager.GetEntitiesWithComponents(component.ComponentIdTtl)

	for _, entityId := range entityIds {
		ttlComponent := s.ComponentManager.GetEntityComponent(component.ComponentIdTtl, entityId).(*component.CTtl)

		newRemaining := ttlComponent.GetRemaining() - 1
		ttlComponent.SetRemaining(newRemaining)
		if newRemaining <= 0 {
			// Remove all components from the entity (effectively removing it)
			s.ComponentManager.RemoveEntity(entityId)
		}
	}
}
