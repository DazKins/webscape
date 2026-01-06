package entity

import (
	"webscape/server/game/component"
	"webscape/server/game/model"
	"webscape/server/math"
	"webscape/server/util"
)

func CreatePlayerEntity(id model.EntityId, name string) []component.Component {
	positionComponent := &component.CPosition{
		Position: math.Vec2{X: 0, Y: 0},
	}

	metadataComponent := &component.CMetadata{
		Metadata: util.JObject(map[string]util.Json{
			"name": util.JString(name),
		}),
	}

	renderableComponent := &component.CRenderable{
		Type: "human",
	}

	healthComponent := &component.CHealth{
		MaxHealth:     100,
		CurrentHealth: 100,
	}

	return []component.Component{
		positionComponent,
		metadataComponent,
		renderableComponent,
		healthComponent,
	}
}
