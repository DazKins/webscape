package component

import "webscape/server/game/model"

func CalculateCombatStats(base *CBaseStats, equipped *CEquipped) *CCombatStats {
	if base == nil {
		base = NewCBaseStats(5, 5, 5)
	}

	minDamage := 1 + base.GetStrength()/2
	maxDamage := 3 + base.GetStrength()
	accuracy := 50 + base.GetDexterity()*2
	evasion := base.GetDexterity()
	armor := base.GetVitality() / 2
	critChance := 0.05 + (float64(base.GetDexterity()) * 0.002)
	critMultiplier := 1.5
	attackRange := 1
	attackSpeedTicks := 2

	if equipped != nil {
		for slot, item := range equipped.GetAllEquippedItems() {
			if item == nil || item.CombatStats == nil {
				continue
			}
			stats := item.CombatStats
			minDamage += stats.MinDamage
			maxDamage += stats.MaxDamage
			accuracy += stats.AccuracyBonus
			armor += stats.ArmorBonus
			critChance += stats.CritBonus
			if slot == model.SlotWeapon {
				if stats.Range > 0 {
					attackRange = stats.Range
				}
				if stats.AttackSpeedTicks > 0 {
					attackSpeedTicks = stats.AttackSpeedTicks
				}
			}
		}
	}

	if maxDamage < minDamage {
		maxDamage = minDamage
	}
	if attackSpeedTicks < 1 {
		attackSpeedTicks = 1
	}
	if critChance < 0 {
		critChance = 0
	}

	return NewCCombatStats(
		minDamage,
		maxDamage,
		accuracy,
		evasion,
		armor,
		critChance,
		critMultiplier,
		attackRange,
		attackSpeedTicks,
	)
}
