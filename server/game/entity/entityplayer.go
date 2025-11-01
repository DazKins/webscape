package entity

import (
	"webscape/server/game/component"
	"webscape/server/game/world"
	"webscape/server/math"
)

func CreatePlayerEntity(id EntityId, name string, world *world.World) *Entity {
	positionComponent := &component.CPosition{}
	positionComponent.SetPosition(math.Vec2{X: 0, Y: 0})

	metadataComponent := &component.CMetadata{}
	metadataComponent.SetMetadata(map[string]any{
		"name": name,
	})

	return NewEntity(id).
		SetComponent(positionComponent).
		SetComponent(metadataComponent)
}
