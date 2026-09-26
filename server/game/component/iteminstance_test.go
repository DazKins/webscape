package component

import (
	"encoding/json"
	"reflect"
	"testing"
	"webscape/server/game/model"
	"webscape/server/util"
)

func TestCustomInstancesDoNotStackOrShareProperties(t *testing.T) {
	inventory := NewCInventory()
	ordinary := model.NewItem("arrow")
	ordinary.Quantity = 10
	custom := model.NewItem("arrow")
	custom.Properties = &model.ItemProperties{CustomName: "Lucky arrow"}
	if !inventory.AddItem(ordinary) || !inventory.AddItem(custom) || inventory.GetItemCount() != 2 {
		t.Fatal("custom item merged with stack")
	}
	if !inventory.AddItem(model.NewItem("arrow")) || ordinary.Quantity != 11 || custom.Quantity != 1 {
		t.Fatal("stacking changed custom instance")
	}
	invalid := custom.Clone()
	invalid.Id = model.NewItemId()
	invalid.Quantity = 2
	if inventory.AddItem(invalid) {
		t.Fatal("custom instances may not form stacks")
	}
	clone := inventory.Clone()
	clone.GetItem(custom.Id).Properties.CustomName = "Changed"
	if custom.Name() != "Lucky arrow" {
		t.Fatal("inventory transaction clone shares properties")
	}
	for _, raw := range []*model.Item{{Id: model.NewItemId(), DefinitionID: "missing", Quantity: 1}, {DefinitionID: "bread", Quantity: 1}} {
		if inventory.AddItem(raw) {
			t.Fatal("inventory accepted invalid identity")
		}
	}
}

func TestIndividualEquipmentPropertiesSurvivePlayerSaves(t *testing.T) {
	item := model.NewItem("ironSword")
	stats := item.CombatStats()
	stats.MaxDamage = 18
	item.Properties = &model.ItemProperties{CustomName: "Reliable Sword", CombatStats: stats}
	inventory := NewCInventory()
	inventory.AddItem(item)
	equipped := NewCEquipped()
	equipped.EquipItem(model.SlotWeapon, item)
	for _, owner := range []Component{inventory, equipped} {
		saved, err := SaveComponent(owner)
		wantVersion := 2
		if owner.GetId() == ComponentIdInventory {
			wantVersion = 3
		}
		if err != nil || saved.Version != wantVersion {
			t.Fatalf("save %s: %v", owner.GetId(), err)
		}
		restored, err := RestoreComponent(owner.GetId(), saved)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(owner, restored) {
			t.Fatalf("properties lost for %s", owner.GetId())
		}
	}
	wire := SerializeItem(item).(util.JObject)
	if wire["definitionId"] != util.JString("ironSword") || wire["name"] != util.JString("Reliable Sword") || wire["hasProperties"] != util.JBool(true) {
		t.Fatal("instance presentation not replicated")
	}
	if wire["combatStats"].(util.JObject)["maxDamage"] != util.JNumber(18) {
		t.Fatal("unique stats not replicated")
	}
	baseline := NewCEquipped()
	baseline.EquipItem(model.SlotWeapon, model.NewItem("ironSword"))
	if CalculateCombatStats(nil, equipped).GetMaxDamage() != CalculateCombatStats(nil, baseline).GetMaxDamage()+8 {
		t.Fatal("combat ignores instance stats")
	}
}

func TestLegacyInventoryMigratesToDefinitionReferences(t *testing.T) {
	sword := model.NewItem("ironSword")
	legacy := map[string]any{"Id": [16]byte(sword.Id), "Quantity": 1, "Name": "Iron Sword", "Type": "weapon", "RenderModel": "ironSword", "EquipmentSlot": "weapon", "CombatStats": sword.CombatStats()}
	for _, test := range []struct {
		id   ComponentId
		data any
	}{
		{ComponentIdInventory, map[string]any{"items": []any{legacy}}},
	} {
		data, _ := json.Marshal(test.data)
		restored, err := RestoreComponent(test.id, SavedComponent{Version: 1, Data: data})
		if err != nil {
			t.Fatalf("%s: %v", test.id, err)
		}
		saved, err := SaveComponent(restored)
		if err != nil || saved.Version != 3 {
			t.Fatalf("%s not upgraded: %v", test.id, err)
		}
		again, err := RestoreComponent(test.id, saved)
		if err != nil || !reflect.DeepEqual(restored, again) {
			t.Fatalf("%s migrated state changed: %v", test.id, err)
		}
	}
}

func TestShopOffersDoNotContainInstanceIdentity(t *testing.T) {
	shop, err := ParseShop(map[string]any{"offers": []any{map[string]any{"definitionId": "bread", "buyPrice": float64(4), "sellPrice": float64(1)}}})
	if err != nil {
		t.Fatal(err)
	}
	wire := shop.Serialize().(util.JObject)["offers"].(util.JArray)[0].(util.JObject)
	item := wire["item"].(util.JObject)
	if _, exists := item["id"]; exists {
		t.Fatal("shop has a fake inventory identity")
	}
	if wire["definitionId"] != util.JString("bread") || item["definitionId"] != util.JString("bread") {
		t.Fatal("missing catalogue identity")
	}
	if !util.JsonEqual(shop.Serialize(), shop.Serialize()) {
		t.Fatal("shop creates transient identities during sync")
	}
}
