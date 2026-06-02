package entity

import (
	"math/rand"
	"strconv"
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
		"color":      util.JString("#" + strconv.FormatInt(int64(rand.Intn(0xffffff+1)), 16)),
	}))

	randomwalkComponent := component.NewCRandomWalk(10)

	renderableComponent := component.NewCRenderable("human")

	healthComponent := component.NewCHealth(100, 100)
	baseStatsComponent := component.NewCBaseStats(6, 5, 6)
	equippedComponent := component.NewCEquipped()
	combatStatsComponent := component.CalculateCombatStats(baseStatsComponent, equippedComponent)

	return []component.Component{
		positionComponent,
		metadataComponent,
		randomwalkComponent,
		renderableComponent,
		healthComponent,
		baseStatsComponent,
		equippedComponent,
		combatStatsComponent,
	}
}
