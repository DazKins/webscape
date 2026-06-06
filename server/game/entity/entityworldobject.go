package entity

import (
	"webscape/server/game/component"
	"webscape/server/game/world"
	"webscape/server/math"
	"webscape/server/util"
)

func CreateAuthoredEntity(entity world.WorldEntity) []component.Component {
	components := []component.Component{}
	position := createPositionComponent(entity.Components)
	if position != nil {
		components = append(components, position)
	}
	if spawn := createSpawnComponent(entity.Id, entity.Components, position); spawn != nil {
		components = append(components, spawn)
	}
	if metadata := createMetadataComponent(entity.Id, entity.Components); metadata != nil {
		components = append(components, metadata)
	}
	if renderable := createRenderableComponent(entity.Components); renderable != nil {
		components = append(components, renderable)
	}
	if lootable := createLootableComponent(entity.Components); lootable != nil {
		components = append(components, lootable)
	}
	if conversation := createConversationComponent(entity.Components); conversation != nil {
		components = append(components, conversation)
	}
	if randomWalk := createRandomWalkComponent(entity.Components); randomWalk != nil {
		components = append(components, randomWalk)
	}
	if health := createHealthComponent(entity.Components); health != nil {
		components = append(components, health)
	}
	baseStats := createBaseStatsComponent(entity.Components)
	if baseStats != nil {
		components = append(components, baseStats)
	}
	equipped := createEquippedComponent(entity.Components)
	if equipped != nil {
		components = append(components, equipped)
	}
	if combatStats := createCombatStatsComponent(entity.Components, baseStats, equipped); combatStats != nil {
		components = append(components, combatStats)
	}
	return components
}

func createPositionComponent(components map[string]any) *component.CPosition {
	raw, ok := components["position"].(map[string]any)
	if !ok {
		return nil
	}
	x, okX := numberToInt(raw["x"])
	y, okY := numberToInt(raw["y"])
	if !okX || !okY {
		return nil
	}
	return component.NewCPosition(math.Vec2{X: x, Y: y})
}

func createSpawnComponent(entityId string, components map[string]any, position *component.CPosition) *component.CSpawn {
	if position == nil {
		return nil
	}
	raw, ok := components["spawn"].(map[string]any)
	if !ok {
		return nil
	}
	respawnTicks, ok := numberToInt(raw["respawnTicks"])
	if !ok {
		respawnTicks = 0
	}
	rawTemplate, ok := raw["entity"].(map[string]any)
	if !ok {
		return nil
	}
	templateComponents, ok := rawTemplate["components"].(map[string]any)
	if !ok {
		return nil
	}
	return component.NewCSpawn(
		position.GetPosition(),
		respawnTicks,
		entityId+"_child",
		templateComponents,
	)
}

func createMetadataComponent(entityId string, components map[string]any) *component.CMetadata {
	raw, ok := components["metadata"].(map[string]any)
	if !ok {
		return nil
	}
	metadata, ok := jsonToUtil(raw).(util.JObject)
	if !ok {
		return nil
	}
	if _, ok := metadata["entityId"]; !ok {
		metadata["entityId"] = util.JString(entityId)
	}
	return component.NewCMetadata(metadata)
}

func createRenderableComponent(components map[string]any) *component.CRenderable {
	raw, ok := components["renderable"].(map[string]any)
	if !ok {
		return nil
	}
	renderType, _ := raw["type"].(string)
	if renderType == "" {
		return nil
	}
	return component.NewCRenderable(renderType)
}

func createLootableComponent(components map[string]any) *component.CLootable {
	if components == nil {
		return nil
	}
	rawLootable, ok := components["lootable"]
	if !ok {
		return nil
	}
	lootable, ok := rawLootable.(map[string]any)
	if !ok {
		return nil
	}

	items := []component.LootItem{}
	rawItems, ok := lootable["items"].([]any)
	if ok {
		for _, rawItem := range rawItems {
			item, ok := rawItem.(map[string]any)
			if !ok {
				continue
			}
			name, _ := item["name"].(string)
			itemType, _ := item["type"].(string)
			if name == "" || itemType == "" {
				continue
			}
			count := 1
			if rawCount, ok := item["count"].(float64); ok && rawCount >= 1 {
				count = int(rawCount)
			}
			items = append(items, component.LootItem{
				Name:  name,
				Type:  itemType,
				Count: count,
			})
		}
	}

	once, _ := lootable["once"].(bool)
	return component.NewCLootable(once, items)
}

