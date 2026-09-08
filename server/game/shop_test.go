package game

import (
	"encoding/json"
	"testing"
	"webscape/server/game/component"
	"webscape/server/game/gameevent"
	"webscape/server/game/model"
	"webscape/server/game/system"
	"webscape/server/math"
	"webscape/server/message"
	"webscape/server/util"
)

func newShopTestGame(t *testing.T) (*Game, model.EntityId, model.EntityId, *component.CInventory) {
	t.Helper()
	g, player, inventory := newDropTestGame(t)
	shop, err := component.ParseShop(map[string]any{"offers": []any{
		map[string]any{"itemId": "ironSword", "buyPrice": float64(40), "sellPrice": float64(16)},
		map[string]any{"itemId": "logs", "buyPrice": float64(5), "sellPrice": float64(2)},
	}})
	if err != nil {
		t.Fatal(err)
	}
	target := g.componentManager.CreateNewEntity(component.NewCPosition(math.Vec2{X: 1, Y: 0}), shop)
	g.update()
	return g, player, target, inventory
}

func TestTradeLifecycleFacingAndReplication(t *testing.T) {
	g, player, target, _ := newShopTestGame(t)
	viewer := model.NewEntityId()
	g.HandleRegister("viewer", viewer, "Viewer")
	sent := map[string][]message.Message{}
	g.RegisterSender(func(client string, msg message.Message) { sent[client] = append(sent[client], msg) })
	g.HandleInteract("player", target, component.InteractionOptionTrade)
	g.update()
	if err := system.ValidateEntityState(g.componentManager, player); err != nil {
		t.Fatal(err)
	}
	active := g.componentManager.GetEntityComponent(component.ComponentIdTrading, player)
	if active == nil || active.(*component.CTrading).TargetEntityId != target {
		t.Fatal("trade interaction did not establish state")
	}
	for a, b := range map[model.EntityId]model.EntityId{player: target, target: player} {
		facing := g.componentManager.GetEntityComponent(component.ComponentIdFacing, a)
		if facing == nil || facing.(*component.CFacing).GetTargetEntityId() != b {
			t.Fatal("participants do not face each other")
		}
	}
	for _, client := range []string{"player", "viewer"} {
		found := false
		for _, msg := range sent[client] {
			if msg.Metadata.Type != message.MessageTypeGameUpdate {
				continue
			}
			var wire struct {
				Data struct {
					Entities []struct {
						ComponentId string `json:"componentId"`
						EntityId    string `json:"entityId"`
						Data        any    `json:"data"`
					} `json:"entities"`
				} `json:"data"`
			}
			if err := json.Unmarshal([]byte(msg.Marshal()), &wire); err != nil {
				t.Fatal(err)
			}
			for _, delta := range wire.Data.Entities {
				if delta.ComponentId == "trading" && delta.EntityId == player.String() {
					found = true
				}
			}
		}
		if found != (client == "player") {
			t.Fatalf("private trading state for %s: %v", client, found)
		}
	}
	g.HandleTradeClose("player", model.NewEntityId())
	if g.componentManager.GetEntityComponent(component.ComponentIdTrading, player) == nil {
		t.Fatal("stale close cancelled another shop")
	}
	sent = map[string][]message.Message{}
	g.HandleMove("player", 0, 1)
	g.update()
	if g.componentManager.GetEntityComponent(component.ComponentIdTrading, player) != nil || g.componentManager.GetEntityComponent(component.ComponentIdFacing, target) != nil {
		t.Fatal("walking did not end trading and facing")
	}
	removed := false
	for _, msg := range sent["player"] {
		if msg.Metadata.Type != message.MessageTypeGameUpdate {
			continue
		}
		var wire struct {
			Data struct {
				Entities []struct {
					ComponentId string `json:"componentId"`
					Data        any    `json:"data"`
				} `json:"entities"`
			} `json:"data"`
		}
		json.Unmarshal([]byte(msg.Marshal()), &wire)
		for _, delta := range wire.Data.Entities {
			if delta.ComponentId == "trading" && delta.Data == nil {
				removed = true
			}
		}
	}
	if !removed {
		t.Fatal("trading removal was not replicated")
	}
}

