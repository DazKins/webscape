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
	renderableComponent := &component.CRenderable{
		Type: "chatmessage",
	}

	return NewEntity(model.NewEntityId()).
		SetComponent(chatMessageComponent).
		SetComponent(renderableComponent)
}
