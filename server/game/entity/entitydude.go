package entity

import (
	"math/rand"
	"strconv"
	"webscape/server/game/component"
	"webscape/server/math"
)

func CreateDudeEntity(
	name string,
	position math.Vec2,
) *Entity {
	positionComponent := &component.CPosition{}
	positionComponent.SetPosition(position)

	metadataComponent := &component.CMetadata{}
	metadataComponent.SetMetadata(map[string]any{
		"name":  name,
		"color": "#" + strconv.FormatInt(int64(rand.Intn(0xffffff+1)), 16),
	})

	interactableComponent := &component.CInteractable{}
	interactableComponent.SetInteractionOptions([]component.InteractionOption{
		component.InteractionOptionTalk,
		component.InteractionOptionTrade,
		component.InteractionOptionAttack,
	})

	randomwalkComponent := &component.CRandomWalk{}
	randomwalkComponent.SetWalkTimer(10)

	return NewEntity(NewEntityId()).
		SetComponent(positionComponent).
		SetComponent(metadataComponent).
		SetComponent(interactableComponent).
		SetComponent(randomwalkComponent)
}
