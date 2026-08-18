package message

import "webscape/server/game/world"

type chunkLoadData struct {
	Coordinate world.ChunkCoord  `json:"coordinate"`
	Terrain    []string          `json:"terrain"`
	Heights    []int             `json:"heights"`
	Walls      []world.WorldWall `json:"walls"`
}

type chunkUpdateData struct {
	Load   []chunkLoadData    `json:"load"`
	Unload []world.ChunkCoord `json:"unload"`
}

func NewChunkUpdateMessage(load []world.Chunk, unload []world.ChunkCoord) Message {
	loads := make([]chunkLoadData, len(load))
	for index, chunk := range load {
		loads[index] = chunkLoadData{Coordinate: chunk.Coordinate, Terrain: chunk.Terrain, Heights: chunk.Heights, Walls: chunk.Walls}
	}
	return newMessage(MessageTypeChunkUpdate, chunkUpdateData{Load: loads, Unload: unload})
}
