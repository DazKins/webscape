package game

import (
	"encoding/json"
	"strings"
	"testing"
	"webscape/server/game/component"
	"webscape/server/game/gameevent"
	"webscape/server/game/model"
	"webscape/server/message"
	"webscape/server/util"
)

func TestGroundPickupRequiresStandingOnDroppedTile(t *testing.T) {
	g, playerID, inventory := newDropTestGame(t)
	movePlayerForTest(g, playerID, 1, 1)
	item := inventory.GetAllItems()[0]
	g.HandleDrop("player", item.Id)
	dropID, _ := firstEntityWithComponent(g, component.ComponentIdDroppedItem)
	dropPosition := g.componentManager.GetEntityComponent(component.ComponentIdPosition, dropID).(*component.CPosition).GetPosition()
	if dropPosition.X != 1 || dropPosition.Y != 1 {
		t.Fatalf("drop position = %v, want player's tile (1,1)", dropPosition)
	}
	movePlayerForTest(g, playerID, 0, 1)
	g.LootEntityFor(playerID, dropID)
	if inventory.HasItem(item.Id) || !g.componentManager.HasEntity(dropID) {
		t.Fatal("pickup succeeded from adjacent tile")
	}
	g.HandleInteract("player", dropID, component.InteractionOptionLoot)
	g.update()
	position := g.componentManager.GetEntityComponent(component.ComponentIdPosition, playerID).(*component.CPosition).GetPosition()
	if position != dropPosition || !g.componentManager.HasEntity(dropID) {
		t.Fatal("pickup should walk onto the item and settle before collecting it")
	}
	g.update()
	if !inventory.HasItem(item.Id) || g.componentManager.HasEntity(dropID) {
		t.Fatal("pickup did not complete on the item's tile")
	}
	if g.componentManager.GetEntityComponent(component.ComponentIdPosition, playerID).(*component.CPosition).GetPosition() != dropPosition {
		t.Fatal("pathing pushed the player off the item's tile")
	}
}

func TestPickupEventIsOnceVisibleAndAfterState(t *testing.T) {
	g, playerID, inventory := newDropTestGame(t)
	g.HandleRegister("viewer", model.NewEntityId(), "Viewer")
	farID := model.NewEntityId()
	g.HandleRegister("far", farID, "Far")
	movePlayerForTest(g, farID, 4, 0)
	g.syncClient("far")
	item := inventory.GetAllItems()[0]
	g.HandleDrop("player", item.Id)
	dropID, _ := firstEntityWithComponent(g, component.ComponentIdDroppedItem)
	g.update()
	sent := map[string][]message.Message{}
	g.RegisterSender(func(clientID string, msg message.Message) { sent[clientID] = append(sent[clientID], msg) })
	g.HandleInteract("player", dropID, component.InteractionOptionLoot)
	g.update()
	for _, clientID := range []string{"player", "viewer"} {
		types := messageTypes(sent[clientID])
		stateIndex := indexOfMessageType(types, message.MessageTypeGameUpdate)
		eventIndex := indexOfMessageType(types, message.MessageTypeItemPickedUp)
		if stateIndex < 0 || eventIndex <= stateIndex {
			t.Fatalf("%s message order = %v, want state then pickup", clientID, types)
		}
		var wire struct {
			Data map[string]any `json:"data"`
		}
		if err := json.Unmarshal([]byte(sent[clientID][eventIndex].Marshal()), &wire); err != nil {
			t.Fatal(err)
		}
		if len(wire.Data) != 1 || wire.Data["playerEntityId"] != playerID.String() {
			t.Fatalf("unsafe or invalid pickup DTO: %#v", wire.Data)
		}
	}
	if indexOfMessageType(messageTypes(sent["far"]), message.MessageTypeItemPickedUp) >= 0 {
		t.Fatal("pickup event leaked outside interest")
	}
	sent = map[string][]message.Message{}
	g.HandleInteract("player", dropID, component.InteractionOptionLoot)
	g.HandleRegister("late", model.NewEntityId(), "Late")
	g.update()
	for clientID, messages := range sent {
		if indexOfMessageType(messageTypes(messages), message.MessageTypeItemPickedUp) >= 0 {
			t.Fatalf("pickup replayed to %s", clientID)
		}
	}
}

