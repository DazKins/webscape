package world

import (
	"testing"
	"testing/fstest"
)

func TestLoadFromGameFSLoadsFirstMap(t *testing.T) {
	gameFS := fstest.MapFS{
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
				"size": { "x": 2, "y": 1 },
				"terrain": ["grass", "road"],
				"entities": [
					{
						"id": "player_spawn",
						"components": {
							"position": { "x": 1, "y": 0 },
							"playerSpawn": {}
						}
					}
				]
			}`),
		},
	}

	world, err := LoadFromGameFS(gameFS)
	if err != nil {
		t.Fatalf("LoadFromGameFS returned error: %v", err)
	}

	if world.GetSizeX() != 2 || world.GetSizeY() != 1 {
		t.Fatalf("world size = (%d, %d), want (2, 1)", world.GetSizeX(), world.GetSizeY())
	}

	terrain := world.GetTerrain()
	if len(terrain) != 2 || terrain[0] != "grass" || terrain[1] != "road" {
		t.Fatalf("terrain = %#v, want grass/road", terrain)
	}
}

func TestLoadFromGameFSRejectsInvalidMapPath(t *testing.T) {
	gameFS := fstest.MapFS{
		"game.json": {
			Data: []byte(`{
				"formatVersion": 1,
				"id": "test_game",
				"files": {
					"maps": ["../outside.json"],
					"conversations": [],
					"quests": []
				}
			}`),
		},
	}

	if _, err := LoadFromGameFS(gameFS); err == nil {
		t.Fatal("LoadFromGameFS returned nil error for invalid map path")
	}
}

func TestLoadFromGameFSRejectsSpawnTemplatePosition(t *testing.T) {
	gameFS := fstest.MapFS{
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
				"size": { "x": 1, "y": 1 },
				"terrain": ["grass"],
				"entities": [
					{
						"id": "rat_spawn_001",
						"components": {
							"position": { "x": 0, "y": 0 },
							"spawn": {
								"respawnTicks": 20,
								"entity": {
									"components": {
										"position": { "x": 0, "y": 0 },
										"metadata": { "name": "Rat", "entityType": "rat" }
									}
								}
							}
						}
					}
				]
			}`),
		},
	}

	if _, err := LoadFromGameFS(gameFS); err == nil {
		t.Fatal("LoadFromGameFS returned nil error for spawn child template with position")
	}
}

func TestLoadFromGameFSLoadsConversationsAndAuthoredConversationComponent(t *testing.T) {
	gameFS := fstest.MapFS{
		"game.json": {
			Data: []byte(`{
				"formatVersion": 1,
				"id": "test_game",
				"files": {
					"maps": ["maps/test.json"],
					"conversations": ["conversations/greeting.json"],
					"quests": []
				}
			}`),
		},
		"maps/test.json": {
			Data: []byte(`{
				"formatVersion": 1,
				"id": "test",
				"size": { "x": 2, "y": 1 },
				"terrain": ["grass", "road"],
				"entities": [
					{
						"id": "player_spawn",
						"components": {
							"position": { "x": 0, "y": 0 },
							"playerSpawn": {}
						}
					},
					{
						"id": "greeter",
						"components": {
							"position": { "x": 1, "y": 0 },
							"conversation": { "conversationId": "greeting" }
						}
					}
				]
			}`),
		},
		"conversations/greeting.json": {
			Data: []byte(`{
				"formatVersion": 1,
				"id": "greeting_doc",
				"conversations": [
					{
						"id": "greeting",
						"startNodeId": "start",
						"nodes": [
							{
								"id": "start",
								"messages": [{ "text": "Hello." }],
								"endConversation": true
							}
						]
					}
				]
			}`),
		},
	}

	world, err := LoadFromGameFS(gameFS)
	if err != nil {
		t.Fatalf("LoadFromGameFS returned error: %v", err)
	}

	loadedConversation, ok := world.GetConversation("greeting")
	if !ok {
		t.Fatal("conversation greeting was not loaded")
	}
	if loadedConversation.StartNodeId != "start" {
		t.Fatalf("conversation startNodeId = %q, want start", loadedConversation.StartNodeId)
	}

	entities := world.GetEntities()
	if len(entities) != 2 {
		t.Fatalf("entities = %#v, want two authored entities", entities)
	}
	conversationComponent, ok := entities[1].Components["conversation"].(map[string]any)
	if !ok || conversationComponent["conversationId"] != "greeting" {
		t.Fatalf("entity conversation = %#v, want conversationId greeting", entities[1].Components["conversation"])
	}
}

func TestLoadFromGameFSLoadsQuests(t *testing.T) {
	gameFS := fstest.MapFS{
		"game.json": {
			Data: []byte(`{
				"formatVersion": 1,
				"id": "test_game",
				"files": {
					"maps": ["maps/test.json"],
					"conversations": [],
					"quests": ["quests/tutorial.json"]
				}
			}`),
		},
		"maps/test.json": {
			Data: []byte(`{
				"formatVersion": 1,
				"id": "test",
				"size": { "x": 1, "y": 1 },
				"terrain": ["grass"]
			}`),
		},
		"quests/tutorial.json": {
			Data: []byte(`{
				"formatVersion": 1,
				"id": "tutorial_quests",
				"quests": [
					{
						"id": "first_errand",
						"startEventId": "conversation:node:guide:start",
						"steps": [
							{
								"id": "talk",
								"description": "Talk to the guide.",
								"requirement": { "eventId": "conversation:node:guide:start", "count": 1 }
							}
						]
					}
				]
			}`),
		},
	}

	world, err := LoadFromGameFS(gameFS)
	if err != nil {
		t.Fatalf("LoadFromGameFS returned error: %v", err)
	}

	quest, ok := world.GetQuest("first_errand")
	if !ok {
		t.Fatal("quest first_errand was not loaded")
	}
	if quest.Steps[0].Requirement.EventId != "conversation:node:guide:start" {
		t.Fatalf("quest event id = %q, want conversation:node:guide:start", quest.Steps[0].Requirement.EventId)
	}
}

func TestLoadFromGameFSRejectsInvalidQuest(t *testing.T) {
	gameFS := fstest.MapFS{
		"game.json": {
			Data: []byte(`{
				"formatVersion": 1,
				"id": "test_game",
				"files": {
					"maps": ["maps/test.json"],
					"conversations": [],
					"quests": ["quests/tutorial.json"]
				}
			}`),
		},
		"maps/test.json": {
			Data: []byte(`{
				"formatVersion": 1,
				"id": "test",
				"size": { "x": 1, "y": 1 },
				"terrain": ["grass"]
			}`),
		},
		"quests/tutorial.json": {
			Data: []byte(`{
				"formatVersion": 1,
				"id": "tutorial_quests",
				"quests": [
					{
						"id": "broken",
						"steps": [
							{
								"id": "missing_event",
								"description": "Broken.",
								"requirement": { "eventId": "", "count": 1 }
							}
						]
					}
				]
			}`),
		},
	}

	if _, err := LoadFromGameFS(gameFS); err == nil {
		t.Fatal("LoadFromGameFS returned nil error for invalid quest")
	}
}
