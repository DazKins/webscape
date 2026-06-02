package message

import (
	"webscape/server/game/component"
	"webscape/server/game/model"
	"webscape/server/util"
)

type entityUpdateData struct {
	EntityId              string    `json:"entityId"`
	ComponentId           string    `json:"componentId"`
	Data                  util.Json `json:"data"`
	AvailableInteractions []string  `json:"availableInteractions"`
}

type gameUpdateData struct {
	Entities []entityUpdateData `json:"entities"`
}

func NewGameUpdateMessage(
	updatedComponents map[component.ComponentId]map[model.EntityId]util.Json,
	removedComponents map[component.ComponentId][]model.EntityId,
	availableInteractions map[model.EntityId][]component.InteractionOption,
) Message {
	entityUpdates := make([]entityUpdateData, 0)

	for componentId, entities := range updatedComponents {
		for entityId, data := range entities {
			entityUpdates = append(entityUpdates, entityUpdateData{
				EntityId:              entityId.String(),
				ComponentId:           componentId.String(),
				Data:                  data,
				AvailableInteractions: serializeAvailableInteractions(availableInteractions[entityId]),
			})
		}
	}

	for componentId, entities := range removedComponents {
		for _, entityId := range entities {
			entityUpdates = append(entityUpdates, entityUpdateData{
				EntityId:              entityId.String(),
				ComponentId:           componentId.String(),
				Data:                  util.JNull{},
				AvailableInteractions: serializeAvailableInteractions(availableInteractions[entityId]),
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

func serializeAvailableInteractions(options []component.InteractionOption) []string {
	result := make([]string, len(options))
	for i, option := range options {
		result[i] = string(option)
	}
	return result
}
