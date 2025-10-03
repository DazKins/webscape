package message

import (
	"webscape/server/game/world"
)

type worldData struct {
	SizeX int      `json:"sizeX"`
	SizeY int      `json:"sizeY"`
	Walls [][]*int `json:"walls"`
}

func getWalls(walls [][]*world.WallDirection) [][]*int {
	wallsInt := make([][]*int, len(walls))
	for i := range walls {
		wallsInt[i] = make([]*int, len(walls[i]))
		for j := range walls[i] {
			if walls[i][j] != nil {
				wall := int(*walls[i][j])
				wallsInt[i][j] = &wall
			}
		}
	}
	return wallsInt
}

func NewWorldMessage(world *world.World) Message {
	return newMessage(
		MessageTypeWorld,
		worldData{
			SizeX: world.GetSizeX(),
			SizeY: world.GetSizeY(),
			Walls: getWalls(world.GetWalls()),
		},
	)
}
