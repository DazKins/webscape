package system

import (
	"testing"
	"webscape/server/game/component"
	"webscape/server/game/gameevent"
	"webscape/server/game/model"
	"webscape/server/math"
)

func TestRangedCombatConsumesOneArrowWhenProjectileLaunches(t *testing.T) {
	system, tick, emitter, attackerId, targetId := newRangedCombatSystem(t, 2)
	targetHealth := healthOf(system.ComponentManager, targetId)
	startingHealth := targetHealth.GetCurrentHealth()

	tick.tick = 10
	system.Update()
	assertCombatPhase(t, system.ComponentManager, attackerId, component.CombatPhaseCasting, 10)
	assertArrowCount(t, system.ComponentManager, attackerId, 2)

	tick.tick = 11
	system.Update()
	assertArrowCount(t, system.ComponentManager, attackerId, 2)

	tick.tick = 12
	system.Update()
	assertCombatPhase(t, system.ComponentManager, attackerId, component.CombatPhaseRecovering, 12)
	assertArrowCount(t, system.ComponentManager, attackerId, 1)
	assertHealth(t, targetHealth, startingHealth)
	assertEventCount(t, emitter.events, gameevent.EventIdCombatProjectileLaunched, 1)
	launched := firstEvent(emitter.events, gameevent.EventIdCombatProjectileLaunched)
	projectile := launched.Payload.(gameevent.CombatProjectileLaunchedPayload)
	if projectile.ProjectileType != "arrow" {
		t.Fatalf("projectile type = %q, want arrow", projectile.ProjectileType)
	}

	tick.tick = 13
	system.Update()
	assertHealth(t, targetHealth, startingHealth-5)
	assertCombatPhase(t, system.ComponentManager, attackerId, component.CombatPhaseCasting, 13)
	resolved := firstEvent(emitter.events, gameevent.EventIdCombatResolved)
	payload := resolved.Payload.(gameevent.CombatResolvedPayload)
	if !payload.DidHit || payload.Damage != 5 || payload.AttackMethod != model.AttackMethodRanged {
		t.Fatalf("combat resolved payload = %#v", payload)
	}
}

func TestRangedCombatStopsWithoutArrows(t *testing.T) {
	system, tick, emitter, attackerId, targetId := newRangedCombatSystem(t, 0)
	targetPosition := system.ComponentManager.GetEntityComponent(component.ComponentIdPosition, targetId).(*component.CPosition)
	targetPosition.SetPosition(math.Vec2{X: 10, Y: 0})
	system.ComponentManager.SetEntityComponent(targetId, targetPosition)

	tick.tick = 20
	system.Update()

	if system.ComponentManager.GetEntityComponent(component.ComponentIdCombatState, attackerId) != nil {
		t.Fatal("ranged combat remained active without arrows")
	}
	if system.ComponentManager.GetEntityComponent(component.ComponentIdPathing, attackerId) != nil {
		t.Fatal("ranged attacker started approaching a target without arrows")
	}
	assertEventCount(t, emitter.events, gameevent.EventIdCombatProjectileLaunched, 0)
	entries := system.ComponentManager.GetEntityComponent(component.ComponentIdCombatLog, attackerId).(*component.CCombatLog).GetEntries()
	if len(entries) != 1 || entries[0].GetText() != "You have no arrows left" {
		t.Fatalf("combat log entries = %#v", entries)
	}
}

func TestRangedCombatCancelsIfArrowDisappearsDuringWindUp(t *testing.T) {
	system, tick, emitter, attackerId, _ := newRangedCombatSystem(t, 1)

	tick.tick = 10
	system.Update()
	assertCombatPhase(t, system.ComponentManager, attackerId, component.CombatPhaseCasting, 10)
	inventory := system.ComponentManager.GetEntityComponent(component.ComponentIdInventory, attackerId).(*component.CInventory)
	inventory.RemoveFirstItemByType(model.ItemTypeArrow)
	system.ComponentManager.SetEntityComponent(attackerId, inventory)

	tick.tick = 11
	system.Update()
	if system.ComponentManager.GetEntityComponent(component.ComponentIdCombatState, attackerId) != nil {
		t.Fatal("ranged wind-up remained active after its arrow disappeared")
	}
	assertEventCount(t, emitter.events, gameevent.EventIdCombatProjectileLaunched, 0)
}

func TestLastArrowLaunchesBeforeRangedCombatStops(t *testing.T) {
	system, tick, emitter, attackerId, targetId := newRangedCombatSystem(t, 1)
	startingHealth := healthOf(system.ComponentManager, targetId).GetCurrentHealth()

	tick.tick = 1
	system.Update()
	tick.tick = 3
	system.Update()
	assertArrowCount(t, system.ComponentManager, attackerId, 0)
	assertEventCount(t, emitter.events, gameevent.EventIdCombatProjectileLaunched, 1)

	tick.tick = 4
	system.Update()
	assertHealth(t, healthOf(system.ComponentManager, targetId), startingHealth-5)
	if system.ComponentManager.GetEntityComponent(component.ComponentIdCombatState, attackerId) != nil {
		t.Fatal("ranged combat remained active after the last arrow resolved")
	}
	assertEventCount(t, emitter.events, gameevent.EventIdCombatProjectileLaunched, 1)
}

func newRangedCombatSystem(
	t *testing.T,
	arrowCount int,
) (*CombatSystem, *mutableCombatTick, *recordingEventEmitter, model.EntityId, model.EntityId) {
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
	inventory := component.NewCInventory()
	for range arrowCount {
		inventory.AddItem(model.CreateArrow())
	}
	attackerId := manager.CreateNewEntity(
		component.NewCPosition(math.Vec2{X: 0, Y: 0}),
		component.NewCHealth(100, 100),
		component.NewCMetadata(nil),
		component.NewCCombatLog(10),
		inventory,
	)
	attackerStats := component.NewCCombatStats(7, 7, -10000, 0, 0, 0, 1.5, 3, 3)
	attackerStats.SetAttackProfile(model.AttackMethodRanged, 2, 1, "arrow")
	manager.SetEntityComponent(attackerId, attackerStats)
	targetId := manager.CreateNewEntity(
		component.NewCPosition(math.Vec2{X: 3, Y: 0}),
		component.NewCHealth(100, 100),
		component.NewCMetadata(nil),
		component.NewCCombatLog(10),
		component.NewCCombatStats(1, 1, 0, 10000, 2, 0, 1.5, 1, 2),
	)
	transitions.BeginCombat(attackerId, component.NewCCombatState(targetId))
	return system, tick, emitter, attackerId, targetId
}

func assertArrowCount(
	t *testing.T,
	manager *component.ComponentManager,
	entityId model.EntityId,
	want int,
) {
	t.Helper()
	inventory := manager.GetEntityComponent(component.ComponentIdInventory, entityId).(*component.CInventory)
	got := 0
	for _, item := range inventory.GetAllItems() {
		if item.Type == model.ItemTypeArrow {
			got++
		}
	}
	if got != want {
		t.Fatalf("arrow count = %d, want %d", got, want)
	}
}