func TestBuySellAreAuthoritativeAndResultsFollowState(t *testing.T) {
	g, player, target, inventory := newShopTestGame(t)
	g.HandleRegister("viewer", model.NewEntityId(), "Viewer")
	if !g.StartTradingFor(player, target) {
		t.Fatal("cannot trade")
	}
	g.update()
	sent := map[string][]message.Message{}
	g.RegisterSender(func(client string, msg message.Message) { sent[client] = append(sent[client], msg) })
	g.HandleTrade("player", target, "buy", "ironSword")
	if inventory.FindByType(model.ItemTypeGold).Quantity != 60 {
		t.Fatal("buy did not charge authored price")
	}
	purchased := inventory.GetAllItems()[inventory.GetItemCount()-1]
	if !model.SameItemKind(purchased, model.CreateIronSword()) {
		t.Fatal("purchase lost weapon stats")
	}
	g.update()
	types := messageTypes(sent["player"])
	state, result := indexOfMessageType(types, message.MessageTypeGameUpdate), indexOfMessageType(types, message.MessageTypeTradeResult)
	if state < 0 || result <= state {
		t.Fatalf("order %v", types)
	}
	if indexOfMessageType(messageTypes(sent["viewer"]), message.MessageTypeTradeResult) >= 0 {
		t.Fatal("trade result leaked to viewer")
	}
	var dto map[string]any
	json.Unmarshal([]byte(sent["player"][result].Marshal()), &dto)
	if len(dto["data"].(map[string]any)) != 3 {
		t.Fatal("unexpected fields in trade DTO")
	}
	g.HandleTrade("player", target, "sell", purchased.Id.String())
	if inventory.HasItem(purchased.Id) || inventory.FindByType(model.ItemTypeGold).Quantity != 76 {
		t.Fatal("sale failed")
	}
	before := inventory.Serialize()
	g.HandleTrade("player", target, "sell", purchased.Id.String())
	g.HandleTrade("player", target, "buy", "unknown")
	g.HandleTrade("player", target, "sell", inventory.FindByType(model.ItemTypeGold).Id.String())
	g.HandleTrade("stranger", target, "buy", "logs")
	if !util.JsonEqual(before, inventory.Serialize()) {
		t.Fatal("invalid requests changed inventory")
	}
	g.HandleTradeClose("player", target)
	g.HandleTrade("player", target, "buy", "logs")
	if !util.JsonEqual(before, inventory.Serialize()) {
		t.Fatal("trade succeeded without active session")
	}
}

func TestTradeRejectsStaleTargetsAndEquippedItems(t *testing.T) {
	for _, reason := range []string{"distance", "visibility", "removed", "combat", "death"} {
		t.Run(reason, func(t *testing.T) {
			g, player, target, inventory := newShopTestGame(t)
			g.StartTradingFor(player, target)
			before := inventory.Serialize()
			switch reason {
			case "distance":
				movePlayerForTest(g, player, 0, 3)
			case "visibility":
				g.clients["player"].loadedChunks = nil
			case "removed":
				g.componentManager.RemoveEntity(target)
			case "combat":
				g.componentManager.SetEntityComponent(target, component.NewCCombatState(player))
			case "death":
				g.componentManager.SetEntityComponent(target, component.NewCHealth(100, 0))
			}
			g.HandleTrade("player", target, "buy", "logs")
			if !util.JsonEqual(before, inventory.Serialize()) {
				t.Fatal("stale trade mutated inventory")
			}
			validator := system.TradingSystem{SystemBase: system.SystemBase{ComponentManager: g.componentManager}, Validator: g}
			validator.Update()
			if g.componentManager.GetEntityComponent(component.ComponentIdTrading, player) != nil {
				t.Fatal("invalid trading state not cleaned up")
			}
		})
	}
	g, player, target, inventory := newShopTestGame(t)
	sword := inventory.GetAllItems()[0]
	g.HandleEquip("player", sword.Id)
	g.StartTradingFor(player, target)
	before := inventory.Serialize()
	g.HandleTrade("player", target, "sell", sword.Id.String())
	if !util.JsonEqual(before, inventory.Serialize()) {
		t.Fatal("sold equipped item")
	}
}

