package spatial

import (
	"webscape/server/game/component"
	"webscape/server/game/model"
	"webscape/server/game/world"
	"webscape/server/math"
	"webscape/server/util"
)

// Index maps entity footprints to every chunk they occupy. Reindex is cheap for
// normal entities and is invoked synchronously by ComponentManager mutations.
type Index struct {
	world      *world.World
	manager    *component.ComponentManager
	byChunk    map[world.ChunkCoord]map[model.EntityId]bool
	byEntity   map[model.EntityId]map[world.ChunkCoord]bool
	byTile     map[math.Vec2]map[model.EntityId]bool
	blockers   map[math.Vec2]int
	footprints map[model.EntityId]footprint
}

type footprint struct {
	position      math.Vec2
	width, height int
	blocks        bool
}

func NewIndex(gameWorld *world.World, manager *component.ComponentManager) *Index {
	index := &Index{world: gameWorld, manager: manager, byChunk: make(map[world.ChunkCoord]map[model.EntityId]bool), byEntity: make(map[model.EntityId]map[world.ChunkCoord]bool)}
	index.byTile = make(map[math.Vec2]map[model.EntityId]bool)
	index.blockers = make(map[math.Vec2]int)
	index.footprints = make(map[model.EntityId]footprint)
	manager.SetEntityChangedHandler(index.Reindex)
	for entityID := range manager.GetComponent(component.ComponentIdPosition) {
		index.Reindex(entityID)
	}
	return index
}

func (i *Index) Reindex(entityID model.EntityId) {
	positionComponent := i.manager.GetEntityComponent(component.ComponentIdPosition, entityID)
	next := footprint{}
	if positionComponent != nil {
		next.position = positionComponent.(*component.CPosition).GetPosition()
		next.width, next.height = entitySize(i.manager, entityID)
		next.blocks = entityBlocksMovement(i.manager, entityID)
	}
	if old, ok := i.footprints[entityID]; ok {
		if positionComponent != nil && old == next {
			return
		}
		i.updateTiles(entityID, old, false)
		delete(i.footprints, entityID)
	}

	for coord := range i.byEntity[entityID] {
		delete(i.byChunk[coord], entityID)
		if len(i.byChunk[coord]) == 0 {
			delete(i.byChunk, coord)
		}
	}
	delete(i.byEntity, entityID)
	if positionComponent == nil {
		return
	}
	i.footprints[entityID] = next
	i.updateTiles(entityID, next, true)
	position := next.position
	width, height := next.width, next.height
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
	entities := i.byTile[math.Vec2{X: x, Y: y}]
	result := make([]model.EntityId, 0, len(entities))
	for id := range entities {
		result = append(result, id)
	}
	return result
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

// BlocksMovement is an allocation-free lookup, maintained by ECS mutations.
func (i *Index) BlocksMovement(x, y int) bool { return i.blockers[math.Vec2{X: x, Y: y}] > 0 }

func (i *Index) updateTiles(id model.EntityId, f footprint, add bool) {
	for x := f.position.X; x < f.position.X+f.width; x++ {
		for y := f.position.Y; y < f.position.Y+f.height; y++ {
			tile := math.Vec2{X: x, Y: y}
			if add {
				if i.byTile[tile] == nil {
					i.byTile[tile] = make(map[model.EntityId]bool)
				}
				i.byTile[tile][id] = true
				if f.blocks {
					i.blockers[tile]++
				}
			} else {
				delete(i.byTile[tile], id)
				if len(i.byTile[tile]) == 0 {
					delete(i.byTile, tile)
				}
				if f.blocks {
					i.blockers[tile]--
					if i.blockers[tile] == 0 {
						delete(i.blockers, tile)
					}
				}
			}
		}
	}
}

func entityBlocksMovement(manager *component.ComponentManager, id model.EntityId) bool {
	value := manager.GetEntityComponent(component.ComponentIdMetadata, id)
	if value == nil {
		return false
	}
	metadata, ok := value.(*component.CMetadata).GetMetadata().(util.JObject)
	if !ok || metadata["blocksMovement"] != util.JBool(true) {
		return false
	}
	openable := manager.GetEntityComponent(component.ComponentIdOpenable, id)
	return openable == nil || !openable.(*component.COpenable).IsOpen()
}
