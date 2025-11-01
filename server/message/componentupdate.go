package message

import (
	"webscape/server/game/component"
	"webscape/server/game/entity"
)

type componentUpdateData struct {
	EntityId    string         `json:"entityId"`
	ComponentId string         `json:"componentId"`
	Data        map[string]any `json:"data"`
}

func NewComponentUpdateMessage(entityId entity.EntityId, componentId component.ComponentId, data map[string]any) Message {
	return newMessage(
		MessageTypeComponentUpdate,
		componentUpdateData{
			EntityId:    entityId.String(),
			ComponentId: componentId.String(),
			Data:        data,
		},
	)
}
