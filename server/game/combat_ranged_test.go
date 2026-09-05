package game

import (
	"testing"
	"webscape/server/game/component"
	"webscape/server/game/model"
	"webscape/server/message"
)

func TestEquippingWoodenBowRestartsCombatWithRangedProfile(t *testing.T) {
	game := NewGameWithWorld(loadInterestWorld(t))
	game.RegisterSender(func(string, message.Message) {})
	playerId := model.NewEntityId()
	targetId := model.NewEntityId()
	game.HandleRegister("player", playerId, "Player")
	game.HandleRegister("target", targetId, "Target")

	combatState := component.NewCCombatState(targetId)
	combatState.BeginCasting(0, 3)
	game.componentManager.SetEntityComponent(playerId, combatState)
	inventory := game.componentManager.GetEntityComponent(component.ComponentIdInventory, playerId).(*component.CInventory)
	var bow *model.Item
	for _, item := range inventory.GetAllItems() {
		if item.Name == "Wooden Bow" {
			bow = item
			break
		}
	}
	if bow == nil {
		t.Fatal("starter inventory has no Wooden Bow")
	}

	game.HandleEquip("player", bow.Id)

	combatState = game.componentManager.GetEntityComponent(component.ComponentIdCombatState, playerId).(*component.CCombatState)
	if combatState.GetTargetId() != targetId || combatState.GetPhase() != component.CombatPhaseApproaching ||
		combatState.GetAttackMethod() != model.AttackMethodRanged || combatState.GetWindUpTicks() != 2 {
		t.Fatalf("combat state after bow equip = target %s phase %s method %s windup %d",
			combatState.GetTargetId(), combatState.GetPhase(), combatState.GetAttackMethod(), combatState.GetWindUpTicks())
	}
}
