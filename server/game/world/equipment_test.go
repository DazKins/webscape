package world

import (
	"fmt"
	"testing"
)

func TestAuthoredEquipmentValidationIncludesSpawnTemplates(t *testing.T) {
	for _, equipment := range []struct {
		raw   string
		valid bool
	}{
		{`{}`, true},
		{`{"slots":{"weapon":"woodcuttingAxe","offhand":"woodenShield"}}`, true},
		{`{"slots":{"weapon":"bread"}}`, false},
		{`{"slots":{"head":"ironSword"}}`, false},
	} {
		for _, template := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/template=%v", equipment.raw, template), func(t *testing.T) {
				props := `"equipped":` + equipment.raw
				if template {
					props = `"spawn":{"respawnTicks":10,"entity":{"components":{` + props + `}}}`
				}
				entities := `[{"id":"spawn","components":{"position":{"x":0,"y":0},"playerSpawn":{}}},{"id":"tutor","components":{"position":{"x":1,"y":0},` + props + `}}]`
				_, err := LoadFromGameFS(chunkFS([]string{"chunks/a.json"}, map[string]string{"chunks/a.json": testChunk("a", 0, 0, entities)}))
				if (err == nil) != equipment.valid {
					t.Fatalf("valid=%v, error=%v", equipment.valid, err)
				}
			})
		}
	}
}
