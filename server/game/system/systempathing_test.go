package system

import (
	"testing"
	"testing/fstest"
	"webscape/server/game/component"
	"webscape/server/game/model"
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
		component.NewCWoodcutting(targetEntityId, math.Vec2{X: 0, Y: 0}),
	)

	pathingSystem := PathingSystem{
		SystemBase: SystemBase{
			ComponentManager: componentManager,
		},
		World: testWorld,
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
	assertPathNotFoundChatMessage(t, componentManager, entityId, 1)

	pathingSystem.Update()

	assertPathNotFoundChatMessage(t, componentManager, entityId, 1)
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

func assertPathNotFoundChatMessage(
	t *testing.T,
	componentManager *component.ComponentManager,
	fromEntityId model.EntityId,
	wantCount int,
) {
	t.Helper()

	chatMessageCount := 0
	for _, comp := range componentManager.GetComponent(component.ComponentIdChatMessage) {
		chatMessage := comp.(*component.CChatMessage)
		if chatMessage.GetFromEntityId() != fromEntityId {
			continue
		}
		chatMessageCount++
		if chatMessage.GetMessage() != pathNotFoundMessage {
			t.Fatalf("chat message = %q, want %q", chatMessage.GetMessage(), pathNotFoundMessage)
		}
	}

	if chatMessageCount != wantCount {
		t.Fatalf("path not found chat message count = %d, want %d", chatMessageCount, wantCount)
	}
}
