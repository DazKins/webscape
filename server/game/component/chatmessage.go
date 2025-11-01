package component

import (
	"webscape/server/game/model"
)

const ComponentIdChatMessage = ComponentId("chatmessage")

type CChatMessage struct {
	FromEntityId model.EntityId
	Message      string
	Ttl          uint64
}

func (c *CChatMessage) GetId() ComponentId {
	return ComponentIdChatMessage
}

func (c *CChatMessage) Serialize() map[string]any {
	return map[string]any{
		"fromEntityId": c.FromEntityId.String(),
		"message":      c.Message,
	}
}