func newDropTestGame(t *testing.T) (*Game, model.EntityId, *component.CInventory) {
	t.Helper()
	g := NewGameWithWorldAndChunkRadius(loadInterestWorld(t), 0)
	g.RegisterSender(func(string, message.Message) {})
	playerID := model.NewEntityId()
	g.HandleRegister("player", playerID, "Player")
	inventory := g.componentManager.GetEntityComponent(component.ComponentIdInventory, playerID).(*component.CInventory)
	return g, playerID, inventory
}

func TestDropTransfersOriginalItemAndReplicatesGroundModel(t *testing.T) {
	g, _, inventory := newDropTestGame(t)
	item := inventory.GetAllItems()[0]
	before := component.SerializeItem(item)
	collected := 0
	g.RegisterGameEventHandler(gameevent.HandlerFunc(func(event gameevent.Event) {
		if strings.HasPrefix(event.Id, "collect:") {
			collected++
		}
	}))
	sent := []message.Message{}
	g.RegisterSender(func(client string, msg message.Message) {
		if client == "player" {
			sent = append(sent, msg)
		}
	})
	g.HandleDrop("player", item.Id)
	g.update()
	dropID, ok := firstEntityWithComponent(g, component.ComponentIdDroppedItem)
	if !ok || inventory.HasItem(item.Id) {
		t.Fatal("drop did not transfer item out of inventory")
	}
	metadata := g.clients["player"].baseline[component.ComponentIdMetadata][dropID].(util.JObject)
	if metadata["name"] != util.JString(item.Name) || metadata["renderModel"] != util.JString("ironSword") || metadata["blocksMovement"] != util.JBool(false) {
		t.Fatalf("ground presentation = %#v", metadata)
	}
	if !gameUpdateIncludesInteraction(sent, dropID.String(), "loot") {
		t.Fatal("drop has no replicated pickup interaction")
	}
	if _, leaked := g.clients["player"].baseline[component.ComponentIdDroppedItem][dropID]; leaked {
		t.Fatal("internal item component leaked onto wire")
	}
	g.HandleDrop("player", item.Id)
	if len(g.componentManager.GetEntitiesWithComponents(component.ComponentIdDroppedItem)) != 1 {
		t.Fatal("repeated drop duplicated item")
	}

	// Late joiners reconstruct the current drop from ordinary state replication.
	otherID := model.NewEntityId()
	g.HandleRegister("other", otherID, "Other")
	if !baselineKnows(g.clients["other"], dropID) {
		t.Fatal("late joiner did not receive dropped item")
	}
	movePlayerForTest(g, otherID, 4, 0)
	g.syncClient("other")
	if baselineKnows(g.clients["other"], dropID) {
		t.Fatal("drop remained visible outside interest")
	}
	g.HandleInteract("other", dropID, component.InteractionOptionLoot)
	if g.componentManager.GetEntityComponent(component.ComponentIdInteracting, otherID) != nil {
		t.Fatal("accepted pickup outside interest")
	}
	movePlayerForTest(g, otherID, 0, 0)
	g.syncClient("other")
	if !baselineKnows(g.clients["other"], dropID) {
		t.Fatal("interest re-entry did not restore ground item")
	}
	known := knownComponentCount(g.clients["player"], dropID)
	sent = nil
	g.HandleInteract("other", dropID, component.InteractionOptionLoot)
	g.update()
	otherInventory := g.componentManager.GetEntityComponent(component.ComponentIdInventory, otherID).(*component.CInventory)
	got := otherInventory.GetItem(item.Id)
	if got != item || !util.JsonEqual(component.SerializeItem(got), before) {
		t.Fatal("pickup failed to preserve original item and equipment stats")
	}
	if g.componentManager.HasEntity(dropID) || tombstoneCount(t, sent, dropID.String()) != known {
		t.Fatal("pickup did not remove the ground entity from server and client")
	}
	g.HandleInteract("player", dropID, component.InteractionOptionLoot)
	g.update()
	if inventory.HasItem(item.Id) || collected != 0 {
		t.Fatal("pickup duplicated item or emitted collection quest events")
	}
	// Recovered equipment must remain usable.
	g.HandleEquip("other", item.Id)
	equipped := g.componentManager.GetEntityComponent(component.ComponentIdEquipped, otherID).(*component.CEquipped)
	if equipped.GetEquippedItem(model.SlotWeapon) != item {
		t.Fatal("recovered sword cannot be equipped")
	}
}

