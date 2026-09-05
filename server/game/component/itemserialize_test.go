package component

import (
	"testing"
	"webscape/server/game/model"
	"webscape/server/util"
)

func TestSerializeItemIncludesEquipmentRenderModel(t *testing.T) {
	serialized, ok := SerializeItem(model.CreateLeatherHelmet()).(util.JObject)
	if !ok {
		t.Fatal("serialized equipment item is not an object")
	}

	renderModel, ok := serialized["renderModel"]
	if !ok {
		t.Fatal("serialized equipment item does not include renderModel")
	}
	if !util.JsonEqual(renderModel, util.JString("leatherHelmet")) {
		t.Fatalf("renderModel = %v, want leatherHelmet", renderModel)
	}
}

func TestSerializeItemOmitsEmptyRenderModel(t *testing.T) {
	serialized, ok := SerializeItem(model.CreateBread()).(util.JObject)
	if !ok {
		t.Fatal("serialized item is not an object")
	}

	if _, ok := serialized["renderModel"]; ok {
		t.Fatal("serialized item includes an empty renderModel")
	}
}

func TestSerializeFishingRodIdentityWithoutCombatStats(t *testing.T) {
	serialized := SerializeItem(model.CreateFishingRod()).(util.JObject)
	if !util.JsonEqual(serialized["name"], util.JString("Fishing Rod")) ||
		!util.JsonEqual(serialized["type"], util.JString("fishingRod")) ||
		!util.JsonEqual(serialized["renderModel"], util.JString("fishingRod")) ||
		!util.JsonEqual(serialized["equipmentSlot"], util.JString("weapon")) {
		t.Fatalf("serialized fishing rod = %#v", serialized)
	}
	if _, ok := serialized["combatStats"]; ok {
		t.Fatal("serialized fishing rod includes combat bonuses")
	}
}

func TestSerializeMagicStaffCombatProfile(t *testing.T) {
	serialized := SerializeItem(model.CreateMagicStaff()).(util.JObject)
	stats := serialized["combatStats"].(util.JObject)
	if stats["attackMethod"] != util.JString("magic") ||
		stats["windUpTicks"] != util.JNumber(2) ||
		stats["travelTicks"] != util.JNumber(1) ||
		stats["projectileType"] != util.JString("magicBolt") ||
		stats["range"] != util.JNumber(4) {
		t.Fatalf("serialized magic staff combat stats = %#v", stats)
	}
}

func TestSerializeWoodenBowCombatProfile(t *testing.T) {
	serialized := SerializeItem(model.CreateWoodenBow()).(util.JObject)
	stats := serialized["combatStats"].(util.JObject)
	if stats["attackMethod"] != util.JString("ranged") ||
		stats["windUpTicks"] != util.JNumber(2) ||
		stats["travelTicks"] != util.JNumber(1) ||
		stats["projectileType"] != util.JString("arrow") ||
		stats["range"] != util.JNumber(3) {
		t.Fatalf("serialized wooden bow combat stats = %#v", stats)
	}
}

func TestSerializeExistingWeaponDefaultsToMeleeProfile(t *testing.T) {
	serialized := SerializeItem(model.CreateIronSword()).(util.JObject)
	stats := serialized["combatStats"].(util.JObject)
	if stats["attackMethod"] != util.JString("melee") ||
		stats["windUpTicks"] != util.JNumber(0) ||
		stats["travelTicks"] != util.JNumber(0) {
		t.Fatalf("serialized sword combat profile = %#v", stats)
	}
}
