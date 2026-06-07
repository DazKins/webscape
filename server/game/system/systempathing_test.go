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
	assertPathNotFoundChatMessage(t, componentManager, entityId, 1)

	pathingSystem.Update()

	assertPathNotFoundChatMessage(t, componentManager, entityId, 1)
}

func loadPathingTestWorld(t *testing.T) *world.World {
	t.Helper()

	testWorld, err := world.LoadFromGameFS(fstest.MapFS{
		"game.json": {
			Data: []byte(`{
				"formatVersion": 1,
				"id": "test_game",
				"files": {
					"maps": ["maps/test.json"],
					"conversations": [],
					"quests": []
				}
			}`),
		},
		"maps/test.json": {
			Data: []byte(`{
				"formatVersion": 1,
				"id": "test",
				"size": { "x": 3, "y": 1 },
				"terrain": ["grass", "grass", "grass"],
				"blockers": [false, true, false],
				"entities": []
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
