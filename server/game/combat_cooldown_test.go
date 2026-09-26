package game

import (
	"testing"
	"webscape/server/game/component"
	"webscape/server/game/gameevent"
	"webscape/server/game/model"
	"webscape/server/math"
)

func TestAttackCooldownSurvivesNewIntentAndWeaponChanges(t *testing.T) {
	for _, action := range []string{"repeat", "retarget", "move", "equip", "unequip"} {
		t.Run(action, func(t *testing.T) {
			g, player, inventory := newDropTestGame(t)
			if action == "unequip" {
				for _, item := range inventory.GetAllItems() {
					if item.Name() == "Iron Sword" {
						g.HandleEquip("player", item.Id)
						break
					}
				}
				equipped := g.componentManager.GetEntityComponent(component.ComponentIdEquipped, player).(*component.CEquipped)
				if equipped.GetEquippedItem(model.SlotWeapon) == nil {
					t.Fatal("unequip test requires an equipped weapon")
				}
			}
			g.componentManager.SetEntityComponent(player, component.NewCPosition(math.Vec2{}))
			g.componentManager.SetEntityComponent(player, component.NewCCombatStats(1, 1, 0, 0, 0, 0, 1.5, 1, 4))
			target := g.componentManager.CreateNewEntity(
				component.NewCPosition(math.Vec2{X: 1}), component.NewCHealth(100, 100),
				component.NewCCombatStats(1, 1, 0, 0, 0, 0, 1.5, 1, 4),
			)
			other := g.componentManager.CreateNewEntity(
				component.NewCPosition(math.Vec2{Y: 1}), component.NewCHealth(100, 100),
				component.NewCCombatStats(1, 1, 0, 0, 0, 0, 1.5, 1, 4),
			)
			attacks := 0
			g.RegisterGameEventHandlerFor(gameevent.EventIdCombatResolved, gameevent.HandlerFunc(func(event gameevent.Event) {
				if event.ActorEntityId == player {
					attacks++
				}
			}))
			g.HandleInteract("player", target, component.InteractionOptionAttack)
			g.update()
			if attacks != 1 {
				t.Fatalf("initial attacks = %d, want 1", attacks)
			}
			active := g.componentManager.GetEntityComponent(component.ComponentIdCombatState, player)
			switch action {
			case "retarget":
				target = other
			case "move":
				g.HandleMove("player", 0, 0)
			case "equip":
				var sword *model.Item
				for _, item := range inventory.GetAllItems() {
					if item.Name() == "Iron Sword" {
						sword = item
						break
					}
				}
				if sword == nil {
					t.Fatal("missing starter sword")
				}
				g.HandleEquip("player", sword.Id)
			case "unequip":
				g.HandleUnequip("player", model.SlotWeapon)
			}
			for tick := 2; tick <= 5; tick++ {
				g.HandleInteract("player", target, component.InteractionOptionAttack)
				if action == "repeat" && g.componentManager.GetEntityComponent(component.ComponentIdCombatState, player) != active {
					t.Fatal("repeated attack replaced the active combat state")
				}
				g.update()
				want := 1
				if tick == 5 {
					want = 2
				}
				if attacks != want {
					t.Fatalf("tick %d: attacks = %d, want %d", tick, attacks, want)
				}
			}
		})
	}
}

func TestRepeatedAttackDoesNotInterruptMagicWindup(t *testing.T) {
	g, player, inventory := newDropTestGame(t)
	g.componentManager.SetEntityComponent(player, component.NewCPosition(math.Vec2{}))
	var staff *model.Item
	for _, item := range inventory.GetAllItems() {
		if item.Name() == "Magic Staff" {
			staff = item
			break
		}
	}
	if staff == nil {
		t.Fatal("missing starter staff")
	}
	g.HandleEquip("player", staff.Id)
	target := g.componentManager.CreateNewEntity(
		component.NewCPosition(math.Vec2{X: 1}), component.NewCHealth(100, 100),
		component.NewCCombatStats(1, 1, 0, 0, 0, 0, 1.5, 1, 4),
	)
	launches := 0
	g.RegisterGameEventHandlerFor(gameevent.EventIdCombatProjectileLaunched, gameevent.HandlerFunc(func(event gameevent.Event) {
		if event.ActorEntityId == player {
			launches++
		}
	}))
	for tick := 1; tick <= 3; tick++ {
		g.HandleInteract("player", target, component.InteractionOptionAttack)
		g.update()
	}
	if launches != 1 {
		t.Fatalf("launches after repeated requests during windup = %d, want 1", launches)
	}
}
