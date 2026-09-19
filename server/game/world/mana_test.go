package world

import (
	"strings"
	"testing"
	"webscape/server/game/model"
)

func TestManaSettingsDefaultsAndValidation(t *testing.T) {
	for _, tc := range []struct {
		name, json string
		valid      bool
	}{
		{"omitted", "", true},
		{"custom", `{"maxMana":40,"staffCastCost":5,"regenAmount":2,"regenIntervalTicks":7}`, true},
		{"null", `null`, false},
		{"missing", `{"maxMana":40}`, false},
		{"negative", `{"maxMana":40,"staffCastCost":-1,"regenAmount":2,"regenIntervalTicks":7}`, false},
		{"zero interval", `{"maxMana":40,"staffCastCost":5,"regenAmount":2,"regenIntervalTicks":0}`, false},
		{"unaffordable", `{"maxMana":40,"staffCastCost":41,"regenAmount":2,"regenIntervalTicks":7}`, false},
		{"excessive regen", `{"maxMana":40,"staffCastCost":5,"regenAmount":41,"regenIntervalTicks":7}`, false},
		{"fraction", `{"maxMana":40,"staffCastCost":1.5,"regenAmount":2,"regenIntervalTicks":7}`, false},
		{"string", `{"maxMana":"40","staffCastCost":5,"regenAmount":2,"regenIntervalTicks":7}`, false},
		{"unknown", `{"maxMana":40,"staffCastCost":5,"regenAmount":2,"regenIntervalTicks":7,"extra":1}`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fs := chunkFS([]string{"chunks/a.json"}, map[string]string{"chunks/a.json": testChunk("a", 0, 0, `[{"id":"spawn","components":{"position":{"x":0,"y":0},"playerSpawn":{}}}]`)})
			if tc.json != "" {
				fs["game.json"].Data = []byte(strings.Replace(string(fs["game.json"].Data), `{`, `{"mana":`+tc.json+`,`, 1))
			}
			w, err := LoadFromGameFS(fs)
			if (err == nil) != tc.valid {
				t.Fatalf("valid=%v error=%v", tc.valid, err)
			}
			if err == nil {
				want := model.DefaultManaSettings()
				if tc.name == "custom" {
					want = model.ManaSettings{MaxMana: 40, StaffCastCost: 5, RegenAmount: 2, RegenIntervalTicks: 7}
				}
				if w.GetManaSettings() != want {
					t.Fatalf("settings = %+v, want %+v", w.GetManaSettings(), want)
				}
			}
		})
	}
}
