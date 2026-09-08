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
		{"valid", `{"offers":[{"itemId":"ironSword","buyPrice":40,"sellPrice":16}]}`, true},
		{"unknown", `{"offers":[{"itemId":"fake","buyPrice":40,"sellPrice":16}]}`, false},
		{"gold cannot be traded", `{"offers":[{"itemId":"gold","buyPrice":40,"sellPrice":16}]}`, false},
		{"fraction", `{"offers":[{"itemId":"logs","buyPrice":4.5,"sellPrice":2}]}`, false},
		{"negative", `{"offers":[{"itemId":"logs","buyPrice":4,"sellPrice":-1}]}`, false},
		{"arbitrage", `{"offers":[{"itemId":"logs","buyPrice":4,"sellPrice":4}]}`, false},
		{"overflow", `{"offers":[{"itemId":"logs","buyPrice":1000001,"sellPrice":2}]}`, false},
		{"duplicate", `{"offers":[{"itemId":"logs","buyPrice":4,"sellPrice":2},{"itemId":"logs","buyPrice":5,"sellPrice":2}]}`, false},
		{"empty", `{"offers":[]}`, false},
		{"null", `null`, false},
		{"unknown field", `{"offers":[{"itemId":"logs","buyPrice":4,"sellPrice":2,"stock":1}]}`, false},
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
