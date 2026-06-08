package system

import (
	"testing"
	"webscape/server/game/component"
	"webscape/server/game/model"
	"webscape/server/math"
)

func TestRandomWalkSystemSkipsConversationTarget(t *testing.T) {
	componentManager := component.NewComponentManager()
	npcEntityId := model.NewEntityId()
	playerEntityId := model.NewEntityId()

	componentManager.SetEntityComponent(npcEntityId, component.NewCPosition(math.Vec2{X: 4, Y: 4}))
	randomWalk := component.NewCRandomWalk(1, 5)
	randomWalk.SetOrigin(math.Vec2{X: 4, Y: 4})
	componentManager.SetEntityComponent(npcEntityId, randomWalk)
	componentManager.SetEntityComponent(
		playerEntityId,
		component.NewCActiveConversation("greeting", npcEntityId, "start"),
	)

	randomWalkSystem := RandomWalkSystem{
		SystemBase: SystemBase{
			ComponentManager: componentManager,
		},
	}

	randomWalkSystem.Update()

	position := componentManager.GetEntityComponent(
		component.ComponentIdPosition,
		npcEntityId,
	).(*component.CPosition)
	if !position.GetPosition().Eq(math.Vec2{X: 4, Y: 4}) {
		t.Fatalf("conversation target moved to %#v, want unchanged", position.GetPosition())
	}

	updatedRandomWalk := componentManager.GetEntityComponent(
		component.ComponentIdRandomWalk,
		npcEntityId,
	).(*component.CRandomWalk)
	if updatedRandomWalk.GetWalkTimer() != 1 {
		t.Fatalf("conversation target walk timer = %d, want 1", updatedRandomWalk.GetWalkTimer())
	}
}

func TestRandomWalkSystemStaysWithinMaxDistance(t *testing.T) {
	componentManager := component.NewComponentManager()
	npcEntityId := model.NewEntityId()
	origin := math.Vec2{X: 4, Y: 4}

	randomWalk := component.NewCRandomWalk(1, 0)
	randomWalk.SetOrigin(origin)
	componentManager.SetEntityComponent(npcEntityId, component.NewCPosition(origin))
	componentManager.SetEntityComponent(npcEntityId, randomWalk)

	randomWalkSystem := RandomWalkSystem{
		SystemBase: SystemBase{
			ComponentManager: componentManager,
		},
	}

	randomWalkSystem.Update()

	position := componentManager.GetEntityComponent(
		component.ComponentIdPosition,
		npcEntityId,
	).(*component.CPosition)
	if !position.GetPosition().Eq(origin) {
		t.Fatalf("position = %#v, want origin %#v", position.GetPosition(), origin)
	}
}
