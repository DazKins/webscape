package message

import (
	"webscape/server/game/component"
	"webscape/server/game/model"
	"webscape/server/util"
)

type componentUpdateData struct {
	EntityId    string    `json:"entityId"`
	ComponentId string    `json:"componentId"`
	Data        util.Json `json:"data"`
}

func NewComponentUpdateMessage(entityId model.EntityId, componentId component.ComponentId, data util.Json) Message {
	return newMessage(
		MessageTypeComponentUpdate,
		componentUpdateData{
			EntityId:    entityId.String(),
			ComponentId: componentId.String(),
			Data:        data,
		},
	)
}
