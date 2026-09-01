package game

import (
	"testing"
	"webscape/server/game/component"
	"webscape/server/game/model"
	"webscape/server/message"
)

func TestEquippingMagicStaffRestartsCombatWithoutLosingTarget(t *testing.T) {
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
	var staff *model.Item
	for _, item := range inventory.GetAllItems() {
		if item.Name == "Magic Staff" {
			staff = item
			break
		}
	}
	if staff == nil {
		t.Fatal("starter inventory has no Magic Staff")
	}

	game.HandleEquip("player", staff.Id)

	combatState = game.componentManager.GetEntityComponent(component.ComponentIdCombatState, playerId).(*component.CCombatState)
	if combatState.GetTargetId() != targetId || combatState.GetPhase() != component.CombatPhaseApproaching ||
		combatState.GetAttackMethod() != model.AttackMethodMagic || combatState.GetWindUpTicks() != 2 {
		t.Fatalf("combat state after staff equip = target %s phase %s method %s windup %d",
			combatState.GetTargetId(), combatState.GetPhase(), combatState.GetAttackMethod(), combatState.GetWindUpTicks())
	}
}
