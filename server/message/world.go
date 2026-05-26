package message

import "webscape/server/game/world"

type worldData struct {
	SizeX    int               `json:"sizeX"`
	SizeY    int               `json:"sizeY"`
	Terrain  []string          `json:"terrain"`
	Blockers [][]bool          `json:"blockers"`
	Walls    []world.WorldWall `json:"walls"`
}

func NewWorldMessage(world *world.World) Message {
	return newMessage(
		MessageTypeWorld,
		worldData{
			SizeX:    world.GetSizeX(),
			SizeY:    world.GetSizeY(),
			Terrain:  world.GetTerrain(),
			Blockers: world.GetBlockers(),
			Walls:    world.GetWalls(),
		},
	)
}
