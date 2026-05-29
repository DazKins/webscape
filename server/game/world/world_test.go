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
				"spawns": [
					{ "type": "player", "x": 1, "y": 0 }
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
