package world

import (
	"fmt"
	"testing"
)

func TestAuthoredShopValidation(t *testing.T) {
	for _, test := range []struct {
		name, shop string
		valid      bool
	}{
		{"valid", `{"offers":[{"definitionId":"ironSword","buyPrice":40,"sellPrice":16}]}`, true},
		{"unknown", `{"offers":[{"definitionId":"fake","buyPrice":40,"sellPrice":16}]}`, false},
		{"gold cannot be traded", `{"offers":[{"definitionId":"gold","buyPrice":40,"sellPrice":16}]}`, false},
		{"fraction", `{"offers":[{"definitionId":"logs","buyPrice":4.5,"sellPrice":2}]}`, false},
		{"negative", `{"offers":[{"definitionId":"logs","buyPrice":4,"sellPrice":-1}]}`, false},
		{"arbitrage", `{"offers":[{"definitionId":"logs","buyPrice":4,"sellPrice":4}]}`, false},
		{"overflow", `{"offers":[{"definitionId":"logs","buyPrice":1000001,"sellPrice":2}]}`, false},
		{"duplicate", `{"offers":[{"definitionId":"logs","buyPrice":4,"sellPrice":2},{"definitionId":"logs","buyPrice":5,"sellPrice":2}]}`, false},
		{"empty", `{"offers":[]}`, false},
		{"null", `null`, false},
		{"unknown field", `{"offers":[{"definitionId":"logs","buyPrice":4,"sellPrice":2,"stock":1}]}`, false},
	} {
		for _, template := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/template=%v", test.name, template), func(t *testing.T) {
				props := `"shop":` + test.shop
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
