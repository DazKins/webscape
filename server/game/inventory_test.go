package game

import (
	"encoding/json"
	"testing"
	"webscape/server/game/component"
	"webscape/server/game/model"
	"webscape/server/message"
	"webscape/server/util"
)

func TestInventoryMoveUsesConnectionOwnership(t *testing.T) {
	g, _, inventory := newDropTestGame(t)
	g.HandleRegister("other", model.NewEntityId(), "Other")
	item := inventory.GetAllItems()[0]
	before := inventory.Serialize()
	g.HandleInventoryMove("other", item.Id, 19, 1)
	g.HandleInventoryMove("unregistered", item.Id, 19, 1)
	g.HandleInventoryMove("player", item.Id, component.InventoryCapacity, 1)
	if !util.JsonEqual(before, inventory.Serialize()) {
		t.Fatal("unauthorized move changed inventory")
	}
	g.HandleInventoryMove("player", item.Id, 19, 2)
	wire := inventory.Serialize().(util.JObject)
	if wire["items"].(util.JArray)[0].(util.JObject)["slot"] != util.JNumber(19) {
		t.Fatal("position not replicated")
	}
}

func TestInventoryMoveAcknowledgementsFollowStateAndAreConnectionScoped(t *testing.T) {
	g, _, inventory := newDropTestGame(t)
	g.HandleRegister("other", model.NewEntityId(), "Other")
	item := inventory.GetAllItems()[0]
	g.HandleInventoryMove("player", item.Id, 19, 1)
	// A rejected request still completes, even when no components change.
	g.HandleInventoryMove("player", model.NewItemId(), 0, 2)
	g.HandleInventoryMove("player", item.Id, 18, 1) // stale duplicate
	if inventory.Serialize().(util.JObject)["items"].(util.JArray)[0].(util.JObject)["slot"] != util.JNumber(19) {
		t.Fatal("stale request moved item")
	}
	var updates = map[string]map[string]any{}
	g.RegisterSender(func(client string, msg message.Message) {
		if msg.Metadata.Type != message.MessageTypeGameUpdate {
			return
		}
		raw, _ := json.Marshal(msg.Data)
		var data map[string]any
		if err := json.Unmarshal(raw, &data); err != nil {
			t.Fatal(err)
		}
		updates[client] = data
	})
	g.syncClient("player")
	g.syncClient("other")
	if updates["player"]["inventoryMoveSequence"] != float64(2) || updates["other"]["inventoryMoveSequence"] != float64(0) {
		t.Fatal("acknowledgement leaked or was omitted")
	}
	g.syncClient("player")
	if len(updates["player"]["entities"].([]any)) != 0 || updates["player"]["inventoryMoveSequence"] != float64(2) {
		t.Fatal("unchanged tick lost acknowledgement")
	}
	g.HandleLeave("player")
	g.HandleRegister("player", model.NewEntityId(), "New session")
	if g.clients["player"].inventoryMoveSequence != 0 {
		t.Fatal("new connection inherited acknowledgement")
	}
}
