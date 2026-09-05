package system

import (
	"testing"
	"testing/fstest"
	"webscape/server/game/component"
	"webscape/server/game/gameevent"
	"webscape/server/game/model"
	"webscape/server/game/spatial"
	"webscape/server/game/world"
	"webscape/server/math"
	"webscape/server/util"
)

func TestPathingSystemRejectsUnreachablePathOnce(t *testing.T) {
	testWorld := loadPathingTestWorld(t)
	componentManager := component.NewComponentManager()
	targetEntityId := model.NewEntityId()
	entityId := componentManager.CreateNewEntity(
		component.NewCPosition(math.Vec2{X: 0, Y: 0}),
		component.NewCPathing(component.PathingTarget{
			Position: util.OptionalSome(math.Vec2{X: 2, Y: 0}),
		}),
		component.NewCInteracting(targetEntityId, component.InteractionOptionTalk),
		component.NewCCombatState(targetEntityId),
		component.NewCWoodcutting(targetEntityId, 0, math.Vec2{X: 0, Y: 0}),
		component.NewCFishing(targetEntityId, 0, math.Vec2{X: 0, Y: 0}),
	)
	emitter := &recordingEventEmitter{}

	pathingSystem := PathingSystem{
		SystemBase: SystemBase{
			ComponentManager: componentManager,
		},
		World:        testWorld,
		EventEmitter: emitter,
	}

	pathingSystem.Update()

	if componentManager.GetEntityComponent(component.ComponentIdPathing, entityId) != nil {
		t.Fatal("pathing component was not removed after unreachable path")
	}
	if componentManager.GetEntityComponent(component.ComponentIdInteracting, entityId) != nil {
		t.Fatal("interacting component was not removed after unreachable path")
	}
	if componentManager.GetEntityComponent(component.ComponentIdCombatState, entityId) != nil {
		t.Fatal("combat state component was not removed after unreachable path")
	}
	if componentManager.GetEntityComponent(component.ComponentIdWoodcutting, entityId) != nil {
		t.Fatal("woodcutting component was not removed after unreachable path")
	}
	if componentManager.GetEntityComponent(component.ComponentIdFishing, entityId) != nil {
		t.Fatal("fishing component was not removed after unreachable path")
	}
	assertPathNotFoundChatEvent(t, emitter.events, entityId, 1)

	pathingSystem.Update()

	assertPathNotFoundChatEvent(t, emitter.events, entityId, 1)
}

func loadPathingTestWorld(t *testing.T) *world.World {
	t.Helper()

	testWorld, err := world.LoadFromGameFS(fstest.MapFS{
		"game.json": {
			Data: []byte(`{
				"formatVersion": 2,
				"id": "test_game",
				"world": { "chunkSize": { "x": 3, "y": 1 } },
				"files": {
					"chunks": ["chunks/test.json"],
					"conversations": [],
					"quests": []
				}
			}`),
		},
		"chunks/test.json": {
			Data: []byte(`{
				"formatVersion": 2,
				"id": "test",
				"coordinate": { "x": 0, "y": 0 },
				"terrain": ["grass", "grass", "grass"],
				"heights": [0, 0, 0],
				"blockers": [false, true, false],
				"entities": [{"id":"player_spawn","components":{"position":{"x":0,"y":0},"playerSpawn":{}}}]
			}`),
		},
	})
	if err != nil {
		t.Fatalf("LoadFromGameFS returned error: %v", err)
	}
	return testWorld
}

func assertPathNotFoundChatEvent(
	t *testing.T,
	events []gameevent.Event,
	fromEntityId model.EntityId,
	wantCount int,
) {
	t.Helper()

	chatEventCount := 0
	for _, event := range events {
		if event.Id != gameevent.EventIdChatSpoken || event.ActorEntityId != fromEntityId {
			continue
		}
		chatEventCount++
		payload, ok := event.Payload.(gameevent.ChatSpokenPayload)
		if !ok || payload.Message != pathNotFoundMessage {
			t.Fatalf("chat event payload = %#v, want %q", event.Payload, pathNotFoundMessage)
		}
	}

	if chatEventCount != wantCount {
		t.Fatalf("path not found chat event count = %d, want %d", chatEventCount, wantCount)
	}
}

