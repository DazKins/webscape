package entity

import (
	"webscape/server/game/component"
	"webscape/server/game/model"
)

const ChatMessageTtl = 10

func CreateChatMessageEntity(fromEntityId model.EntityId, message string) *Entity {
	chatMessageComponent := &component.CChatMessage{
		FromEntityId: fromEntityId,
		Message:      message,
		Ttl:          ChatMessageTtl,
	}

	return NewEntity(model.NewEntityId()).
		SetComponent(chatMessageComponent)
}
