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
