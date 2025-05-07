package message

import (
	"webscape/server/game/world"
)

type worldData struct {
	SizeX int `json:"sizeX"`
	SizeY int `json:"sizeY"`
}

func NewWorldMessage(world *world.World) Message {
	return newMessage(
		MessageTypeWorld,
		worldData{
			SizeX: world.GetSizeX(),
			SizeY: world.GetSizeY(),
		},
	)
}
