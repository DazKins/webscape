package entity

import (
	"webscape/server/game/entity/component"
	"webscape/server/game/world"
	"webscape/server/math"
	"webscape/server/util"
)

func CreatePlayerEntity(id util.Optional[EntityId], name string, world *world.World) *Entity {
	positionComponent := component.NewCPosition(math.Vec2{X: 0, Y: 0})
	pathingComponent := component.NewCPathing(world, positionComponent)
	metadataComponent := component.NewCMetadata(map[string]any{
		"name": name,
	})

	return NewEntity(id).
		AddComponent(positionComponent).
		AddComponent(pathingComponent).
		AddComponent(metadataComponent)
}
