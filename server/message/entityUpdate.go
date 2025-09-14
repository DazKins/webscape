package message

import (
	"webscape/server/game/entity"
)

type entityUpdateData struct {
	EntityId  string `json:"entityId"`
	PositionX int    `json:"positionX"`
	PositionY int    `json:"positionY"`
	VelocityX int    `json:"velocityX"`
	VelocityY int    `json:"velocityY"`
	Color     string `json:"color"`
}

func NewEntityUpdateMessage(entity *entity.Entity) Message {
	return newMessage(
		MessageTypeEntityUpdate,
		entityUpdateData{
			EntityId:  entity.ID.String(),
			PositionX: entity.Position.X,
			PositionY: entity.Position.Y,
			VelocityX: entity.Velocity.X,
			VelocityY: entity.Velocity.Y,
			Color:     entity.Color,
		},
	)
}
