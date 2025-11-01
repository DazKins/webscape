package entity

import (
	"webscape/server/game/component"
	"webscape/server/game/model"
	"webscape/server/math"
	"webscape/server/util"
)

func CreatePlayerEntity(id model.EntityId, name string) *Entity {
	positionComponent := &component.CPosition{
		Position: math.Vec2{X: 0, Y: 0},
	}

	metadataComponent := &component.CMetadata{
		Metadata: util.JObject(map[string]util.Json{
			"name": util.JString(name),
		}),
	}

	return NewEntity(id).
		SetComponent(positionComponent).
		SetComponent(metadataComponent)
}
