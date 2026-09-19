package world

import (
	"fmt"
	"testing"
)

func TestAuthoredBankerValidation(t *testing.T) {
	for _, test := range []struct {
		name, shop string
		valid      bool
	}{
		{"valid", `{}`, true}, {"null", `null`, false}, {"array", `[]`, false}, {"unknown field", `{"items":[]}`, false},
	} {
		for _, template := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/template=%v", test.name, template), func(t *testing.T) {
				props := `"banker":` + test.shop
				if template {
					props = `"spawn":{"respawnTicks":10,"entity":{"components":{` + props + `}}}`
				}
				entities := `[{"id":"spawn","components":{"position":{"x":0,"y":0},"playerSpawn":{}}},{"id":"merchant","components":{"position":{"x":1,"y":0},` + props + `}}]`
				_, err := LoadFromGameFS(chunkFS([]string{"chunks/a.json"}, map[string]string{"chunks/a.json": testChunk("a", 0, 0, entities)}))
				if (err == nil) != test.valid {
					t.Fatalf("valid=%v, error=%v", test.valid, err)
				}
			})
		}
	}
}
