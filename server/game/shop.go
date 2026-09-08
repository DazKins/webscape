package game

import (
	"webscape/server/game/component"
	"webscape/server/game/gameevent"
	"webscape/server/game/model"
	"webscape/server/message"
)

// CanTradeWith is used at entry, on every transaction, and each tick.
func (g *Game) CanTradeWith(player, target model.EntityId) bool {
	client, ok := g.clientIdToEntityId.GetKey(player)
	if !ok || !g.entityVisibleToClient(client, target) || player == target {
		return false
	}
	if g.componentManager.GetEntityComponent(component.ComponentIdShop, target) == nil {
		return false
	}
	for _, id := range []model.EntityId{player, target} {
		if health := g.componentManager.GetEntityComponent(component.ComponentIdHealth, id); health != nil && health.(*component.CHealth).GetCurrentHealth() <= 0 {
			return false
		}
		if g.componentManager.GetEntityComponent(component.ComponentIdCombatState, id) != nil {
			return false
		}
	}
	a := g.componentManager.GetEntityComponent(component.ComponentIdPosition, player)
	b := g.componentManager.GetEntityComponent(component.ComponentIdPosition, target)
	if a == nil || b == nil {
		return false
	}
	pa, pb := a.(*component.CPosition).GetPosition(), b.(*component.CPosition).GetPosition()
	dx, dy := pa.X-pb.X, pa.Y-pb.Y
	if dx < 0 {
		dx = -dx
	}
	if dy < 0 {
		dy = -dy
	}
	return dx+dy <= 1
}

func (g *Game) StartTradingFor(player, target model.EntityId) bool {
	if !g.CanTradeWith(player, target) {
		return false
	}
	g.stateTransitions.BeginSocialInteraction(player, &component.CTrading{TargetEntityId: target})
	return true
}

func (g *Game) HandleTradeClose(client string, target model.EntityId) {
	g.stateMutex.Lock()
	defer g.stateMutex.Unlock()
	player, ok := g.clientIdToEntityId.Get(client)
	if !ok {
		return
	}
	active := g.componentManager.GetEntityComponent(component.ComponentIdTrading, player)
	if active != nil && active.(*component.CTrading).TargetEntityId == target {
		g.componentManager.RemoveComponent(component.ComponentIdTrading, player)
	}
}

// A request carries ids and intent only. All price/quantity decisions use server state.
func (g *Game) HandleTrade(client string, target model.EntityId, action, itemID string) {
	g.stateMutex.Lock()
	defer g.stateMutex.Unlock()
	player, ok := g.clientIdToEntityId.Get(client)
	if !ok {
		return
	}
	active := g.componentManager.GetEntityComponent(component.ComponentIdTrading, player)
	if active == nil || active.(*component.CTrading).TargetEntityId != target || !g.CanTradeWith(player, target) {
		g.tradeResult(player, target, false, "Move next to the shopkeeper and choose Trade.")
		return
	}
	rawInventory := g.componentManager.GetEntityComponent(component.ComponentIdInventory, player)
	if rawInventory == nil {
		return
	}
	inventory := rawInventory.(*component.CInventory)
	shop := g.componentManager.GetEntityComponent(component.ComponentIdShop, target).(*component.CShop)
	for _, offer := range shop.Offers {
		if action == "buy" && offer.ItemId == itemID {
			gold := inventory.FindByType(model.ItemTypeGold)
			if gold == nil || gold.Quantity < offer.BuyPrice {
				g.tradeResult(player, target, false, "Not enough gold.")
				return
			}
			if !inventory.Exchange(gold.Id, offer.BuyPrice, model.CreateShopItem(offer.ItemId)) {
				g.tradeResult(player, target, false, "Your backpack is full.")
				return
			}
			g.tradeResult(player, target, true, "Bought "+offer.Item.Name+".")
			return
		}
		if action == "sell" {
			for _, item := range inventory.GetAllItems() {
				if item.Id.String() != itemID || !model.SameItemKind(item, offer.Item) {
					continue
				}
				if !inventory.Exchange(item.Id, 1, model.CreateGold(offer.SellPrice)) {
					g.tradeResult(player, target, false, "There is no room for more gold.")
					return
				}
				g.tradeResult(player, target, true, "Sold "+item.Name+".")
				return
			}
		}
	}
	g.tradeResult(player, target, false, "This shop does not trade that item.")
}

func (g *Game) tradeResult(player, target model.EntityId, success bool, text string) {
	event := gameevent.New(gameevent.EventIdTradeResolved, player)
	event.TargetEntityId = target
	event.Payload = gameevent.TradeResolvedPayload{Success: success, Text: text}
	g.EmitGameEvent(event)
}

func (g *Game) projectTradeResult(event gameevent.Event) {
	payload, ok := event.Payload.(gameevent.TradeResolvedPayload)
	if !ok {
		return
	}
	client, ok := g.clientIdToEntityId.GetKey(event.ActorEntityId)
	if !ok || !g.entityVisibleToClient(client, event.ActorEntityId) {
		return
	}
	g.pendingClientEvents = append(g.pendingClientEvents, pendingClientEvent{
		clientIDs: []string{client}, message: message.NewTradeResultMessage(event.TargetEntityId, payload.Success, payload.Text),
	})
}
