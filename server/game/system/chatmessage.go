package system

import "webscape/server/game/component"

type ChatMessageSystem struct {
	SystemBase
}

func (s *ChatMessageSystem) Update() {
	entities := s.entityGetter.GetEntitiesWithComponents(component.ComponentIdChatMessage)

	for _, entity := range entities {
		chatMessageComponent := entity.GetComponent(component.ComponentIdChatMessage).(*component.CChatMessage)
		chatMessageComponent.Ttl -= 1
		if chatMessageComponent.Ttl <= 0 {
			entity.RemoveComponent(component.ComponentIdChatMessage)
		}
	}
}
