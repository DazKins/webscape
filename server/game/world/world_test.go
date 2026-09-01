package world

import (
	"fmt"
	"strings"
	"testing"
	"testing/fstest"
)

func TestLoadChunksConvertsNegativeLocalPositionsToGlobal(t *testing.T) {
	gameFS := chunkFS([]string{"chunks/a.json", "chunks/b.json"}, map[string]string{
		"chunks/a.json": testChunk("a", 0, 0, `[{"id":"player_spawn","components":{"position":{"x":1,"y":1},"playerSpawn":{}}}]`),
		"chunks/b.json": testChunk("b", -2, 1, `[{"id":"tree","components":{"position":{"x":3,"y":2},"metadata":{"width":1,"height":1}}}]`),
	})
	loaded, err := LoadFromGameFS(gameFS)
	if err != nil {
		t.Fatalf("LoadFromGameFS: %v", err)
	}
	entities := loaded.GetEntities()
	position, ok := entityPosition(entities[1])
	if !ok || position.X != -5 || position.Y != 6 {
		t.Fatalf("global position = %#v, want (-5,6)", position)
	}
	coord, local := loaded.GlobalToChunk(-1, -1)
	if coord != (ChunkCoord{X: -1, Y: -1}) || local.X != 3 || local.Y != 3 {
		t.Fatalf("negative conversion = %#v %#v", coord, local)
	}
}

func TestMissingChunkIsBlockedAndAdjacentChunksConnect(t *testing.T) {
	loaded, err := LoadFromGameFS(chunkFS([]string{"chunks/a.json", "chunks/b.json"}, map[string]string{
		"chunks/a.json": testChunk("a", 0, 0, `[{"id":"player_spawn","components":{"position":{"x":0,"y":0},"playerSpawn":{}}}]`),
		"chunks/b.json": testChunk("b", 1, 0, `[]`),
	}))
	if err != nil {
		t.Fatal(err)
	}
	if loaded.GetStaticWall(4, 1) {
		t.Fatal("existing adjacent chunk seam was blocked")
	}
	if !loaded.GetStaticWall(-1, 1) || !loaded.GetStaticWall(1, 4) {
		t.Fatal("missing chunk must be blocked")
	}
}

func TestLoadChunksRejectsMalformedArrays(t *testing.T) {
	bad := strings.Replace(testChunk("a", 0, 0, `[{"id":"player_spawn","components":{"position":{"x":0,"y":0},"playerSpawn":{}}}]`),
		`"terrain":["grass","grass","grass","grass","grass","grass","grass","grass","grass","grass","grass","grass","grass","grass","grass","grass"]`, `"terrain":["grass"]`, 1)
	_, err := LoadFromGameFS(chunkFS([]string{"chunks/a.json"}, map[string]string{"chunks/a.json": bad}))
	if err == nil || !strings.Contains(err.Error(), "terrain length must be 16") {
		t.Fatalf("error = %v", err)
	}
}

