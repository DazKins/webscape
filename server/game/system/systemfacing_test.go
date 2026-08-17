package system

import (
	"testing"
	"webscape/server/game/component"
	"webscape/server/game/model"
	"webscape/server/math"
)

func TestFacingSystemFacesConversationParticipantsAndCleansUp(t *testing.T) {
	manager := component.NewComponentManager()
	playerId := model.NewEntityId()
	targetId := model.NewEntityId()
	manager.SetEntityComponents(playerId,
		component.NewCPosition(math.Vec2{X: 1, Y: 1}),
		component.NewCActiveConversation("greeting", targetId, "start"),
	)
	manager.SetEntityComponent(targetId, component.NewCPosition(math.Vec2{X: 2, Y: 1}))
	system := FacingSystem{SystemBase: SystemBase{ComponentManager: manager}}

	system.Update()
	assertFacingTarget(t, manager, playerId, targetId)
	assertFacingTarget(t, manager, targetId, playerId)

	manager.RemoveComponent(component.ComponentIdActiveConversation, playerId)
	system.Update()
	assertNoFacing(t, manager, playerId)
	assertNoFacing(t, manager, targetId)
}

func TestFacingSystemCombatOverridesConversation(t *testing.T) {
	manager := component.NewComponentManager()
	entityId := model.NewEntityId()
	conversationTargetId := model.NewEntityId()
	combatTargetId := model.NewEntityId()
	manager.SetEntityComponents(entityId,
		component.NewCPosition(math.Vec2{X: 1, Y: 1}),
		component.NewCActiveConversation("greeting", conversationTargetId, "start"),
		component.NewCCombatState(combatTargetId),
	)
	manager.SetEntityComponent(conversationTargetId, component.NewCPosition(math.Vec2{X: 2, Y: 1}))
	manager.SetEntityComponent(combatTargetId, component.NewCPosition(math.Vec2{X: 3, Y: 1}))
	system := FacingSystem{SystemBase: SystemBase{ComponentManager: manager}}

	system.Update()
	assertFacingTarget(t, manager, entityId, combatTargetId)
	assertFacingTarget(t, manager, combatTargetId, entityId)
}

func TestFacingSystemPrefersNearestIncomingRelationship(t *testing.T) {
	manager := component.NewComponentManager()
	targetId := model.NewEntityId()
	nearAttackerId := model.NewEntityId()
	farAttackerId := model.NewEntityId()
	manager.SetEntityComponent(targetId, component.NewCPosition(math.Vec2{X: 0, Y: 0}))
	manager.SetEntityComponents(nearAttackerId,
		component.NewCPosition(math.Vec2{X: 1, Y: 0}),
		component.NewCCombatState(targetId),
	)
	manager.SetEntityComponents(farAttackerId,
		component.NewCPosition(math.Vec2{X: 4, Y: 0}),
		component.NewCCombatState(targetId),
	)
	system := FacingSystem{SystemBase: SystemBase{ComponentManager: manager}}

	system.Update()
	assertFacingTarget(t, manager, targetId, nearAttackerId)
}

func TestFacingSystemIgnoresRelationshipWithMissingTarget(t *testing.T) {
	manager := component.NewComponentManager()
	entityId := model.NewEntityId()
	missingTargetId := model.NewEntityId()
	manager.SetEntityComponents(entityId,
		component.NewCPosition(math.Vec2{X: 1, Y: 1}),
		component.NewCCombatState(missingTargetId),
		component.NewCFacing(missingTargetId),
	)
	system := FacingSystem{SystemBase: SystemBase{ComponentManager: manager}}

	system.Update()
	assertNoFacing(t, manager, entityId)
}

func assertFacingTarget(
	t *testing.T,
	manager *component.ComponentManager,
	entityId model.EntityId,
	wantTargetId model.EntityId,
) {
	t.Helper()
	value := manager.GetEntityComponent(component.ComponentIdFacing, entityId)
	if value == nil {
		t.Fatalf("entity %q has no facing component", entityId)
	}
	gotTargetId := value.(*component.CFacing).GetTargetEntityId()
	if gotTargetId != wantTargetId {
		t.Fatalf("entity %q faces %q, want %q", entityId, gotTargetId, wantTargetId)
	}
}

func assertNoFacing(t *testing.T, manager *component.ComponentManager, entityId model.EntityId) {
	t.Helper()
	if value := manager.GetEntityComponent(component.ComponentIdFacing, entityId); value != nil {
		t.Fatalf("entity %q unexpectedly has facing %#v", entityId, value)
	}
}
