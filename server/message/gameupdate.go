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

func NewGameUpdateMessage(
	updatedComponents map[component.ComponentId]map[model.EntityId]util.Json,
	removedComponents map[component.ComponentId][]model.EntityId,
) Message {
	entityUpdates := make([]entityUpdateData, 0)

	for componentId, entities := range updatedComponents {
		for entityId, data := range entities {
			entityUpdates = append(entityUpdates, entityUpdateData{
				EntityId:    entityId.String(),
				ComponentId: componentId.String(),
				Data:        data,
			})
		}
	}

	for componentId, entities := range removedComponents {
		for _, entityId := range entities {
			entityUpdates = append(entityUpdates, entityUpdateData{
				EntityId:    entityId.String(),
				ComponentId: componentId.String(),
				Data:        util.JNull{},
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
