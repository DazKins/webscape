package entity

import (
	"webscape/server/game/component"
	"webscape/server/math"
	"webscape/server/util"
)

func CreateDudeEntity(
	name string,
	entityType string,
	position math.Vec2,
) []component.Component {
	positionComponent := component.NewCPosition(position)

	metadataComponent := component.NewCMetadata(util.JObject(map[string]util.Json{
		"name":       util.JString(name),
		"entityType": util.JString(entityType),
	}))

	randomwalkComponent := component.NewCRandomWalk(10, 5)
	randomwalkComponent.SetOrigin(position)

	renderableComponent := component.NewCRenderable("human")
	appearanceComponent, _ := component.NewCAppearance(component.RandomAppearance())

	healthComponent := component.NewCHealth(100, 100)
	baseStatsComponent := component.NewCBaseStats(6, 5, 6)
	equippedComponent := component.NewCEquipped()
	combatStatsComponent := component.CalculateCombatStats(baseStatsComponent, equippedComponent)
	locomotionComponent := component.NewCLocomotion(component.LocomotionPhaseIdle, 0)

	return []component.Component{
		positionComponent,
		metadataComponent,
		randomwalkComponent,
		renderableComponent,
		appearanceComponent,
		healthComponent,
		baseStatsComponent,
		equippedComponent,
		combatStatsComponent,
		locomotionComponent,
	}
}
