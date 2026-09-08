package system

import (
	"testing"
	"webscape/server/game/component"
	"webscape/server/game/model"
	"webscape/server/math"
)

func TestSocialParticipantsShareFacingAndPauseWandering(t *testing.T) {
	for _, trading := range []bool{false, true} {
		manager := component.NewComponentManager()
		player, target := model.NewEntityId(), model.NewEntityId()
		active := component.Component(component.NewCActiveConversation("greeting", target, "start"))
		if trading {
			active = &component.CTrading{TargetEntityId: target}
		}
		manager.SetEntityComponents(player, component.NewCPosition(math.Vec2{}), active)
		manager.SetEntityComponents(target, component.NewCPosition(math.Vec2{X: 1}), component.NewCRandomWalk(0, 5))
		facing := FacingSystem{SystemBase: SystemBase{ComponentManager: manager}}
		facing.Update()
		assertFacingTarget(t, manager, player, target)
		assertFacingTarget(t, manager, target, player)
		// World is nil: any attempted random move would fail. Active participants must be skipped.
		walker := RandomWalkSystem{SystemBase: SystemBase{ComponentManager: manager}}
		walker.Update()
		if !isSocialParticipant(manager, target) {
			t.Fatal("NPC not held by social interaction")
		}
		manager.RemoveComponent(active.GetId(), player)
		facing.Update()
		assertNoFacing(t, manager, player)
		assertNoFacing(t, manager, target)
		if isSocialParticipant(manager, target) {
			t.Fatal("NPC remains held after interaction")
		}
	}
}

func TestTradingTransitionsCancelIncompatibleState(t *testing.T) {
	for _, action := range []string{"move", "interact", "combat", "death", "fish", "chop", "talk"} {
		t.Run(action, func(t *testing.T) {
			manager := component.NewComponentManager()
			player, target := model.NewEntityId(), model.NewEntityId()
			manager.SetEntityComponents(player, component.NewCLocomotion(component.LocomotionPhaseIdle, 0), component.NewCInteracting(target, component.InteractionOptionTrade), component.NewCPathing(component.PathingTarget{}))
			transitions := NewEntityStateTransitions(manager, nil)
			transitions.BeginSocialInteraction(player, &component.CTrading{TargetEntityId: target})
			if err := ValidateEntityState(manager, player); err != nil {
				t.Fatal(err)
			}
			switch action {
			case "move":
				transitions.BeginPathing(player, component.NewCPathing(component.PathingTarget{}))
			case "interact":
				transitions.BeginInteraction(player, component.NewCPathing(component.PathingTarget{}), component.NewCInteracting(target, component.InteractionOptionTalk))
			case "combat":
				transitions.BeginCombat(player, component.NewCCombatState(target))
			case "death":
				transitions.HandleDeath(player)
			case "fish":
				transitions.BeginFishing(player, component.NewCFishing(target, 0, math.Vec2{}))
			case "chop":
				transitions.BeginWoodcutting(player, component.NewCWoodcutting(target, 0, math.Vec2{}))
			case "talk":
				transitions.BeginSocialInteraction(player, component.NewCActiveConversation("greeting", target, "start"))
			}
			if manager.GetEntityComponent(component.ComponentIdTrading, player) != nil {
				t.Fatal("transition left trading active")
			}
		})
	}
}