func TestDropRejectsUnownedAndEquippedItems(t *testing.T) {
	g, playerID, inventory := newDropTestGame(t)
	item := inventory.GetAllItems()[0]
	for _, request := range []struct {
		client string
		itemID model.ItemId
	}{
		{"unknown", item.Id}, {"player", model.NewItemId()},
	} {
		g.HandleDrop(request.client, request.itemID)
	}
	otherID := model.NewEntityId()
	g.HandleRegister("other", otherID, "Other")
	g.HandleDrop("other", item.Id)
	if !inventory.HasItem(item.Id) {
		t.Fatal("invalid client or item removed inventory item")
	}
	g.HandleEquip("player", item.Id)
	g.HandleDrop("player", item.Id)
	equipped := g.componentManager.GetEntityComponent(component.ComponentIdEquipped, playerID).(*component.CEquipped)
	if equipped.GetEquippedItem(model.SlotWeapon) != item || len(g.componentManager.GetEntitiesWithComponents(component.ComponentIdDroppedItem)) != 0 {
		t.Fatal("drop accepted an unowned or equipped item")
	}
}

func TestDropPickupFullInventoryAndCompetingPlayers(t *testing.T) {
	g, _, inventory := newDropTestGame(t)
	pickups := 0
	g.RegisterGameEventHandlerFor(gameevent.EventIdItemPickedUp, gameevent.HandlerFunc(func(event gameevent.Event) {
		if _, ok := event.Payload.(gameevent.ItemPickedUpPayload); !ok {
			t.Fatal("pickup payload has wrong type")
		}
		pickups++
	}))
	item := inventory.GetAllItems()[0]
	g.HandleDrop("player", item.Id)
	dropID, _ := firstEntityWithComponent(g, component.ComponentIdDroppedItem)
	for !inventory.IsFull() {
		inventory.AddItem(model.CreateArrow())
	}
	g.HandleInteract("player", dropID, component.InteractionOptionLoot)
	g.update()
	if !g.componentManager.HasEntity(dropID) || inventory.HasItem(item.Id) || pickups != 0 {
		t.Fatal("full inventory lost the ground item or exceeded capacity")
	}
	inventory.RemoveItem(inventory.GetAllItems()[0].Id)
	otherID := model.NewEntityId()
	g.HandleRegister("other", otherID, "Other")
	// Both commands are queued before the same tick. Only one transfer may win.
	g.HandleInteract("player", dropID, component.InteractionOptionLoot)
	g.HandleInteract("other", dropID, component.InteractionOptionLoot)
	g.update()
	otherInventory := g.componentManager.GetEntityComponent(component.ComponentIdInventory, otherID).(*component.CInventory)
	if inventory.HasItem(item.Id) == otherInventory.HasItem(item.Id) || g.componentManager.HasEntity(dropID) || pickups != 1 {
		t.Fatal("competing pickups must transfer the item to exactly one player")
	}
}
