package entity

import (
	"webscape/server/game/component"
	"webscape/server/game/model"
)

const ChatMessageTtl = 10

func CreateChatMessageEntity(fromEntityId model.EntityId, message string) []component.Component {
	chatMessageComponent := component.NewCChatMessage(fromEntityId, message)
	renderableComponent := component.NewCRenderable("chatmessage")
	ttlComponent := component.NewCTtl(ChatMessageTtl)

	return []component.Component{
		chatMessageComponent,
		renderableComponent,
		ttlComponent,
	}
}