func createConversationComponent(components map[string]any) *component.CConversation {
	raw, ok := components["conversation"].(map[string]any)
	if !ok {
		return nil
	}
	conversationId, _ := raw["conversationId"].(string)
	if conversationId == "" {
		return nil
	}
	return component.NewCConversation(conversationId)
}

func createRandomWalkComponent(components map[string]any) *component.CRandomWalk {
	raw, ok := components["randomwalk"].(map[string]any)
	if !ok {
		return nil
	}
	walkTimer, ok := numberToInt(raw["walkTimer"])
	if !ok {
		walkTimer = 10
	}
	return component.NewCRandomWalk(walkTimer)
}

func createHealthComponent(components map[string]any) *component.CHealth {
	raw, ok := components["health"].(map[string]any)
	if !ok {
		return nil
	}
	maxHealth, ok := numberToInt(raw["maxHealth"])
	if !ok || maxHealth < 1 {
		maxHealth = 100
	}
	currentHealth, ok := numberToInt(raw["currentHealth"])
	if !ok {
		currentHealth = maxHealth
	}
	return component.NewCHealth(maxHealth, currentHealth)
}

func createBaseStatsComponent(components map[string]any) *component.CBaseStats {
	raw, ok := components["basestats"].(map[string]any)
	if !ok {
		return nil
	}
	strength, ok := numberToInt(raw["strength"])
	if !ok {
		strength = 5
	}
	dexterity, ok := numberToInt(raw["dexterity"])
	if !ok {
		dexterity = 5
	}
	vitality, ok := numberToInt(raw["vitality"])
	if !ok {
		vitality = 5
	}
	return component.NewCBaseStats(strength, dexterity, vitality)
}

func createEquippedComponent(components map[string]any) *component.CEquipped {
	if _, ok := components["equipped"]; !ok {
		return nil
	}
	return component.NewCEquipped()
}

func createCombatStatsComponent(
	components map[string]any,
	baseStats *component.CBaseStats,
	equipped *component.CEquipped,
) *component.CCombatStats {
	raw, hasCombatStats := components["combatstats"].(map[string]any)
	if !hasCombatStats {
		return nil
	}
	if len(raw) == 0 {
		return component.CalculateCombatStats(baseStats, equipped)
	}
	minDamage, _ := numberToInt(raw["minDamage"])
	maxDamage, _ := numberToInt(raw["maxDamage"])
	accuracy, _ := numberToInt(raw["accuracy"])
	evasion, _ := numberToInt(raw["evasion"])
	armor, _ := numberToInt(raw["armor"])
	attackRange, _ := numberToInt(raw["attackRange"])
	attackSpeedTicks, _ := numberToInt(raw["attackSpeedTicks"])
	critChance, _ := numberToFloat(raw["critChance"])
	critMultiplier, _ := numberToFloat(raw["critMultiplier"])
	return component.NewCCombatStats(
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

func jsonToUtil(value any) util.Json {
	switch value := value.(type) {
	case nil:
		return util.JNull{}
	case bool:
		return util.JBool(value)
	case float64:
		return util.JNumber(value)
	case string:
		return util.JString(value)
	case []any:
		result := make(util.JArray, len(value))
		for i, item := range value {
			result[i] = jsonToUtil(item)
		}
		return result
	case map[string]any:
		result := make(util.JObject)
		for key, item := range value {
			result[key] = jsonToUtil(item)
		}
		return result
	default:
		return util.JString("")
	}
}

func numberToInt(value any) (int, bool) {
	switch value := value.(type) {
	case float64:
		return int(value), true
	case int:
		return value, true
	default:
		return 0, false
	}
}

func numberToFloat(value any) (float64, bool) {
	switch value := value.(type) {
	case float64:
		return value, true
	case int:
		return float64(value), true
	default:
		return 0, false
	}
}
