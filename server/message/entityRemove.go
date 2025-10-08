package message

import "webscape/server/game/entity"

type entityRemoveData struct {
	Id string `json:"id"`
}

func NewEntityRemoveMessage(entityId entity.EntityId) Message {
	return newMessage(
		MessageTypeEntityRemove,
		entityRemoveData{
			Id: entityId.String(),
		},
	)
}
