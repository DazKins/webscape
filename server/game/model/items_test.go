package model

import "testing"

func TestWoodcuttingAxeIsAWeaponSlotAxeBelowIronSwordStats(t *testing.T) {
	axe := CreateWoodcuttingAxe()
	sword := CreateIronSword()
	if axe.Name != "Woodcutting Axe" || axe.Type != "axe" || axe.RenderModel != "woodcuttingAxe" {
		t.Fatalf("axe identity = %#v", axe)
	}
	if axe.EquipmentSlot == nil || *axe.EquipmentSlot != SlotWeapon {
		t.Fatalf("axe equipment slot = %#v, want weapon", axe.EquipmentSlot)
	}
	if axe.CombatStats == nil || sword.CombatStats == nil {
		t.Fatal("axe and sword must have combat stats")
	}
	if axe.CombatStats.MinDamage >= sword.CombatStats.MinDamage ||
		axe.CombatStats.MaxDamage >= sword.CombatStats.MaxDamage ||
		axe.CombatStats.AccuracyBonus >= sword.CombatStats.AccuracyBonus {
		t.Fatalf("axe combat stats %#v are not modestly below sword %#v", axe.CombatStats, sword.CombatStats)
	}
}

func TestFishingRodIsAWeaponSlotToolWithoutCombatBonuses(t *testing.T) {
	rod := CreateFishingRod()
	if rod.Name != "Fishing Rod" || rod.Type != "fishingRod" || rod.RenderModel != "fishingRod" {
		t.Fatalf("rod identity = %#v", rod)
	}
	if rod.EquipmentSlot == nil || *rod.EquipmentSlot != SlotWeapon {
		t.Fatalf("rod equipment slot = %#v, want weapon", rod.EquipmentSlot)
	}
	if rod.CombatStats != nil {
		t.Fatalf("rod combat stats = %#v, want nil", rod.CombatStats)
	}
}

func TestMagicStaffCombatProfile(t *testing.T) {
	staff := CreateMagicStaff()
	stats := staff.CombatStats
	if stats == nil || stats.Range != 4 || stats.AttackSpeedTicks != 3 ||
		stats.AttackMethod != AttackMethodMagic || stats.WindUpTicks != 2 ||
		stats.TravelTicks != 1 || stats.ProjectileType != "magicBolt" {
		t.Fatalf("magic staff combat stats = %#v", stats)
	}
}
