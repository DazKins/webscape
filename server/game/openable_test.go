package game

import (
	"testing"
	"testing/fstest"
	"webscape/server/game/component"
	"webscape/server/game/model"
	"webscape/server/game/world"
	"webscape/server/message"
)

func TestOpenableDoorOffersOpenThenClose(t *testing.T) {
	testWorld := loadOpenableTestWorld(t, false)
	game := NewGameWithWorld(testWorld)
	game.RegisterBroadcaster(func(message.Message) {})
	game.RegisterSender(func(string, message.Message) {})

	doorEntityId, ok := firstEntityWithComponent(game, component.ComponentIdOpenable)
	if !ok {
		t.Fatal("no door entity has COpenable")
	}
	assertInteractionOptions(t, game.getInteractionOptionsForEntity(doorEntityId), component.InteractionOptionOpen)

	playerEntityId := model.NewEntityId()
	game.HandleJoin("client-1", playerEntityId, "player")
	game.HandleInteract("client-1", doorEntityId, component.InteractionOptionOpen)
	game.update()

	openable := game.componentManager.GetEntityComponent(component.ComponentIdOpenable, doorEntityId).(*component.COpenable)
	if !openable.IsOpen() {
		t.Fatal("door isOpen = false after open interaction, want true")
	}
	assertInteractionOptions(t, game.getInteractionOptionsForEntity(doorEntityId), component.InteractionOptionClose)

	game.HandleInteract("client-1", doorEntityId, component.InteractionOptionClose)
	game.update()

	if openable.IsOpen() {
		t.Fatal("door isOpen = true after close interaction, want false")
	}
	assertInteractionOptions(t, game.getInteractionOptionsForEntity(doorEntityId), component.InteractionOptionOpen)
}

func TestOpeningDoorEmitsQuestEvent(t *testing.T) {
	testWorld := loadOpenableTestWorld(t, true)
	game := NewGameWithWorld(testWorld)
	game.RegisterBroadcaster(func(message.Message) {})
	game.RegisterSender(func(string, message.Message) {})

	doorEntityId, ok := firstEntityWithComponent(game, component.ComponentIdOpenable)
	if !ok {
		t.Fatal("no door entity has COpenable")
	}

	playerEntityId := model.NewEntityId()
	game.HandleJoin("client-1", playerEntityId, "player")
	game.HandleInteract("client-1", doorEntityId, component.InteractionOptionOpen)
	game.update()

	assertQuestProgress(t, game, playerEntityId, "open_the_door", "", 0, 0, true)
}

func loadOpenableTestWorld(t *testing.T, includeQuest bool) *world.World {
	t.Helper()

	questPaths := "[]"
	questFile := fstest.MapFS{}
	if includeQuest {
		questPaths = `["quests/open_the_door.json"]`
		questFile["quests/open_the_door.json"] = &fstest.MapFile{
			Data: []byte(`{
				"formatVersion": 2,
				"id": "open_the_door_quests",
				"quests": [
					{
						"id": "open_the_door",
						"startEventId": "interact:object:door_001:open",
						"steps": [
							{
								"id": "open",
								"description": "Open the door.",
								"requirement": { "eventId": "interact:object:door_001:open", "count": 1 }
							}
						],
						"rewards": {
							"items": [
								{ "name": "Door Token", "type": "quest", "count": 1 }
							]
						}
					}
				]
			}`),
		}
	}

	gameFS := fstest.MapFS{
		"game.json": {
			Data: []byte(`{
				"formatVersion": 1,
				"id": "test_game",
				"files": {
					"maps": ["maps/test.json"],
					"conversations": [],
					"quests": ` + questPaths + `
				}
			}`),
		},
		"maps/test.json": {
			Data: []byte(`{
				"formatVersion": 1,
				"id": "test",
				"size": { "x": 2, "y": 1 },
				"terrain": ["grass", "grass"],
				"entities": [
					{
						"id": "player_spawn",
						"components": {
							"position": { "x": 0, "y": 0 },
							"playerSpawn": {}
						}
					},
					{
						"id": "door_001",
						"components": {
							"position": { "x": 1, "y": 0 },
							"metadata": {
								"objectId": "door_001",
								"name": "door_001",
								"type": "door",
								"width": 1,
								"height": 1,
								"blocksMovement": true
							},
							"renderable": { "type": "door" },
							"openable": { "isOpen": false }
						}
					}
				]
			}`),
		},
	}
	for path, file := range questFile {
		gameFS[path] = file
	}

	testWorld, err := world.LoadFromGameFS(gameFS)
	if err != nil {
		t.Fatalf("LoadFromGameFS returned error: %v", err)
	}
	return testWorld
}

func assertInteractionOptions(t *testing.T, options []component.InteractionOption, want component.InteractionOption) {
	t.Helper()

	if len(options) != 1 {
		t.Fatalf("interaction options = %#v, want only %q", options, want)
	}
	if options[0] != want {
		t.Fatalf("interaction option = %q, want %q", options[0], want)
	}
}
