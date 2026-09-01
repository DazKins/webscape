package system

import (
	"testing"
	"webscape/server/game/component"
	"webscape/server/math"
)

func TestAttackInteractionUsesEquippedCombatRange(t *testing.T) {
	manager := component.NewComponentManager()
	targetId := manager.CreateNewEntity(component.NewCPosition(math.Vec2{X: 4, Y: 0}))
	attackerId := manager.CreateNewEntity(
		component.NewCPosition(math.Vec2{X: 0, Y: 0}),
		component.NewCInteracting(targetId, component.InteractionOptionAttack),
		component.NewCCombatStats(1, 1, 0, 0, 0, 0, 1.5, 4, 3),
	)
	system := InteractionSystem{SystemBase: SystemBase{ComponentManager: manager}}

	system.Update()

	combat := manager.GetEntityComponent(component.ComponentIdCombatState, attackerId)
	if combat == nil || combat.(*component.CCombatState).GetTargetId() != targetId {
		t.Fatalf("combat state = %#v, want target %s at range 4", combat, targetId)
	}
	if manager.GetEntityComponent(component.ComponentIdInteracting, attackerId) != nil {
		t.Fatal("attack interaction was not finished at weapon range")
	}
}
