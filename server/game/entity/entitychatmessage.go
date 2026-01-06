package entity

import (
	"webscape/server/game/component"
	"webscape/server/game/model"
)

const ChatMessageTtl = 10

func CreateChatMessageEntity(fromEntityId model.EntityId, message string) []component.Component {
	chatMessageComponent := &component.CChatMessage{
		FromEntityId: fromEntityId,
		Message:      message,
	}
	renderableComponent := &component.CRenderable{
		Type: "chatmessage",
	}
	ttlComponent := &component.CTtl{
		Remaining: ChatMessageTtl,
	}

	return []component.Component{
		chatMessageComponent,
		renderableComponent,
		ttlComponent,
	}
}
