package message

import "webscape/server/game/model"

type entityRemoveData struct {
	Id string `json:"id"`
}

func NewEntityRemoveMessage(entityId model.EntityId) Message {
	return newMessage(
		MessageTypeEntityRemove,
		entityRemoveData{
			Id: entityId.String(),
		},
	)
}
