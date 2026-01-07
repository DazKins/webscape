package component

import (
	"webscape/server/game/model"
	"webscape/server/util"
)

const ComponentIdChatMessage = ComponentId("chatmessage")

type CChatMessage struct {
	fromEntityId model.EntityId
	message      string
}

func NewCChatMessage(fromEntityId model.EntityId, message string) *CChatMessage {
	return &CChatMessage{
		fromEntityId: fromEntityId,
		message:      message,
	}
}

func (c *CChatMessage) GetId() ComponentId {
	return ComponentIdChatMessage
}

func (c *CChatMessage) Serialize() util.Json {
	return util.JObject(map[string]util.Json{
		"fromEntityId": util.JString(c.fromEntityId.String()),
		"message":      util.JString(c.message),
	})
}

func (c *CChatMessage) GetFromEntityId() model.EntityId {
	return c.fromEntityId
}

func (c *CChatMessage) SetFromEntityId(fromEntityId model.EntityId) {
	c.fromEntityId = fromEntityId
}

func (c *CChatMessage) GetMessage() string {
	return c.message
}

func (c *CChatMessage) SetMessage(message string) {
	c.message = message
}
