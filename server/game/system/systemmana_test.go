package system

import (
	"testing"
	"webscape/server/game/component"
	"webscape/server/game/gameevent"
	"webscape/server/game/model"
)

func TestMagicSpendsManaOnlyOnceAtLaunch(t *testing.T) {
	s, tick, emitter, attacker, target := newMagicCombatSystem(t)
	mana := component.NewCMana(100, 10)
	s.ComponentManager.SetEntityComponent(attacker, mana)
	tick.tick = 1
	s.Update()
	if mana.GetCurrentMana() != 10 {
		t.Fatal("wind-up spent mana")
	}
	tick.tick = 2
	s.Update()
	if mana.GetCurrentMana() != 10 {
		t.Fatal("casting spent mana before launch")
	}
	tick.tick = 3
	s.Update()
	s.Update() // Re-evaluation during recovery must not launch or charge twice.
	if mana.GetCurrentMana() != 0 {
		t.Fatalf("launch left %d mana, want 0", mana.GetCurrentMana())
	}
	assertEventCount(t, emitter.events, gameevent.EventIdCombatProjectileLaunched, 1)
	tick.tick = 4
	s.Update()
	assertHealth(t, healthOf(s.ComponentManager, target), 95)
	if mana.GetCurrentMana() != 0 {
		t.Fatal("impact changed mana")
	}
	if s.ComponentManager.GetEntityComponent(component.ComponentIdCombatState, attacker) != nil {
		t.Fatal("empty mana did not stop next attack")
	}
	entries := s.ComponentManager.GetEntityComponent(component.ComponentIdCombatLog, attacker).(*component.CCombatLog).GetEntries()
	if entries[len(entries)-1].GetText() != "Not enough mana" {
		t.Fatalf("missing mana feedback: %v", entries)
	}
}

func TestMagicRejectsInsufficientOrMissingMana(t *testing.T) {
	for _, missing := range []bool{false, true} {
		s, tick, emitter, attacker, target := newMagicCombatSystem(t)
		s.ComponentManager.SetEntityComponent(attacker, component.NewCMana(100, 9))
		if missing {
			s.ComponentManager.RemoveComponent(component.ComponentIdMana, attacker)
		}
		for n := uint64(1); n <= 5; n++ {
			tick.tick = n
			s.Update()
		}
		assertEventCount(t, emitter.events, gameevent.EventIdCombatProjectileLaunched, 0)
		assertHealth(t, healthOf(s.ComponentManager, target), 100)
		if s.ComponentManager.GetEntityComponent(component.ComponentIdCombatState, attacker) != nil {
			t.Fatal("unaffordable cast started")
		}
		entries := s.ComponentManager.GetEntityComponent(component.ComponentIdCombatLog, attacker).(*component.CCombatLog).GetEntries()
		if len(entries) != 1 || entries[0].GetText() != "Not enough mana" {
			t.Fatalf("mana feedback repeated or missing: %v", entries)
		}
	}
}

func TestMagicRechecksManaAtLaunch(t *testing.T) {
	s, tick, emitter, attacker, _ := newMagicCombatSystem(t)
	tick.tick = 1
	s.Update()
	mana := s.ComponentManager.GetEntityComponent(component.ComponentIdMana, attacker).(*component.CMana)
	mana.Spend(95)
	tick.tick = 3
	s.Update()
	assertEventCount(t, emitter.events, gameevent.EventIdCombatProjectileLaunched, 0)
	if mana.GetCurrentMana() != 5 {
		t.Fatal("rejected launch spent mana")
	}
}

func TestCancelledCastDoesNotSpendMana(t *testing.T) {
	for _, reason := range []string{"movement", "target death", "caster death"} {
		t.Run(reason, func(t *testing.T) {
			s, tick, emitter, attacker, target := newMagicCombatSystem(t)
			tick.tick = 1
			s.Update()
			switch reason {
			case "movement":
				s.entityStateTransitions(tick).BeginPathing(attacker, component.NewCPathing(component.PathingTarget{}))
			case "target death":
				healthOf(s.ComponentManager, target).SetCurrentHealth(0)
			case "caster death":
				healthOf(s.ComponentManager, attacker).SetCurrentHealth(0)
			}
			tick.tick = 3
			s.Update()
			mana := s.ComponentManager.GetEntityComponent(component.ComponentIdMana, attacker).(*component.CMana)
			if mana.GetCurrentMana() != 100 {
				t.Fatal("cancelled cast spent mana")
			}
			assertEventCount(t, emitter.events, gameevent.EventIdCombatProjectileLaunched, 0)
		})
	}
}

func TestLaunchedSpellDoesNotRefundManaWhenTargetDies(t *testing.T) {
	s, tick, _, attacker, target := newMagicCombatSystem(t)
	tick.tick = 1
	s.Update()
	tick.tick = 3
	s.Update()
	healthOf(s.ComponentManager, target).SetCurrentHealth(0)
	tick.tick = 4
	s.Update()
	mana := s.ComponentManager.GetEntityComponent(component.ComponentIdMana, attacker).(*component.CMana)
	if mana.GetCurrentMana() != 90 {
		t.Fatal("launched spell refunded mana")
	}
}

func TestManaRegenerationCadenceCapAndRespawn(t *testing.T) {
	manager := component.NewComponentManager()
	tick := &mutableCombatTick{tick: 1}
	mana := component.NewCMana(100, 98)
	health := component.NewCHealth(100, 100)
	id := manager.CreateNewEntity(mana, health, component.NewCPlayer("Player"))
	s := &ManaSystem{SystemBase: SystemBase{ComponentManager: manager}, TickSource: tick,
		Settings: model.ManaSettings{MaxMana: 100, StaffCastCost: 10, RegenAmount: 3, RegenIntervalTicks: 2}}
	s.Update()
	if mana.GetCurrentMana() != 98 {
		t.Fatal("regeneration ignored interval")
	}
	tick.tick = 2
	s.Update()
	if mana.GetCurrentMana() != 100 {
		t.Fatal("regeneration did not cap at maximum")
	}
	mana.Spend(80)
	health.SetCurrentHealth(0)
	tick.tick = 4
	s.Update()
	if mana.GetCurrentMana() != 20 {
		t.Fatal("dead entity regenerated mana")
	}
	(&HealthSystem{SystemBase: s.SystemBase, TickSource: tick}).Update()
	if mana.GetCurrentMana() != 100 || health.GetCurrentHealth() != 100 {
		t.Fatal("respawn did not refill vitals")
	}
	if manager.GetEntityComponent(component.ComponentIdMana, id) != mana {
		t.Fatal("respawn lost mana component")
	}
}
