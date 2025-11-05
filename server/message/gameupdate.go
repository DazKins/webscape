package message

import (
	"webscape/server/game/component"
	"webscape/server/game/model"
	"webscape/server/util"
)

type entityUpdateData struct {
	EntityId    string    `json:"entityId"`
	ComponentId string    `json:"componentId"`
	Data        util.Json `json:"data"`
}

type gameUpdateData struct {
	Entities []entityUpdateData `json:"entities"`
}

func NewGameUpdateMessage(updatedComponents map[model.EntityId]map[component.ComponentId]util.Json) Message {
	entityUpdates := make([]entityUpdateData, 0)
	for entityId, components := range updatedComponents {
		for componentId, data := range components {
			entityUpdates = append(entityUpdates, entityUpdateData{
				EntityId:    entityId.String(),
				ComponentId: componentId.String(),
				Data:        data,
			})
		}
	}

	return newMessage(
		MessageTypeGameUpdate,
		gameUpdateData{
			Entities: entityUpdates,
		},
	)
}