func TestLoadChunksRejectsDuplicateCoordinateAndEntityId(t *testing.T) {
	tests := []struct {
		name string
		b    string
		want string
	}{
		{"coordinate", testChunk("b", 0, 0, `[]`), "duplicate chunk coordinate"},
		{"chunk id", testChunk("a", 1, 0, `[]`), "duplicate chunk id"},
		{"entity id", testChunk("b", 1, 0, `[{"id":"player_spawn","components":{"position":{"x":0,"y":0}}}]`), "duplicate entity id"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := LoadFromGameFS(chunkFS([]string{"chunks/a.json", "chunks/b.json"}, map[string]string{
				"chunks/a.json": testChunk("a", 0, 0, `[{"id":"player_spawn","components":{"position":{"x":0,"y":0},"playerSpawn":{}}}]`),
				"chunks/b.json": test.b,
			}))
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestLoadChunksRequiresOneSpawnAndContainedFootprints(t *testing.T) {
	tests := []struct{ entities, want string }{
		{`[]`, "exactly one playerSpawn"},
		{`[{"id":"one","components":{"position":{"x":0,"y":0},"playerSpawn":{}}},{"id":"two","components":{"position":{"x":1,"y":0},"playerSpawn":{}}}]`, "exactly one playerSpawn"},
		{`[{"id":"player_spawn","components":{"position":{"x":0,"y":0},"playerSpawn":{}}},{"id":"wide","components":{"position":{"x":3,"y":0},"metadata":{"width":2,"height":1}}}]`, "footprint is out of chunk bounds"},
	}
	for _, test := range tests {
		_, err := LoadFromGameFS(chunkFS([]string{"chunks/a.json"}, map[string]string{"chunks/a.json": testChunk("a", 0, 0, test.entities)}))
		if err == nil || !strings.Contains(err.Error(), test.want) {
			t.Fatalf("error = %v, want %q", err, test.want)
		}
	}
}

func TestLoadChunksRejectsLegacyProjectAndInvalidPath(t *testing.T) {
	legacy := fstest.MapFS{"game.json": {Data: []byte(`{"formatVersion":1,"id":"old","files":{"maps":["maps/a.json"]}}`)}}
	if _, err := LoadFromGameFS(legacy); err == nil || !strings.Contains(err.Error(), "unsupported game format version 1") {
		t.Fatalf("legacy error = %v", err)
	}
	invalid := chunkFS([]string{"../outside.json"}, nil)
	if _, err := LoadFromGameFS(invalid); err == nil || !strings.Contains(err.Error(), "invalid chunk path") {
		t.Fatalf("path error = %v", err)
	}
}

func TestLoadChunksValidatesWoodcuttableComponents(t *testing.T) {
	valid := `{"maxDurability":5,"respawnTicks":60,"yield":{"name":"Logs","type":"material","count":1}}`
	tests := []struct {
		name         string
		woodcuttable string
		want         string
	}{
		{name: "not object", woodcuttable: `[]`, want: "woodcuttable must be an object"},
		{name: "durability zero", woodcuttable: `{"maxDurability":0,"respawnTicks":60,"yield":{"name":"Logs","type":"material","count":1}}`, want: "maxDurability must be a positive integer"},
		{name: "durability fractional", woodcuttable: `{"maxDurability":1.5,"respawnTicks":60,"yield":{"name":"Logs","type":"material","count":1}}`, want: "maxDurability must be a positive integer"},
		{name: "respawn zero", woodcuttable: `{"maxDurability":5,"respawnTicks":0,"yield":{"name":"Logs","type":"material","count":1}}`, want: "respawnTicks must be a positive integer"},
		{name: "yield missing", woodcuttable: `{"maxDurability":5,"respawnTicks":60}`, want: "yield must be an object"},
		{name: "yield name empty", woodcuttable: `{"maxDurability":5,"respawnTicks":60,"yield":{"name":" ","type":"material","count":1}}`, want: "yield.name must be a non-empty string"},
		{name: "yield not material", woodcuttable: `{"maxDurability":5,"respawnTicks":60,"yield":{"name":"Logs","type":"weapon","count":1}}`, want: `yield.type must be "material"`},
		{name: "count zero", woodcuttable: `{"maxDurability":5,"respawnTicks":60,"yield":{"name":"Logs","type":"material","count":0}}`, want: "yield.count must be a positive integer"},
		{name: "count fractional", woodcuttable: `{"maxDurability":5,"respawnTicks":60,"yield":{"name":"Logs","type":"material","count":1.5}}`, want: "yield.count must be a positive integer"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			entities := fmt.Sprintf(`[{"id":"player_spawn","components":{"position":{"x":0,"y":0},"playerSpawn":{}}},{"id":"tree","components":{"position":{"x":1,"y":0},"woodcuttable":%s}}]`, test.woodcuttable)
			_, err := LoadFromGameFS(chunkFS([]string{"chunks/a.json"}, map[string]string{"chunks/a.json": testChunk("a", 0, 0, entities)}))
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %v, want %q", err, test.want)
			}
		})
	}

	entities := fmt.Sprintf(`[{"id":"player_spawn","components":{"position":{"x":0,"y":0},"playerSpawn":{}}},{"id":"tree","components":{"position":{"x":1,"y":0},"woodcuttable":%s}}]`, valid)
	if _, err := LoadFromGameFS(chunkFS([]string{"chunks/a.json"}, map[string]string{"chunks/a.json": testChunk("a", 0, 0, entities)})); err != nil {
		t.Fatalf("valid woodcuttable rejected: %v", err)
	}
}

func TestLoadChunksValidatesFishableComponents(t *testing.T) {
	valid := `{"catchChancePercent":5,"yield":{"name":"Raw Fish","type":"fish","count":1}}`
	tests := []struct {
		name     string
		fishable string
		want     string
	}{
		{name: "not object", fishable: `[]`, want: "fishable must be an object"},
		{name: "chance zero", fishable: `{"catchChancePercent":0,"yield":{"name":"Raw Fish","type":"fish","count":1}}`, want: "catchChancePercent must be an integer from 1 to 100"},
		{name: "chance over 100", fishable: `{"catchChancePercent":101,"yield":{"name":"Raw Fish","type":"fish","count":1}}`, want: "catchChancePercent must be an integer from 1 to 100"},
		{name: "chance fractional", fishable: `{"catchChancePercent":5.5,"yield":{"name":"Raw Fish","type":"fish","count":1}}`, want: "catchChancePercent must be an integer from 1 to 100"},
		{name: "yield missing", fishable: `{"catchChancePercent":5}`, want: "yield must be an object"},
		{name: "yield name empty", fishable: `{"catchChancePercent":5,"yield":{"name":" ","type":"fish","count":1}}`, want: "yield.name must be a non-empty string"},
		{name: "yield type empty", fishable: `{"catchChancePercent":5,"yield":{"name":"Raw Fish","type":" ","count":1}}`, want: "yield.type must be a non-empty string"},
		{name: "count zero", fishable: `{"catchChancePercent":5,"yield":{"name":"Raw Fish","type":"fish","count":0}}`, want: "yield.count must be a positive integer"},
		{name: "count fractional", fishable: `{"catchChancePercent":5,"yield":{"name":"Raw Fish","type":"fish","count":1.5}}`, want: "yield.count must be a positive integer"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			entities := fmt.Sprintf(`[{"id":"player_spawn","components":{"position":{"x":0,"y":0},"playerSpawn":{}}},{"id":"spot","components":{"position":{"x":1,"y":0},"fishable":%s}}]`, test.fishable)
			_, err := LoadFromGameFS(chunkFS([]string{"chunks/a.json"}, map[string]string{"chunks/a.json": testChunk("a", 0, 0, entities)}))
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %v, want %q", err, test.want)
			}
		})
	}

	entities := fmt.Sprintf(`[{"id":"player_spawn","components":{"position":{"x":0,"y":0},"playerSpawn":{}}},{"id":"spot","components":{"position":{"x":1,"y":0},"fishable":%s}}]`, valid)
	if _, err := LoadFromGameFS(chunkFS([]string{"chunks/a.json"}, map[string]string{"chunks/a.json": testChunk("a", 0, 0, entities)})); err != nil {
		t.Fatalf("valid fishable rejected: %v", err)
	}
}

