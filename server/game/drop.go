package game

import (
	"webscape/server/game/component"
	"webscape/server/game/gameevent"
	"webscape/server/game/model"
	"webscape/server/util"
)

func (g *Game) HandleDrop(clientID string, itemID model.ItemId) {
	g.stateMutex.Lock()
	defer g.stateMutex.Unlock()

	playerID, ok := g.clientIdToEntityId.Get(clientID)
	if !ok {
		return
	}
	inventoryValue := g.componentManager.GetEntityComponent(component.ComponentIdInventory, playerID)
	if inventoryValue == nil || g.componentManager.GetEntityComponent(component.ComponentIdPosition, playerID) == nil {
		return
	}
	inventory := inventoryValue.(*component.CInventory)
	item := inventory.GetItem(itemID)
	if item == nil {
		return
	}

	position := g.componentManager.GetEntityComponent(component.ComponentIdPosition, playerID).(*component.CPosition).GetPosition()
	if !inventory.RemoveItem(itemID) {
		return
	}
	g.componentManager.SetEntityComponent(playerID, inventory)
	g.componentManager.CreateNewEntity(
		component.NewCPosition(position),
		component.NewCRenderable("droppeditem"),
		component.NewCMetadata(util.JObject{
			"name":           util.JString(item.Name),
			"renderModel":    util.JString(item.GroundRenderModel()),
			"width":          util.JNumber(1),
			"height":         util.JNumber(1),
			"blocksMovement": util.JBool(false),
		}),
		&component.CDroppedItem{Item: item},
	)
}

func (g *Game) pickUpDroppedItem(playerID, targetID model.EntityId, drop *component.CDroppedItem) {
	playerPosition := g.componentManager.GetEntityComponent(component.ComponentIdPosition, playerID)
	dropPosition := g.componentManager.GetEntityComponent(component.ComponentIdPosition, targetID)
	if playerPosition == nil || dropPosition == nil ||
		playerPosition.(*component.CPosition).GetPosition() != dropPosition.(*component.CPosition).GetPosition() {
		return
	}
	// This is a transfer, not a new collection: repeated drop/pickup must not
	// generate collection events or advance collection quests.
	if g.addItemToPlayerInventory(playerID, drop.Item, false) {
		g.componentManager.RemoveEntity(targetID)
		g.EmitGameEvent(gameevent.NewItemPickedUp(playerID, targetID))
	}
}
