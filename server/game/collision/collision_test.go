package collision

import (
	"testing"
	"webscape/server/game/component"
	"webscape/server/math"
	"webscape/server/util"
)

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
