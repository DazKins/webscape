package component

import (
	"webscape/server/game/model"
	"webscape/server/util"
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

func (c *CChatMessage) Serialize() util.Json {
	return util.JObject(map[string]util.Json{
		"fromEntityId": util.JString(c.FromEntityId.String()),
		"message":      util.JString(c.Message),
	})
}
