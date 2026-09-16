package system

import (
	"testing"
	"webscape/server/game/component"
	"webscape/server/game/gameevent"
	"webscape/server/game/model"
	"webscape/server/util"
)

func TestResetCancelsIncomingAndOutgoingProjectiles(t *testing.T) {
	for _, resetAttacker := range []bool{false, true} {
		s, tick, emitter, attacker, target := newMagicCombatSystem(t)
		startingHealth := healthOf(s.ComponentManager, target).GetCurrentHealth()
		tick.tick = 10
		s.Update()
		tick.tick = 12
		s.Update()
		if len(s.pendingImpacts) != 1 {
			t.Fatal("projectile was not launched")
		}
		id := target
		if resetAttacker {
			id = attacker
		}
		s.ResetEntity(id)
		tick.tick = 13
		s.Update()
		assertHealth(t, healthOf(s.ComponentManager, target), startingHealth)
		assertEventCount(t, emitter.events, gameevent.EventIdCombatResolved, 0)
		if len(s.pendingImpacts) != 0 {
			t.Fatal("projectile survived reset")
		}
		if s.ComponentManager.GetEntityComponent(component.ComponentIdCombatState, attacker) != nil {
			t.Fatal("attack survived reset")
		}
	}
}

func TestResetStopsApproachingAttackButPreservesUnrelatedCombat(t *testing.T) {
	s, _, _, attacker, target := newMagicCombatSystem(t)
	approaching := model.NewEntityId()
	s.ComponentManager.SetEntityComponents(approaching,
		component.NewCInteracting(target, component.InteractionOptionAttack),
		component.NewCPathing(component.PathingTarget{EntityId: util.OptionalSome(target)}),
	)
	unrelated := model.NewEntityId()
	unrelatedState := component.NewCCombatState(attacker)
	s.ComponentManager.SetEntityComponent(unrelated, unrelatedState)
	s.pendingImpacts = []pendingCombatImpact{{attackerId: unrelated, targetId: attacker}, {attackerId: target, targetId: unrelated}}
	s.ResetEntity(target)
	if s.ComponentManager.GetEntityComponent(component.ComponentIdInteracting, approaching) != nil || s.ComponentManager.GetEntityComponent(component.ComponentIdPathing, approaching) != nil {
		t.Fatal("approaching attack survived reset")
	}
	if s.ComponentManager.GetEntityComponent(component.ComponentIdCombatState, unrelated) != unrelatedState {
		t.Fatal("unrelated combat was cleared")
	}
	if len(s.pendingImpacts) != 1 || s.pendingImpacts[0].attackerId != unrelated {
		t.Fatal("wrong projectiles cancelled")
	}
}
