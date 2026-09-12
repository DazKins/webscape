package component

import (
	"testing"
	"webscape/server/game/model"
)

func TestAuthoredEquipmentUsesFreshCatalogItems(t *testing.T) {
	raw := map[string]any{"slots": map[string]any{"weapon": "magicStaff", "offhand": "woodenShield"}}
	first, err := ParseAuthoredEquipment(raw)
	if err != nil {
		t.Fatal(err)
	}
	second, err := ParseAuthoredEquipment(raw)
	if err != nil {
		t.Fatal(err)
	}
	staff := first.GetEquippedItem(model.SlotWeapon)
	if !model.SameItemKind(staff, model.CreateMagicStaff()) {
		t.Fatal("staff lost catalog stats")
	}
	if staff.Id == second.GetEquippedItem(model.SlotWeapon).Id {
		t.Fatal("entities share item identity")
	}
	if first.GetEquippedItem(model.SlotOffhand) == nil {
		t.Fatal("shield missing")
	}
}

func TestAuthoredEquipmentRejectsInvalidSlotsAndItems(t *testing.T) {
	for _, raw := range []any{
		nil, "staff", map[string]any{"weapon": "magicStaff"},
		map[string]any{"slots": nil},
		map[string]any{"slots": map[string]any{"head": "magicStaff"}},
		map[string]any{"slots": map[string]any{"weapon": "bread"}},
		map[string]any{"slots": map[string]any{"weapon": "unknown"}},
		map[string]any{"slots": map[string]any{"back": "woodenShield"}},
		map[string]any{"slots": map[string]any{"weapon": 4}},
	} {
		if _, err := ParseAuthoredEquipment(raw); err == nil {
			t.Fatalf("accepted invalid equipment %#v", raw)
		}
	}
	for _, raw := range []any{map[string]any{}, map[string]any{"slots": map[string]any{}}} {
		if _, err := ParseAuthoredEquipment(raw); err != nil {
			t.Fatalf("legacy empty equipment: %v", err)
		}
	}
}
