package model

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestItemDefinitionsHashTracksValues(t *testing.T) {
	original := itemDefinitions
	t.Cleanup(func() { itemDefinitions = original })
	hash := ItemDefinitionsHash()
	// Rebuild in reverse order; neither insertion order nor copied stats change it.
	ids := ItemDefinitionIDs()
	reordered := make(map[string]ItemDefinition, len(ids))
	for i := len(ids) - 1; i >= 0; i-- {
		reordered[ids[i]], _ = GetItemDefinition(ids[i])
	}
	itemDefinitions = reordered
	if ItemDefinitionsHash() != hash {
		t.Fatal("equivalent definitions changed the content fingerprint")
	}
	itemDefinitions["ironSword"].CombatStats.MaxDamage++
	if ItemDefinitionsHash() == hash {
		t.Fatal("changed combat stats did not change the content fingerprint")
	}
}

func TestCatalogueCreatesIndependentInstances(t *testing.T) {
	for _, id := range ItemDefinitionIDs() {
		a, b := NewItem(id), NewItem(id)
		if a == nil || b == nil || a.Id == b.Id || a.DefinitionID != id || a.Quantity != 1 || a.HasProperties() {
			t.Fatalf("bad instances for %s", id)
		}
		if err := a.ValidateSaved(); err != nil {
			t.Fatalf("%s: %v", id, err)
		}
	}
	if NewItem("missing") != nil {
		t.Fatal("unknown definition created an item")
	}
}

func TestDefinitionAndInstancePropertiesAreIsolated(t *testing.T) {
	a, b := NewItem("ironSword"), NewItem("ironSword")
	definition := a.Definition()
	definition.CombatStats.MaxDamage = 99
	if b.CombatStats().MaxDamage != 10 {
		t.Fatal("definition lookup leaked mutable stats")
	}
	a.Properties = &ItemProperties{CustomName: "Old Faithful", CombatStats: a.CombatStats()}
	a.Properties.CombatStats.MaxDamage = 20
	if a.Name() != "Old Faithful" || a.CombatStats().MaxDamage != 20 || b.Name() != "Iron Sword" || b.CombatStats().MaxDamage != 10 {
		t.Fatal("instance properties leaked")
	}
	cloned := a.Clone()
	cloned.Properties.CustomName = "Clone"
	cloned.Properties.CombatStats.MaxDamage = 30
	if a.Name() != "Old Faithful" || a.CombatStats().MaxDamage != 20 {
		t.Fatal("clone shares properties")
	}
	data, err := json.Marshal(a)
	if err != nil {
		t.Fatal(err)
	}
	var restored Item
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(a, &restored) {
		t.Fatalf("properties lost: %s", data)
	}
	for _, forbidden := range []string{`"type"`, `"renderModel"`, `"equipmentSlot"`} {
		if strings.Contains(string(data), forbidden) {
			t.Fatalf("saved shared definition data: %s", data)
		}
	}
}

func TestLegacyItemsMigrateWithoutChangingIdentity(t *testing.T) {
	for _, id := range ItemDefinitionIDs() {
		original := NewItem(id)
		definition := original.Definition()
		renderModel := ""
		if original.IsEquipable() {
			renderModel = definition.RenderModel
		}
		legacy := map[string]any{
			"Id": [16]byte(original.Id), "Name": definition.Name, "Type": definition.Type,
			"Quantity": 1, "RenderModel": renderModel, "EquipmentSlot": original.GetEquipmentSlot(), "CombatStats": definition.CombatStats,
		}
		data, _ := json.Marshal(legacy)
		var item Item
		if err := json.Unmarshal(data, &item); err != nil {
			t.Fatalf("%s: %v", id, err)
		}
		if !reflect.DeepEqual(original, &item) {
			t.Fatalf("%s migrated incorrectly: %#v", id, item)
		}
	}
	sword := NewItem("ironSword")
	stats := sword.CombatStats()
	stats.MaxDamage = 17
	data, _ := json.Marshal(map[string]any{"Id": sword.Id, "Quantity": 1, "Name": "Iron Sword", "Type": "weapon", "RenderModel": "ironSword", "EquipmentSlot": "weapon", "CombatStats": stats})
	var restored Item
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatal(err)
	}
	if restored.Properties == nil || restored.CombatStats().MaxDamage != 17 {
		t.Fatal("legacy unique stats lost")
	}
}

func TestRejectInvalidItemsAndReferences(t *testing.T) {
	base := NewItem("ironSword")
	data, _ := json.Marshal(base)
	var fields map[string]any
	json.Unmarshal(data, &fields)
	for _, change := range []map[string]any{
		{"definitionId": "missing"}, {"quantity": 2}, {"id": "bad-uuid"},
		{"properties": map[string]any{"unknown": true}},
		{"properties": map[string]any{"customName": " "}},
		{"properties": map[string]any{"combatStats": map[string]any{"minDamage": 20, "maxDamage": 2}}},
	} {
		invalid := map[string]any{}
		for key, value := range fields {
			invalid[key] = value
		}
		for key, value := range change {
			invalid[key] = value
		}
		raw, _ := json.Marshal(invalid)
		var item Item
		if json.Unmarshal(raw, &item) == nil {
			t.Fatalf("accepted %s", raw)
		}
	}
	for _, raw := range []string{
		`{"definitionId":"missing","count":1}`, `{"definitionId":"bread","count":0}`,
		`{"definitionId":"bread","count":1.5}`, `{"definitionId":"bread","count":2147483648}`,
		`{"definitionId":"bread","count":1,"name":""}`, `{"name":"Unknown","type":"quest","count":1}`,
	} {
		var ref ItemReference
		if json.Unmarshal([]byte(raw), &ref) == nil {
			t.Fatalf("accepted %s", raw)
		}
	}
}