func TestGoldDropPickupAndAuthoredLootMergeWithoutExtraSlots(t *testing.T) {
	g, player, inventory := newDropTestGame(t)
	gold := inventory.FindByType(model.ItemTypeGold)
	g.HandleDrop("player", gold.Id)
	drop, _ := firstEntityWithComponent(g, component.ComponentIdDroppedItem)
	if inventory.FindByType(model.ItemTypeGold) != nil {
		t.Fatal("drop left coins behind")
	}
	inventory.AddItem(model.CreateGold(5))
	for !inventory.IsFull() {
		inventory.AddItem(model.CreateBread())
	}
	collected := 0
	g.RegisterGameEventHandlerFor("collect:item:gold", gameevent.HandlerFunc(func(e gameevent.Event) { collected += e.Count }))
	g.LootEntityFor(player, drop)
	if g.componentManager.HasEntity(drop) || inventory.FindByType(model.ItemTypeGold).Quantity != 105 || collected != 0 {
		t.Fatal("pickup did not transfer complete stack into full inventory")
	}
	loot := component.NewCLootable(true, []component.LootItem{{Name: "Gold", Type: "gold", Count: 50}})
	chest := g.componentManager.CreateNewEntity(loot)
	g.LootEntityFor(player, chest)
	if !loot.IsLooted() || inventory.FindByType(model.ItemTypeGold).Quantity != 155 || collected != 50 {
		t.Fatal("authored gold loot did not merge or emit quantity")
	}
}

func TestShopBaselineReconstructsStateWithoutReplayingTrades(t *testing.T) {
	g, player, target, _ := newShopTestGame(t)
	g.StartTradingFor(player, target)
	g.HandleTrade("player", target, "buy", "logs")
	g.update()
	sent := []message.Message{}
	g.RegisterSender(func(client string, msg message.Message) {
		if client == "player" {
			sent = append(sent, msg)
		}
	})
	g.syncClient("player")
	if len(sent) != 0 {
		t.Fatal("unchanged shop state produced a delta")
	}
	g.clients["player"].baseline = make(map[component.ComponentId]map[model.EntityId]util.Json)
	g.syncClient("player")
	if indexOfMessageType(messageTypes(sent), message.MessageTypeGameUpdate) < 0 || indexOfMessageType(messageTypes(sent), message.MessageTypeTradeResult) >= 0 {
		t.Fatal("baseline must reconstruct state without replaying trade results")
	}
}

func TestCompetingGoldPickupsDoNotDuplicateCoins(t *testing.T) {
	g, player, inventory := newDropTestGame(t)
	gold := inventory.FindByType(model.ItemTypeGold)
	g.HandleDrop("player", gold.Id)
	drop, _ := firstEntityWithComponent(g, component.ComponentIdDroppedItem)
	other := model.NewEntityId()
	g.HandleRegister("other", other, "Other")
	g.update()
	g.HandleInteract("player", drop, component.InteractionOptionLoot)
	g.HandleInteract("other", drop, component.InteractionOptionLoot)
	g.update()
	total := 0
	for _, id := range []model.EntityId{player, other} {
		inv := g.componentManager.GetEntityComponent(component.ComponentIdInventory, id).(*component.CInventory)
		if stack := inv.FindByType(model.ItemTypeGold); stack != nil {
			total += stack.Quantity
		}
	}
	if total != 200 || g.componentManager.HasEntity(drop) {
		t.Fatalf("competing pickups left %d coins", total)
	}
}
