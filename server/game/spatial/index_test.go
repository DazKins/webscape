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

func TestTileBlockersTrackDoorsOverlapsAndFootprints(t *testing.T) {
	manager := component.NewComponentManager()
	index := NewIndex(world.NewWorld(4, 4), manager)
	metadata := component.NewCMetadata(util.JObject{"width": util.JNumber(2), "blocksMovement": util.JBool(true)})
	door := manager.CreateNewEntity(component.NewCPosition(math.Vec2{X: -1, Y: 1}), metadata, component.NewCOpenable(false))
	if !index.BlocksMovement(-1, 1) || !index.BlocksMovement(0, 1) || index.BlocksMovement(1, 1) {
		t.Fatal("wrong footprint occupancy")
	}
	if len(index.EntitiesAt(2, 1)) != 0 {
		t.Fatal("tile query included other chunk entities")
	}
	overlapping := manager.CreateNewEntity(component.NewCPosition(math.Vec2{X: 0, Y: 1}), component.NewCMetadata(util.JObject{"blocksMovement": util.JBool(true)}))
	openable := manager.GetEntityComponent(component.ComponentIdOpenable, door).(*component.COpenable)
	openable.SetOpen(true)
	manager.SetEntityComponent(door, openable)
	if index.BlocksMovement(-1, 1) || !index.BlocksMovement(0, 1) {
		t.Fatal("opening door lost overlapping blocker")
	}
	manager.RemoveEntity(overlapping)
	if index.BlocksMovement(0, 1) {
		t.Fatal("removed blocker remains")
	}
	openable.SetOpen(false)
	manager.SetEntityComponent(door, openable)
	position := manager.GetEntityComponent(component.ComponentIdPosition, door).(*component.CPosition)
	position.SetPosition(math.Vec2{X: 3, Y: 1})
	manager.SetEntityComponent(door, position)
	if index.BlocksMovement(-1, 1) || !index.BlocksMovement(4, 1) {
		t.Fatal("movement did not update tiles")
	}
	manager.SetEntityComponent(door, component.NewCMetadata(util.JObject{"blocksMovement": util.JBool(true)}))
	if !index.BlocksMovement(3, 1) || index.BlocksMovement(4, 1) {
		t.Fatal("resize did not update tiles")
	}
	manager.RemoveComponent(component.ComponentIdMetadata, door)
	if index.BlocksMovement(3, 1) {
		t.Fatal("metadata removal left blocker")
	}
	manager.RemoveComponent(component.ComponentIdPosition, door)
	if len(index.EntitiesAt(3, 1)) != 0 {
		t.Fatal("position removal left tile membership")
	}
}

func TestBlockerLookupDoesNotAllocate(t *testing.T) {
	manager := component.NewComponentManager()
	index := NewIndex(world.NewWorld(4, 4), manager)
	manager.CreateNewEntity(component.NewCPosition(math.Vec2{}), component.NewCMetadata(util.JObject{"blocksMovement": util.JBool(true)}))
	if allocations := testing.AllocsPerRun(100, func() {
		if !index.BlocksMovement(0, 0) {
			panic("missing blocker")
		}
	}); allocations != 0 {
		t.Fatalf("allocations=%f", allocations)
	}
}
