package component

import (
	"encoding/json"
	"testing"
)

func TestManaSaveRoundTripAndInvalidSaves(t *testing.T) {
	mana := NewCMana(120, 37)
	saved, err := mana.Save()
	if err != nil {
		t.Fatal(err)
	}
	restored, err := RestoreComponent(ComponentIdMana, saved)
	if err != nil {
		t.Fatal(err)
	}
	got := restored.(*CMana)
	if got.GetCurrentMana() != 37 || got.GetMaxMana() != 120 {
		t.Fatal("mana lost during save")
	}
	for _, data := range []string{`{"maxMana":100,"currentMana":-1}`, `{"maxMana":0,"currentMana":0}`, `{"maxMana":100,"currentMana":101}`, `{"maxMana":100}`, `null`} {
		if _, err := RestoreComponent(ComponentIdMana, SavedComponent{Version: 1, Data: json.RawMessage(data)}); err == nil {
			t.Fatalf("accepted invalid mana %s", data)
		}
	}
	if got.Spend(38) || got.Spend(-1) || got.GetCurrentMana() != 37 {
		t.Fatal("invalid spending changed mana")
	}
}
