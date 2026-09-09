package component

import (
	"webscape/server/game/model"
)

type savedCombatStats struct {
	MinDamage        int                `json:"minDamage"`
	MaxDamage        int                `json:"maxDamage"`
	Accuracy         int                `json:"accuracy"`
	Evasion          int                `json:"evasion"`
	Armor            int                `json:"armor"`
	CritChance       float64            `json:"critChance"`
	CritMultiplier   float64            `json:"critMultiplier"`
	AttackRange      int                `json:"attackRange"`
	AttackSpeedTicks int                `json:"attackSpeedTicks"`
	AttackMethod     model.AttackMethod `json:"attackMethod"`
	WindUpTicks      int                `json:"windUpTicks"`
	TravelTicks      int                `json:"travelTicks"`
	ProjectileType   string             `json:"projectileType"`
}

func (c *CCombatStats) Save() (SavedComponent, error) {
	return marshalSaved(savedCombatStats{MinDamage: c.minDamage, MaxDamage: c.maxDamage, Accuracy: c.accuracy, Evasion: c.evasion, Armor: c.armor, CritChance: c.critChance, CritMultiplier: c.critMultiplier, AttackRange: c.attackRange, AttackSpeedTicks: c.attackSpeedTicks, AttackMethod: c.attackMethod, WindUpTicks: c.windUpTicks, TravelTicks: c.travelTicks, ProjectileType: c.projectileType})
}
func init() { registerComponentRestore(ComponentIdCombatStats, 1, restoreCombatStats) }
func restoreCombatStats(saved SavedComponent) (Component, error) {
	var s savedCombatStats
	if err := decodeSaved(saved.Data, &s); err != nil {
		return nil, err
	}
	return &CCombatStats{minDamage: s.MinDamage, maxDamage: s.MaxDamage, accuracy: s.Accuracy, evasion: s.Evasion, armor: s.Armor, critChance: s.CritChance, critMultiplier: s.CritMultiplier, attackRange: s.AttackRange, attackSpeedTicks: s.AttackSpeedTicks, attackMethod: s.AttackMethod, windUpTicks: s.WindUpTicks, travelTicks: s.TravelTicks, projectileType: s.ProjectileType}, nil
}
