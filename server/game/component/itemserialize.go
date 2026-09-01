package component

import (
	"webscape/server/game/model"
	"webscape/server/util"
)

func SerializeItem(item *model.Item) util.Json {
	itemObj := util.JObject(map[string]util.Json{
		"id":   util.JString(item.Id.String()),
		"name": util.JString(item.Name),
		"type": util.JString(item.Type),
	})
	if item.RenderModel != "" {
		itemObj["renderModel"] = util.JString(item.RenderModel)
	}
	if item.EquipmentSlot != nil {
		itemObj["equipmentSlot"] = util.JString(string(*item.EquipmentSlot))
	}
	if item.CombatStats != nil {
		attackMethod := item.CombatStats.AttackMethod
		if attackMethod == "" {
			attackMethod = model.AttackMethodMelee
		}
		itemObj["combatStats"] = util.JObject(map[string]util.Json{
			"minDamage":        util.JNumber(item.CombatStats.MinDamage),
			"maxDamage":        util.JNumber(item.CombatStats.MaxDamage),
			"accuracyBonus":    util.JNumber(item.CombatStats.AccuracyBonus),
			"armorBonus":       util.JNumber(item.CombatStats.ArmorBonus),
			"critBonus":        util.JNumber(item.CombatStats.CritBonus),
			"range":            util.JNumber(item.CombatStats.Range),
			"attackSpeedTicks": util.JNumber(item.CombatStats.AttackSpeedTicks),
			"attackMethod":     util.JString(string(attackMethod)),
			"windUpTicks":      util.JNumber(item.CombatStats.WindUpTicks),
			"travelTicks":      util.JNumber(item.CombatStats.TravelTicks),
			"projectileType":   util.JString(item.CombatStats.ProjectileType),
		})
	}
	return itemObj
}
