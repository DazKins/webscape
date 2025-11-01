package entity

import (
	"math/rand"
	"strconv"
	"webscape/server/game/component"
	"webscape/server/game/model"
	"webscape/server/math"
	"webscape/server/util"
)

func CreateDudeEntity(
	name string,
	position math.Vec2,
) *Entity {
	positionComponent := &component.CPosition{
		Position: position,
	}

	metadataComponent := &component.CMetadata{
		Metadata: util.JObject(map[string]util.Json{
			"name":  util.JString(name),
			"color": util.JString("#" + strconv.FormatInt(int64(rand.Intn(0xffffff+1)), 16)),
		}),
	}

	interactableComponent := &component.CInteractable{
		InteractionOptions: []component.InteractionOption{
			component.InteractionOptionTalk,
			component.InteractionOptionTrade,
			component.InteractionOptionAttack,
		},
	}

	randomwalkComponent := &component.CRandomWalk{
		WalkTimer: 10,
	}

	return NewEntity(model.NewEntityId()).
		SetComponent(positionComponent).
		SetComponent(metadataComponent).
		SetComponent(interactableComponent).
		SetComponent(randomwalkComponent)
}
