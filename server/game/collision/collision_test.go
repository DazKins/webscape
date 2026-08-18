package collision

import (
	"testing"
	"testing/fstest"
	"webscape/server/game/component"
	"webscape/server/game/world"
	"webscape/server/math"
	"webscape/server/util"
)

func TestPathCrossesAdjacentChunkSeamButNotVoid(t *testing.T) {
	gameWorld, err := world.LoadFromGameFS(fstest.MapFS{
		"game.json":     {Data: []byte(`{"formatVersion":2,"id":"seams","world":{"chunkSize":{"x":2,"y":2}},"files":{"chunks":["chunks/a.json","chunks/b.json"],"conversations":[],"quests":[]}}`)},
		"chunks/a.json": {Data: []byte(`{"formatVersion":2,"id":"a","coordinate":{"x":0,"y":0},"terrain":["grass","grass","grass","grass"],"heights":[0,0,0,0],"blockers":[false,false,false,false],"walls":[],"entities":[{"id":"player_spawn","components":{"position":{"x":0,"y":0},"playerSpawn":{}}}]}`)},
		"chunks/b.json": {Data: []byte(`{"formatVersion":2,"id":"b","coordinate":{"x":1,"y":0},"terrain":["grass","grass","grass","grass"],"heights":[0,0,0,0],"blockers":[false,false,false,false],"walls":[],"entities":[]}`)},
	})
	if err != nil {
		t.Fatal(err)
	}
	checker := Checker{World: gameWorld}
	if _, err := checker.GetPath(math.Vec2{X: 1, Y: 0}, math.Vec2{X: 2, Y: 0}); err != nil {
		t.Fatalf("adjacent seam path failed: %v", err)
	}
	if _, err := checker.GetPath(math.Vec2{X: 0, Y: 0}, math.Vec2{X: -1, Y: 0}); err == nil {
		t.Fatal("path entered missing chunk")
	}
}

func TestOpenableClosedDoorBlocksAndOpenDoorPermitsPath(t *testing.T) {
	componentManager := component.NewComponentManager()
	doorEntityId := componentManager.CreateNewEntity(
		component.NewCPosition(math.Vec2{X: 1, Y: 0}),
		component.NewCMetadata(util.JObject(map[string]util.Json{
			"blocksMovement": util.JBool(true),
		})),
		component.NewCOpenable(false),
	)

	checker := Checker{
		World:            newPathTestWorld(),
		ComponentManager: componentManager,
	}

	if _, err := checker.GetPath(math.Vec2{X: 0, Y: 0}, math.Vec2{X: 2, Y: 0}); err == nil {
		t.Fatal("closed door path succeeded, want no path")
	}

	openable := componentManager.GetEntityComponent(component.ComponentIdOpenable, doorEntityId).(*component.COpenable)
	openable.SetOpen(true)
	componentManager.SetEntityComponent(doorEntityId, openable)

	path, err := checker.GetPath(math.Vec2{X: 0, Y: 0}, math.Vec2{X: 2, Y: 0})
	if err != nil {
		t.Fatalf("open door path returned error: %v", err)
	}
	if path.Size() != 2 {
		t.Fatalf("open door path size = %d, want 2", path.Size())
	}
}

func newPathTestWorld() StaticWorld {
	return pathTestWorld{}
}

type pathTestWorld struct{}

func (pathTestWorld) GetStaticWall(x int, y int) bool {
	return x < 0 || x >= 3 || y != 0
}
