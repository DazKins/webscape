package spatial

import (
	"testing"
	"webscape/server/game/component"
	"webscape/server/game/world"
	"webscape/server/math"
	"webscape/server/util"
)

func TestIndexTracksFootprintsAcrossSeamsMovementResizeAndDeletion(t *testing.T) {
	gameWorld := world.NewWorld(4, 4)
	manager := component.NewComponentManager()
	index := NewIndex(gameWorld, manager)
	entityID := manager.CreateNewEntity(
		component.NewCPosition(math.Vec2{X: 3, Y: 1}),
		component.NewCMetadata(util.JObject{"width": util.JNumber(2), "height": util.JNumber(1)}),
	)
	assertIndexed(t, index, world.ChunkCoord{X: 0, Y: 0}, entityID, true)
	assertIndexed(t, index, world.ChunkCoord{X: 1, Y: 0}, entityID, true)

	position := manager.GetEntityComponent(component.ComponentIdPosition, entityID).(*component.CPosition)
	position.SetPosition(math.Vec2{X: -1, Y: 1})
	manager.SetEntityComponent(entityID, position)
	assertIndexed(t, index, world.ChunkCoord{X: -1, Y: 0}, entityID, true)
	assertIndexed(t, index, world.ChunkCoord{X: 0, Y: 0}, entityID, true)
	assertIndexed(t, index, world.ChunkCoord{X: 1, Y: 0}, entityID, false)

	manager.SetEntityComponent(entityID, component.NewCMetadata(util.JObject{"width": util.JNumber(1), "height": util.JNumber(1)}))
	assertIndexed(t, index, world.ChunkCoord{X: 0, Y: 0}, entityID, false)
	manager.RemoveEntity(entityID)
	assertIndexed(t, index, world.ChunkCoord{X: -1, Y: 0}, entityID, false)
}

func assertIndexed(t *testing.T, index *Index, coord world.ChunkCoord, entityID interface{ String() string }, want bool) {
	t.Helper()
	found := false
	for _, candidate := range index.EntitiesInChunk(coord) {
		if candidate.String() == entityID.String() {
			found = true
		}
	}
	if found != want {
		t.Fatalf("entity membership in %#v = %v, want %v", coord, found, want)
	}
}
