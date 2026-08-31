package system

import (
	"testing"
	"webscape/server/game/component"
	"webscape/server/game/model"
	"webscape/server/math"
)

func TestMarkLocomotionMovingCancelsFishing(t *testing.T) {
	manager := component.NewComponentManager()
	targetEntityId := model.NewEntityId()
	entityId := manager.CreateNewEntity(
		component.NewCLocomotion(component.LocomotionPhaseIdle, 1),
		component.NewCFishing(targetEntityId, 1, math.Vec2{}),
	)

	transitions := NewEntityStateTransitions(manager, fixedTickSource(2))
	transitions.BeginMoving(entityId)

	if manager.GetEntityComponent(component.ComponentIdFishing, entityId) != nil {
		t.Fatal("movement did not cancel fishing")
	}
	locomotion := manager.GetEntityComponent(
		component.ComponentIdLocomotion,
		entityId,
	).(*component.CLocomotion)
	if locomotion.GetPhase() != component.LocomotionPhaseMoving ||
		locomotion.GetPhaseStartedTick() != 2 || locomotion.GetLastMovementTick() != 2 {
		t.Fatalf(
			"locomotion = %q started %d last moved %d",
			locomotion.GetPhase(),
			locomotion.GetPhaseStartedTick(),
			locomotion.GetLastMovementTick(),
		)
	}
}
