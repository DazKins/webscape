package message

import (
	"webscape/server/game/entity"
)

type entityUpdateData struct {
	EntityId  string `json:"entityId"`
	PositionX int    `json:"positionX"`
	PositionY int    `json:"positionY"`
	Color     string `json:"color"`
}

func NewEntityUpdateMessage(entity *entity.Entity) Message {
	return newMessage(
		MessageTypeEntityUpdate,
		entityUpdateData{
			EntityId:  entity.ID.String(),
			PositionX: entity.Position.X,
			PositionY: entity.Position.Y,
			Color:     entity.Color,
		},
	)
}
