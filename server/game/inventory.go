package game

import (
	"webscape/server/game/component"
	"webscape/server/game/model"
)

func (g *Game) HandleInventoryMove(clientID string, itemID model.ItemId, slot int, sequence uint64) {
	g.stateMutex.Lock()
	defer g.stateMutex.Unlock()
	playerID, ok := g.clientIdToEntityId.Get(clientID)
	if !ok {
		return
	}
	state := g.clients[clientID]
	if state == nil || sequence == 0 || sequence > 9007199254740991 || sequence <= state.inventoryMoveSequence {
		return
	}
	// Acknowledge rejected moves too so the client can discard its prediction.
	state.inventoryMoveSequence = sequence
	value := g.componentManager.GetEntityComponent(component.ComponentIdInventory, playerID)
	if value == nil {
		return
	}
	inventory := value.(*component.CInventory)
	if inventory.MoveItem(itemID, slot) {
		g.componentManager.SetEntityComponent(playerID, inventory)
	}
}
