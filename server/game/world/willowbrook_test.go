package world

import (
	"testing"
	"webscape/server/math"
)

// Authored scenery must not strand a tutor, shop, fishing bank, or quest chest.
// Doors are considered open for this flood fill because players can open them.
func TestWillowbrookLandmarksAreReachable(t *testing.T) {
	w, err := LoadFromGameFolder("../../../game-project")
	if err != nil {
		t.Fatal(err)
	}
	passable := map[math.Vec2]bool{}
	for coord, chunk := range w.chunks {
		origin := w.ChunkOrigin(coord)
		for i, blocked := range chunk.Blockers {
			passable[origin.Add(math.Vec2{X: i % w.chunkSize.X, Y: i / w.chunkSize.X})] = !blocked
		}
		for _, wall := range chunk.Walls {
			passable[origin.Add(math.Vec2{X: wall.X, Y: wall.Y})] = false
		}
	}
	occupied := map[math.Vec2]string{}
	for _, entity := range w.entities {
		pos, _ := entityPosition(entity)
		if entity.Components["spawn"] != nil {
			continue
		}
		width, height := entitySize(entity)
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				tile := pos.Add(math.Vec2{X: x, Y: y})
				if previous, exists := occupied[tile]; exists {
					t.Errorf("%s overlaps %s", entity.Id, previous)
				}
				occupied[tile] = entity.Id
				if metadata, ok := entity.Components["metadata"].(map[string]any); ok && metadata["blocksMovement"] == true && entity.Components["openable"] == nil {
					passable[tile] = false
				}
			}
		}
	}
	if !passable[w.playerSpawn] {
		t.Fatal("player spawn is blocked")
	}
	reached := map[math.Vec2]bool{w.playerSpawn: true}
	queue := []math.Vec2{w.playerSpawn}
	directions := []math.Vec2{{X: 1}, {X: -1}, {Y: 1}, {Y: -1}}
	for len(queue) > 0 {
		pos := queue[0]
		queue = queue[1:]
		for _, dir := range directions {
			next := pos.Add(dir)
			if passable[next] && !reached[next] {
				reached[next] = true
				queue = append(queue, next)
			}
		}
	}
	for _, entity := range w.entities {
		c := entity.Components
		if c["conversation"] == nil && c["shop"] == nil && c["lootable"] == nil && c["fishable"] == nil && c["openable"] == nil && c["spawn"] == nil {
			continue
		}
		pos, _ := entityPosition(entity)
		reachable := false
		for _, dir := range directions {
			if reached[pos.Add(dir)] {
				reachable = true
			}
		}
		if !reachable {
			t.Errorf("landmark %s has no reachable adjacent tile", entity.Id)
		}
	}
}
