package component

import (
	"webscape/server/game/model"
	"webscape/server/util"
)

// Definition metadata is resolved at the transport boundary for existing UI
// consumers. The authoritative instance and its saved payload contain only IDs
// and per-instance state.
func SerializeItem(item *model.Item) util.Json {
	itemObj := serializeItemPresentation(item.Definition(), item.Name(), item.CombatStats())
	itemObj["id"] = util.JString(item.Id.String())
	itemObj["quantity"] = util.JNumber(item.Quantity)
	itemObj["stackable"] = util.JBool(item.IsStackable())
	itemObj["hasProperties"] = util.JBool(item.HasProperties())
	return itemObj
}

func SerializeItemDefinition(id string) util.Json {
	definition, _ := model.GetItemDefinition(id)
	obj := serializeItemPresentation(definition, definition.Name, definition.CombatStats)
	obj["stackable"] = util.JBool(definition.Stackable)
	return obj
}

func serializeItemPresentation(definition model.ItemDefinition, name string, stats *model.ItemCombatStats) util.JObject {
	itemObj := util.JObject{
		"definitionId": util.JString(definition.ID),
		"name":         util.JString(name),
		"type":         util.JString(definition.Type),
		"renderModel":  util.JString(definition.RenderModel),
	}
	if definition.EquipmentSlot != "" {
		itemObj["equipmentSlot"] = util.JString(string(definition.EquipmentSlot))
	}
	if stats != nil {
		attackMethod := stats.AttackMethod
		if attackMethod == "" {
			attackMethod = model.AttackMethodMelee
		}
		itemObj["combatStats"] = util.JObject(map[string]util.Json{
			"minDamage":        util.JNumber(stats.MinDamage),
			"maxDamage":        util.JNumber(stats.MaxDamage),
			"accuracyBonus":    util.JNumber(stats.AccuracyBonus),
			"armorBonus":       util.JNumber(stats.ArmorBonus),
			"critBonus":        util.JNumber(stats.CritBonus),
			"range":            util.JNumber(stats.Range),
			"attackSpeedTicks": util.JNumber(stats.AttackSpeedTicks),
			"attackMethod":     util.JString(string(attackMethod)),
			"windUpTicks":      util.JNumber(stats.WindUpTicks),
			"travelTicks":      util.JNumber(stats.TravelTicks),
			"projectileType":   util.JString(stats.ProjectileType),
		})
	}
	return itemObj
}
