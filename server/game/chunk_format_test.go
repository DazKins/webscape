package game

import (
	"encoding/json"
	"testing"
	"testing/fstest"
)

// chunkTestFS keeps behavioral tests focused on game systems while their small
// inline fixtures are expressed in the retired map shape.
func chunkTestFS(t *testing.T, source fstest.MapFS) fstest.MapFS {
	t.Helper()
	var manifest map[string]any
	if err := json.Unmarshal(source["game.json"].Data, &manifest); err != nil {
		t.Fatal(err)
	}
	files := manifest["files"].(map[string]any)
	maps, _ := files["maps"].([]any)
	chunks := make([]any, 0, len(maps))
	chunkSize := map[string]any{"x": float64(1), "y": float64(1)}
	for _, rawPath := range maps {
		path := rawPath.(string)
		var document map[string]any
		if err := json.Unmarshal(source[path].Data, &document); err != nil {
			t.Fatal(err)
		}
		size := document["size"].(map[string]any)
		chunkSize = size
		document["formatVersion"] = float64(2)
		document["coordinate"] = map[string]any{"x": float64(0), "y": float64(0)}
		delete(document, "size")
		count := int(size["x"].(float64) * size["y"].(float64))
		if blockers, ok := document["blockers"].([]any); !ok || len(blockers) == 0 {
			document["blockers"] = make([]bool, count)
		}
		entities, _ := document["entities"].([]any)
		hasSpawn := false
		for _, raw := range entities {
			entity := raw.(map[string]any)
			components, _ := entity["components"].(map[string]any)
			_, hasSpawn = components["playerSpawn"]
			if hasSpawn {
				break
			}
		}
		if !hasSpawn {
			document["entities"] = append(entities, map[string]any{"id": "player_spawn", "components": map[string]any{"position": map[string]any{"x": float64(0), "y": float64(0)}, "playerSpawn": map[string]any{}}})
		}
		data, err := json.Marshal(document)
		if err != nil {
			t.Fatal(err)
		}
		source[path] = &fstest.MapFile{Data: data}
		chunks = append(chunks, path)
	}
	manifest["formatVersion"] = float64(2)
	manifest["world"] = map[string]any{"chunkSize": chunkSize}
	delete(files, "maps")
	files["chunks"] = chunks
	data, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	source["game.json"] = &fstest.MapFile{Data: data}
	return source
}
