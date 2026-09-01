package system

import (
	"testing"
	"webscape/server/game/component"
	"webscape/server/game/gameevent"
	"webscape/server/game/model"
	"webscape/server/math"
)

type mutableCombatTick struct{ tick uint64 }

func (s *mutableCombatTick) CurrentTick() uint64 { return s.tick }

func TestMagicCombatWindsUpTravelsAndStartsOnCadence(t *testing.T) {
	system, tick, emitter, attackerId, targetId := newMagicCombatSystem(t)
	targetHealth := healthOf(system.ComponentManager, targetId)
	startingHealth := targetHealth.GetCurrentHealth()

	tick.tick = 10
	system.Update()
	assertCombatPhase(t, system.ComponentManager, attackerId, component.CombatPhaseCasting, 10)
	assertHealth(t, targetHealth, startingHealth)
	assertEventCount(t, emitter.events, gameevent.EventIdCombatProjectileLaunched, 0)

	tick.tick = 11
	system.Update()
	assertHealth(t, targetHealth, startingHealth)
	assertEventCount(t, emitter.events, gameevent.EventIdCombatProjectileLaunched, 0)

	tick.tick = 12
	system.Update()
	assertCombatPhase(t, system.ComponentManager, attackerId, component.CombatPhaseRecovering, 12)
	assertHealth(t, targetHealth, startingHealth)
	assertEventCount(t, emitter.events, gameevent.EventIdCombatProjectileLaunched, 1)

	tick.tick = 13
	system.Update()
	assertHealth(t, targetHealth, startingHealth-5)
	assertCombatPhase(t, system.ComponentManager, attackerId, component.CombatPhaseCasting, 13)
	assertEventCount(t, emitter.events, gameevent.EventIdCombatResolved, 1)
	resolved := firstEvent(emitter.events, gameevent.EventIdCombatResolved)
	payload := resolved.Payload.(gameevent.CombatResolvedPayload)
	if !payload.DidHit || payload.Damage != 5 || payload.AttackMethod != model.AttackMethodMagic {
		t.Fatalf("combat resolved payload = %#v", payload)
	}
}

func TestMagicCastLocksTargetMovement(t *testing.T) {
	system, tick, emitter, attackerId, targetId := newMagicCombatSystem(t)
	startingHealth := healthOf(system.ComponentManager, targetId).GetCurrentHealth()

	tick.tick = 20
	system.Update()
	targetPosition := system.ComponentManager.GetEntityComponent(component.ComponentIdPosition, targetId).(*component.CPosition)
	targetPosition.SetPosition(math.Vec2{X: 12, Y: 12})
	system.ComponentManager.SetEntityComponent(targetId, targetPosition)

	tick.tick = 21
	system.Update()
	tick.tick = 22
	system.Update()
	assertEventCount(t, emitter.events, gameevent.EventIdCombatProjectileLaunched, 1)
	tick.tick = 23
	system.Update()
	assertHealth(t, healthOf(system.ComponentManager, targetId), startingHealth-5)
	if system.ComponentManager.GetEntityComponent(component.ComponentIdPathing, attackerId) == nil {
		t.Fatal("attacker did not resume approaching the moved target after the locked cast")
	}
}

func TestUnlaunchedMagicCastIsCancelledButLaunchedImpactSurvivesCasterDeath(t *testing.T) {
	t.Run("unlaunched movement cancellation", func(t *testing.T) {
		system, tick, emitter, attackerId, targetId := newMagicCombatSystem(t)
		startingHealth := healthOf(system.ComponentManager, targetId).GetCurrentHealth()
		tick.tick = 1
		system.Update()
		system.entityStateTransitions(tick).BeginPathing(attackerId, component.NewCPathing(component.PathingTarget{}))
		tick.tick = 4
		system.Update()
		assertHealth(t, healthOf(system.ComponentManager, targetId), startingHealth)
		assertEventCount(t, emitter.events, gameevent.EventIdCombatProjectileLaunched, 0)
	})

	t.Run("launched impact", func(t *testing.T) {
		system, tick, _, attackerId, targetId := newMagicCombatSystem(t)
		startingHealth := healthOf(system.ComponentManager, targetId).GetCurrentHealth()
		tick.tick = 1
		system.Update()
		tick.tick = 3
		system.Update()
		system.entityStateTransitions(tick).HandleDeath(attackerId)
		healthOf(system.ComponentManager, attackerId).SetCurrentHealth(0)
		tick.tick = 4
		system.Update()
		assertHealth(t, healthOf(system.ComponentManager, targetId), startingHealth-5)
	})
}

