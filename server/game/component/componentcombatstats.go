package component

import "webscape/server/util"

const ComponentIdBaseStats = ComponentId("basestats")
const ComponentIdCombatStats = ComponentId("combatstats")

type CBaseStats struct {
	strength  int
	dexterity int
	vitality  int
}

func NewCBaseStats(strength int, dexterity int, vitality int) *CBaseStats {
	return &CBaseStats{
		strength:  strength,
		dexterity: dexterity,
		vitality:  vitality,
	}
}

func (c *CBaseStats) GetId() ComponentId {
	return ComponentIdBaseStats
}

func (c *CBaseStats) Serialize() util.Json {
	return util.JObject(map[string]util.Json{
		"strength":  util.JNumber(c.strength),
		"dexterity": util.JNumber(c.dexterity),
		"vitality":  util.JNumber(c.vitality),
	})
}

func (c *CBaseStats) GetStrength() int {
	return c.strength
}

func (c *CBaseStats) SetStrength(strength int) {
	c.strength = strength
}

func (c *CBaseStats) GetDexterity() int {
	return c.dexterity
}

func (c *CBaseStats) SetDexterity(dexterity int) {
	c.dexterity = dexterity
}

func (c *CBaseStats) GetVitality() int {
	return c.vitality
}

func (c *CBaseStats) SetVitality(vitality int) {
	c.vitality = vitality
}

type CCombatStats struct {
	minDamage        int
	maxDamage        int
	accuracy         int
	evasion          int
	armor            int
	critChance       float64
	critMultiplier   float64
	attackRange      int
	attackSpeedTicks int
}

func NewCCombatStats(
	minDamage int,
	maxDamage int,
	accuracy int,
	evasion int,
	armor int,
	critChance float64,
	critMultiplier float64,
	attackRange int,
	attackSpeedTicks int,
) *CCombatStats {
	return &CCombatStats{
		minDamage:        minDamage,
		maxDamage:        maxDamage,
		accuracy:         accuracy,
		evasion:          evasion,
		armor:            armor,
		critChance:       critChance,
		critMultiplier:   critMultiplier,
		attackRange:      attackRange,
		attackSpeedTicks: attackSpeedTicks,
	}
}

func (c *CCombatStats) GetId() ComponentId {
	return ComponentIdCombatStats
}

func (c *CCombatStats) Serialize() util.Json {
	return util.JObject(map[string]util.Json{
		"minDamage":        util.JNumber(c.minDamage),
		"maxDamage":        util.JNumber(c.maxDamage),
		"accuracy":         util.JNumber(c.accuracy),
		"evasion":          util.JNumber(c.evasion),
		"armor":            util.JNumber(c.armor),
		"critChance":       util.JNumber(c.critChance),
		"critMultiplier":   util.JNumber(c.critMultiplier),
		"attackRange":      util.JNumber(c.attackRange),
		"attackSpeedTicks": util.JNumber(c.attackSpeedTicks),
	})
}

func (c *CCombatStats) GetMinDamage() int {
	return c.minDamage
}

func (c *CCombatStats) SetMinDamage(minDamage int) {
	c.minDamage = minDamage
}

func (c *CCombatStats) GetMaxDamage() int {
	return c.maxDamage
}

func (c *CCombatStats) SetMaxDamage(maxDamage int) {
	c.maxDamage = maxDamage
}

func (c *CCombatStats) GetAccuracy() int {
	return c.accuracy
}

func (c *CCombatStats) SetAccuracy(accuracy int) {
	c.accuracy = accuracy
}

func (c *CCombatStats) GetEvasion() int {
	return c.evasion
}

func (c *CCombatStats) SetEvasion(evasion int) {
	c.evasion = evasion
}

func (c *CCombatStats) GetArmor() int {
	return c.armor
}

func (c *CCombatStats) SetArmor(armor int) {
	c.armor = armor
}

func (c *CCombatStats) GetCritChance() float64 {
	return c.critChance
}

func (c *CCombatStats) SetCritChance(critChance float64) {
	c.critChance = critChance
}

func (c *CCombatStats) GetCritMultiplier() float64 {
	return c.critMultiplier
}

func (c *CCombatStats) SetCritMultiplier(critMultiplier float64) {
	c.critMultiplier = critMultiplier
}

func (c *CCombatStats) GetAttackRange() int {
	return c.attackRange
}

func (c *CCombatStats) SetAttackRange(attackRange int) {
	c.attackRange = attackRange
}

func (c *CCombatStats) GetAttackSpeedTicks() int {
	return c.attackSpeedTicks
}

func (c *CCombatStats) SetAttackSpeedTicks(attackSpeedTicks int) {
	c.attackSpeedTicks = attackSpeedTicks
}
