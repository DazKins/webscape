package system

import "webscape/server/game/component"

type ChatMessageSystem struct {
	SystemBase
}

func (s *ChatMessageSystem) Update() {
	entityIds := s.ComponentManager.GetEntitiesWithComponents(component.ComponentIdChatMessage)

	for _, entityId := range entityIds {
		chatMessageComponent := s.ComponentManager.GetEntityComponent(component.ComponentIdChatMessage, entityId).(*component.CChatMessage)
		chatMessageComponent.Ttl -= 1
		if chatMessageComponent.Ttl <= 0 {
			s.ComponentManager.RemoveComponent(component.ComponentIdChatMessage, entityId)
			s.ComponentManager.RemoveComponent(component.ComponentIdRenderable, entityId)
		}
	}
}
