package game

import (
	"webscape/server/game/component"
	"webscape/server/game/gameevent"
	"webscape/server/game/model"
	"webscape/server/message"
)

func (g *Game) CanBankWith(player, target model.EntityId) bool {
	return g.canUseNPC(player, target, component.ComponentIdBanker)
}
func (g *Game) StartBankingFor(player, target model.EntityId) bool {
	if !g.CanBankWith(player, target) {
		return false
	}
	g.stateTransitions.BeginSocialInteraction(player, &component.CBanking{TargetEntityId: target})
	return true
}

func (g *Game) HandleBankClose(client string, target model.EntityId) {
	g.stateMutex.Lock()
	defer g.stateMutex.Unlock()
	player, ok := g.clientIdToEntityId.Get(client)
	if !ok {
		return
	}
	active := g.componentManager.GetEntityComponent(component.ComponentIdBanking, player)
	if active != nil && active.(*component.CBanking).TargetEntityId == target {
		g.componentManager.RemoveComponent(component.ComponentIdBanking, player)
	}
}

func (g *Game) HandleBankTransfer(client string, target model.EntityId, deposit bool, itemID model.ItemId, quantity int) {
	g.stateMutex.Lock()
	defer g.stateMutex.Unlock()
	player, ok := g.clientIdToEntityId.Get(client)
	if !ok {
		return
	}
	active := g.componentManager.GetEntityComponent(component.ComponentIdBanking, player)
	if active == nil || active.(*component.CBanking).TargetEntityId != target || !g.CanBankWith(player, target) {
		g.bankResult(player, target, false, "Move next to a banker and choose Bank.")
		return
	}
	bank, bankOK := g.componentManager.GetEntityComponent(component.ComponentIdBank, player).(*component.CBank)
	inventory, inventoryOK := g.componentManager.GetEntityComponent(component.ComponentIdInventory, player).(*component.CInventory)
	if !bankOK || !inventoryOK {
		return
	}
	item := bank.GetItem(itemID)
	if deposit {
		item = inventory.GetItem(itemID)
	}
	if item == nil || quantity <= 0 || quantity > item.Quantity {
		g.bankResult(player, target, false, "That item or quantity is no longer available.")
		return
	}
	if !bank.Transfer(inventory, deposit, itemID, quantity) {
		text := "Your backpack is full or that stack has reached its limit."
		if deposit {
			text = "That bank stack has reached its limit."
		}
		g.bankResult(player, target, false, text)
		return
	}
	text := "Withdrawn to your backpack."
	if deposit {
		text = "Deposited in your bank."
	}
	g.bankResult(player, target, true, text)
}
func (g *Game) bankResult(player, target model.EntityId, success bool, text string) {
	event := gameevent.New(gameevent.EventIdBankResolved, player)
	event.TargetEntityId = target
	event.Payload = gameevent.BankResolvedPayload{Success: success, Text: text}
	g.EmitGameEvent(event)
}

func (g *Game) projectBankResult(event gameevent.Event) {
	payload, ok := event.Payload.(gameevent.BankResolvedPayload)
	if !ok {
		return
	}
	client, ok := g.clientIdToEntityId.GetKey(event.ActorEntityId)
	if !ok || !g.entityVisibleToClient(client, event.ActorEntityId) {
		return
	}
	g.pendingClientEvents = append(g.pendingClientEvents, pendingClientEvent{
		clientIDs: []string{client}, message: message.NewBankResultMessage(event.TargetEntityId, payload.Success, payload.Text),
	})
}
