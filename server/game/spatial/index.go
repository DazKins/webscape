package spatial

import (
	"webscape/server/game/component"
	"webscape/server/game/model"
	"webscape/server/game/world"
	"webscape/server/util"
)

// Index maps entity footprints to every chunk they occupy. Reindex is cheap for
// normal entities and is invoked synchronously by ComponentManager mutations.
type Index struct {
	world    *world.World
	manager  *component.ComponentManager
	byChunk  map[world.ChunkCoord]map[model.EntityId]bool
	byEntity map[model.EntityId]map[world.ChunkCoord]bool
}

func NewIndex(gameWorld *world.World, manager *component.ComponentManager) *Index {
	index := &Index{world: gameWorld, manager: manager, byChunk: make(map[world.ChunkCoord]map[model.EntityId]bool), byEntity: make(map[model.EntityId]map[world.ChunkCoord]bool)}
	manager.SetEntityChangedHandler(index.Reindex)
	for entityID := range manager.GetComponent(component.ComponentIdPosition) {
		index.Reindex(entityID)
	}
	return index
}

func (i *Index) Reindex(entityID model.EntityId) {
	for coord := range i.byEntity[entityID] {
		delete(i.byChunk[coord], entityID)
		if len(i.byChunk[coord]) == 0 {
			delete(i.byChunk, coord)
		}
	}
	delete(i.byEntity, entityID)
	positionComponent := i.manager.GetEntityComponent(component.ComponentIdPosition, entityID)
	if positionComponent == nil {
		return
	}
	position := positionComponent.(*component.CPosition).GetPosition()
	width, height := entitySize(i.manager, entityID)
	minCoord, _ := i.world.GlobalToChunk(position.X, position.Y)
	maxCoord, _ := i.world.GlobalToChunk(position.X+width-1, position.Y+height-1)
	occupied := make(map[world.ChunkCoord]bool)
	for y := minCoord.Y; y <= maxCoord.Y; y++ {
		for x := minCoord.X; x <= maxCoord.X; x++ {
			coord := world.ChunkCoord{X: x, Y: y}
			if i.byChunk[coord] == nil {
				i.byChunk[coord] = make(map[model.EntityId]bool)
			}
			i.byChunk[coord][entityID] = true
			occupied[coord] = true
		}
	}
	i.byEntity[entityID] = occupied
}

func (i *Index) EntitiesInChunk(coord world.ChunkCoord) []model.EntityId {
	result := make([]model.EntityId, 0, len(i.byChunk[coord]))
	for entityID := range i.byChunk[coord] {
		result = append(result, entityID)
	}
	return result
}

func (i *Index) EntitiesInChunks(coords map[world.ChunkCoord]bool) map[model.EntityId]bool {
	result := make(map[model.EntityId]bool)
	for coord := range coords {
		for entityID := range i.byChunk[coord] {
			result[entityID] = true
		}
	}
	return result
}

func (i *Index) EntitiesAt(x, y int) []model.EntityId {
	coord, _ := i.world.GlobalToChunk(x, y)
	return i.EntitiesInChunk(coord)
}

func (i *Index) EntityChunks(entityID model.EntityId) map[world.ChunkCoord]bool {
	result := make(map[world.ChunkCoord]bool)
	for coord := range i.byEntity[entityID] {
		result[coord] = true
	}
	return result
}

func entitySize(manager *component.ComponentManager, entityID model.EntityId) (int, int) {
	width, height := 1, 1
	metadataComponent := manager.GetEntityComponent(component.ComponentIdMetadata, entityID)
	if metadataComponent == nil {
		return width, height
	}
	metadata, ok := metadataComponent.(*component.CMetadata).GetMetadata().(util.JObject)
	if !ok {
		return width, height
	}
	if value, ok := metadata["width"].(util.JNumber); ok && value >= 1 {
		width = int(value)
	}
	if value, ok := metadata["height"].(util.JNumber); ok && value >= 1 {
		height = int(value)
	}
	return width, height
}