type countingPathIndex struct {
	*spatial.Index
	checks int
}

func (i *countingPathIndex) BlocksMovement(x, y int) bool {
	i.checks++
	return i.Index.BlocksMovement(x, y)
}

func TestPathingReusesPlanAndReplansWhenTargetMoves(t *testing.T) {
	manager := component.NewComponentManager()
	w := world.NewWorld(16, 16)
	index := &countingPathIndex{Index: spatial.NewIndex(w, manager)}
	target := manager.CreateNewEntity(component.NewCPosition(math.Vec2{X: 12, Y: 0}))
	pathing := component.NewCPathing(component.PathingTarget{EntityId: util.OptionalSome(target)})
	id := manager.CreateNewEntity(component.NewCPosition(math.Vec2{}), pathing)
	s := PathingSystem{SystemBase: SystemBase{ComponentManager: manager}, World: w, SpatialIndex: index}
	s.Update()
	if pathing.GetPath() == nil {
		t.Fatal("no initial path")
	}
	plan := pathing.GetPath()
	index.checks = 0
	s.Update()
	if pathing.GetPath() != plan || index.checks > 3 {
		t.Fatalf("did not reuse path: checks=%d", index.checks)
	}
	// Combat must preserve the path component while chasing the same target.
	combat := CombatSystem{SystemBase: SystemBase{ComponentManager: manager}}
	combat.setPathingToEntity(id, target)
	if manager.GetEntityComponent(component.ComponentIdPathing, id) != pathing {
		t.Fatal("combat discarded cached path")
	}
	targetPosition := manager.GetEntityComponent(component.ComponentIdPosition, target).(*component.CPosition)
	targetPosition.SetPosition(math.Vec2{X: 12, Y: 6})
	manager.SetEntityComponent(target, targetPosition)
	s.Update()
	if pathing.GetPath() == plan {
		t.Fatal("target movement did not replan")
	}
	manager.RemoveEntity(target)
	s.Update()
	if manager.GetEntityComponent(component.ComponentIdPathing, id) != nil {
		t.Fatal("missing target did not cancel path")
	}
}

func TestCachedPathReplansWhenDoorCloses(t *testing.T) {
	manager := component.NewComponentManager()
	w := world.NewWorld(8, 3)
	index := spatial.NewIndex(w, manager)
	door := manager.CreateNewEntity(component.NewCPosition(math.Vec2{X: 2, Y: 1}), component.NewCMetadata(util.JObject{"blocksMovement": util.JBool(true)}), component.NewCOpenable(true))
	pathing := component.NewCPathing(component.PathingTarget{Position: util.OptionalSome(math.Vec2{X: 7, Y: 1})})
	position := component.NewCPosition(math.Vec2{X: 0, Y: 1})
	manager.CreateNewEntity(position, pathing)
	s := PathingSystem{SystemBase: SystemBase{ComponentManager: manager}, World: w, SpatialIndex: index}
	s.Update()
	original := pathing.GetPath()
	if position.GetPosition() != (math.Vec2{X: 1, Y: 1}) {
		t.Fatal("unexpected first step")
	}
	openable := manager.GetEntityComponent(component.ComponentIdOpenable, door).(*component.COpenable)
	openable.SetOpen(false)
	manager.SetEntityComponent(door, openable)
	s.Update()
	if pathing.GetPath() == original || position.GetPosition() == (math.Vec2{X: 2, Y: 1}) {
		t.Fatal("cached path walked through closed door")
	}
	for range 12 {
		s.Update()
	}
	if position.GetPosition() != (math.Vec2{X: 7, Y: 1}) {
		t.Fatal("did not reach destination around door")
	}
}
