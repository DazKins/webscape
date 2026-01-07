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

	equippedComponent := component.NewCEquipped()

	return []component.Component{
		positionComponent,
		metadataComponent,
		renderableComponent,
		healthComponent,
		inventoryComponent,
		equippedComponent,
	}
}
