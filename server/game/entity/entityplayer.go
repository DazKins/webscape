package entity

import (
	"webscape/server/game/component"
	"webscape/server/game/model"
	"webscape/server/math"
	"webscape/server/util"
)

func CreatePlayerEntity(id model.EntityId, name string) []component.Component {
	positionComponent := component.NewCPosition(math.Vec2{X: 0, Y: 0})

	metadataComponent := component.NewCMetadata(util.JObject(map[string]util.Json{
		"name": util.JString(name),
	}))

	renderableComponent := component.NewCRenderable("human")

	healthComponent := component.NewCHealth(100, 100)

	inventoryComponent := component.NewCInventory()
	// Add some test items to the player's inventory
	inventoryComponent.AddItem(model.CreateIronSword())
	inventoryComponent.AddItem(model.CreateLeatherHelmet())
	inventoryComponent.AddItem(model.CreateHealthPotion())
	inventoryComponent.AddItem(model.CreateBread())
	inventoryComponent.AddItem(model.CreateIronOre())
	inventoryComponent.AddItem(model.CreateMysteriousKey())

	equippedComponent := component.NewCEquipped()
	baseStatsComponent := component.NewCBaseStats(8, 7, 7)
	combatStatsComponent := component.CalculateCombatStats(baseStatsComponent, equippedComponent)
	playerComponent := component.NewCPlayer(name)
	combatLogComponent := component.NewCCombatLog(10)

	return []component.Component{
		positionComponent,
		metadataComponent,
		renderableComponent,
		healthComponent,
		inventoryComponent,
		equippedComponent,
		baseStatsComponent,
		combatStatsComponent,
		playerComponent,
		combatLogComponent,
	}
}
