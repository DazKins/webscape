package entity

import (
	"fmt"
	"math/rand"
	"webscape/server/game/entity/component"
	"webscape/server/game/world"
	"webscape/server/math"
	"webscape/server/util"
)

func CreateDudeEntity(
	world *world.World,
	id util.Optional[EntityId],
	name string,
	position math.Vec2,
) *Entity {
	positionComponent := component.NewCPosition(position)
	metadataComponent := component.NewCMetadata(map[string]any{
		"name":  name,
		"color": fmt.Sprintf("#%06X", rand.Intn(0xFFFFFF)),
	})
	interactableComponent := component.NewCInteractable([]component.InteractionOption{
		component.InteractionOptionTalk,
		component.InteractionOptionTrade,
		component.InteractionOptionAttack,
	})
	randomWalkComponent := component.NewCRandomWalk(world, positionComponent)

	return NewEntity(id).
		AddComponent(positionComponent).
		AddComponent(metadataComponent).
		AddComponent(interactableComponent).
		AddComponent(randomWalkComponent)
}