func TestLoadChunksValidatesDirectAndSpawnedAppearances(t *testing.T) {
	valid := `{"skinTone":"tan","hairStyle":"swept","hairColor":"chestnut","tunicColor":"forest","trousersColor":"charcoal","shoeColor":"darkBrown"}`
	invalid := []struct {
		name       string
		appearance string
		want       string
	}{
		{name: "not object", appearance: `[]`, want: "appearance must be an object"},
		{name: "partial", appearance: `{"skinTone":"tan"}`, want: "appearance.hairStyle must be a non-empty string"},
		{name: "unknown value", appearance: `{"skinTone":"green","hairStyle":"swept","hairColor":"chestnut","tunicColor":"forest","trousersColor":"charcoal","shoeColor":"darkBrown"}`, want: "appearance.skinTone must be one of"},
		{name: "extra field", appearance: `{"skinTone":"tan","hairStyle":"swept","hairColor":"chestnut","tunicColor":"forest","trousersColor":"charcoal","shoeColor":"darkBrown","hat":"cap"}`, want: "appearance contains unknown field"},
	}

	for _, location := range []string{"direct", "spawned"} {
		for _, test := range invalid {
			t.Run(location+" "+test.name, func(t *testing.T) {
				var subject string
				if location == "direct" {
					subject = fmt.Sprintf(`{"id":"npc","components":{"position":{"x":1,"y":0},"renderable":{"type":"human"},"appearance":%s}}`, test.appearance)
				} else {
					subject = fmt.Sprintf(`{"id":"npc_spawn","components":{"position":{"x":1,"y":0},"spawn":{"respawnTicks":1,"entity":{"components":{"renderable":{"type":"human"},"appearance":%s}}}}}`, test.appearance)
				}
				entities := `[{"id":"player_spawn","components":{"position":{"x":0,"y":0},"playerSpawn":{}}},` + subject + `]`
				_, err := LoadFromGameFS(chunkFS([]string{"chunks/a.json"}, map[string]string{"chunks/a.json": testChunk("a", 0, 0, entities)}))
				if err == nil || !strings.Contains(err.Error(), test.want) {
					t.Fatalf("error = %v, want %q", err, test.want)
				}
			})
		}
	}

	entities := fmt.Sprintf(`[{"id":"player_spawn","components":{"position":{"x":0,"y":0},"playerSpawn":{}}},{"id":"npc","components":{"position":{"x":1,"y":0},"renderable":{"type":"human"},"appearance":%s}},{"id":"npc_spawn","components":{"position":{"x":2,"y":0},"spawn":{"respawnTicks":1,"entity":{"components":{"renderable":{"type":"human"},"appearance":%s}}}}}]`, valid, valid)
	if _, err := LoadFromGameFS(chunkFS([]string{"chunks/a.json"}, map[string]string{"chunks/a.json": testChunk("a", 0, 0, entities)})); err != nil {
		t.Fatalf("valid direct and spawned appearances rejected: %v", err)
	}
}

func chunkFS(paths []string, documents map[string]string) fstest.MapFS {
	quoted := make([]string, len(paths))
	for i, path := range paths {
		quoted[i] = fmt.Sprintf("%q", path)
	}
	result := fstest.MapFS{"game.json": {Data: []byte(fmt.Sprintf(`{"formatVersion":2,"id":"test","world":{"chunkSize":{"x":4,"y":4}},"files":{"chunks":[%s],"conversations":[],"quests":[]}}`, strings.Join(quoted, ",")))}}
	for path, document := range documents {
		result[path] = &fstest.MapFile{Data: []byte(document)}
	}
	return result
}

func testChunk(id string, x, y int, entities string) string {
	values := `"grass","grass","grass","grass","grass","grass","grass","grass","grass","grass","grass","grass","grass","grass","grass","grass"`
	return fmt.Sprintf(`{"formatVersion":2,"id":%q,"coordinate":{"x":%d,"y":%d},"terrain":[%s],"heights":[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0],"blockers":[false,false,false,false,false,false,false,false,false,false,false,false,false,false,false,false],"walls":[],"entities":%s}`, id, x, y, values, entities)
}
