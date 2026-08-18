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