func TestMagicKillClearsLaterImpactsBeforePlayerRespawn(t *testing.T) {
	system, tick, _, attackerId, targetId := newMagicCombatSystem(t)
	targetHealth := healthOf(system.ComponentManager, targetId)
	targetHealth.SetCurrentHealth(4)
	system.ComponentManager.SetEntityComponent(targetId, component.NewCPlayer("target"))
	system.pendingImpacts = []pendingCombatImpact{
		{attackerId: attackerId, targetId: targetId, attackerName: "mage", minDamage: 7, maxDamage: 7, critMultiplier: 1.5, attackMethod: model.AttackMethodMagic, impactTick: 5},
		{attackerId: attackerId, targetId: targetId, attackerName: "mage", minDamage: 7, maxDamage: 7, critMultiplier: 1.5, attackMethod: model.AttackMethodMagic, impactTick: 6},
	}
	tick.tick = 5
	system.resolvePendingImpacts()
	if len(system.pendingImpacts) != 0 {
		t.Fatalf("pending impacts = %#v, want cleared after kill", system.pendingImpacts)
	}
	(&HealthSystem{SystemBase: system.SystemBase, TickSource: tick}).Update()
	respawnedHealth := healthOf(system.ComponentManager, targetId)
	if respawnedHealth.GetCurrentHealth() != respawnedHealth.GetMaxHealth() {
		t.Fatal("player did not respawn for test")
	}
	tick.tick = 6
	system.resolvePendingImpacts()
	assertHealth(t, respawnedHealth, respawnedHealth.GetMaxHealth())
}

func TestDeathOutsideCombatInvalidatesPendingImpactsBeforeRespawn(t *testing.T) {
	system, tick, _, attackerId, targetId := newMagicCombatSystem(t)
	system.pendingImpacts = []pendingCombatImpact{{
		attackerId: attackerId, targetId: targetId, minDamage: 7, maxDamage: 7,
		critMultiplier: 1.5, attackMethod: model.AttackMethodMagic, impactTick: 8,
	}}
	targetHealth := healthOf(system.ComponentManager, targetId)
	targetHealth.SetCurrentHealth(0)
	system.ComponentManager.SetEntityComponent(targetId, component.NewCPlayer("target"))
	healthSystem := &HealthSystem{
		SystemBase: system.SystemBase, TickSource: tick, CombatImpactInvalidator: system,
	}
	healthSystem.Update()
	if len(system.pendingImpacts) != 0 {
		t.Fatalf("pending impacts = %#v, want none after death", system.pendingImpacts)
	}
	tick.tick = 8
	system.resolvePendingImpacts()
	assertHealth(t, targetHealth, targetHealth.GetMaxHealth())
}

func newMagicCombatSystem(t *testing.T) (*CombatSystem, *mutableCombatTick, *recordingEventEmitter, model.EntityId, model.EntityId) {
	t.Helper()
	manager := component.NewComponentManager()
	tick := &mutableCombatTick{}
	emitter := &recordingEventEmitter{}
	transitions := NewEntityStateTransitions(manager, tick)
	system := &CombatSystem{
		SystemBase:   SystemBase{ComponentManager: manager, StateTransitions: transitions},
		EventEmitter: emitter,
		TickSource:   tick,
	}
	attackerId := manager.CreateNewEntity(
		component.NewCPosition(math.Vec2{X: 0, Y: 0}),
		component.NewCHealth(100, 100),
		component.NewCMetadata(nil),
		component.NewCCombatLog(10),
	)
	attackerStats := component.NewCCombatStats(7, 7, -10000, 0, 0, 0, 1.5, 4, 3)
	attackerStats.SetAttackProfile(model.AttackMethodMagic, 2, 1, "magicBolt")
	manager.SetEntityComponent(attackerId, attackerStats)
	targetId := manager.CreateNewEntity(
		component.NewCPosition(math.Vec2{X: 4, Y: 0}),
		component.NewCHealth(100, 100),
		component.NewCMetadata(nil),
		component.NewCCombatLog(10),
		component.NewCCombatStats(1, 1, 0, 10000, 2, 0, 1.5, 1, 2),
	)
	transitions.BeginCombat(attackerId, component.NewCCombatState(targetId))
	return system, tick, emitter, attackerId, targetId
}

func healthOf(manager *component.ComponentManager, entityId model.EntityId) *component.CHealth {
	return manager.GetEntityComponent(component.ComponentIdHealth, entityId).(*component.CHealth)
}

func assertHealth(t *testing.T, health *component.CHealth, want int) {
	t.Helper()
	if health.GetCurrentHealth() != want {
		t.Fatalf("health = %d, want %d", health.GetCurrentHealth(), want)
	}
}

func assertCombatPhase(t *testing.T, manager *component.ComponentManager, entityId model.EntityId, phase component.CombatPhase, started uint64) {
	t.Helper()
	state := manager.GetEntityComponent(component.ComponentIdCombatState, entityId).(*component.CCombatState)
	if state.GetPhase() != phase || state.GetPhaseStartedTick() != started {
		t.Fatalf("combat state phase = %s tick = %d, want %s tick %d", state.GetPhase(), state.GetPhaseStartedTick(), phase, started)
	}
}

func assertEventCount(t *testing.T, events []gameevent.Event, id string, want int) {
	t.Helper()
	count := 0
	for _, event := range events {
		if event.Id == id {
			count++
		}
	}
	if count != want {
		t.Fatalf("event %q count = %d, want %d", id, count, want)
	}
}

func firstEvent(events []gameevent.Event, id string) gameevent.Event {
	for _, event := range events {
		if event.Id == id {
			return event
		}
	}
	return gameevent.Event{}
}
