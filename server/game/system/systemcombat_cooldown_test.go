package system

import (
	"testing"
	"webscape/server/game/component"
	"webscape/server/game/gameevent"
)

func TestCancelledCastPreservesAttackCooldown(t *testing.T) {
	system, tick, emitter, attacker, target := newMagicCombatSystem(t)
	tick.tick = 10
	system.Update()
	assertCombatPhase(t, system.ComponentManager, attacker, component.CombatPhaseCasting, 10)

	// Replacing an unlaunched cast cancels the projectile, but not its cadence.
	system.StateTransitions.EndCombat(attacker)
	system.StateTransitions.BeginCombat(attacker, component.NewCCombatState(target))
	for tick.tick = 11; tick.tick <= 12; tick.tick++ {
		system.Update()
		state := system.ComponentManager.GetEntityComponent(component.ComponentIdCombatState, attacker).(*component.CCombatState)
		if state.GetPhase() != component.CombatPhaseApproaching {
			t.Fatalf("tick %d: new cast started before cooldown expired", tick.tick)
		}
	}
	system.Update() // tick 13: the original cooldown expires.
	assertCombatPhase(t, system.ComponentManager, attacker, component.CombatPhaseCasting, 13)
	assertEventCount(t, emitter.events, gameevent.EventIdCombatProjectileLaunched, 0)
	tick.tick = 15
	system.Update()
	assertEventCount(t, emitter.events, gameevent.EventIdCombatProjectileLaunched, 1)
}
