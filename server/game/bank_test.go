package game

import (
	"encoding/json"
	"strings"
	"testing"
	"webscape/server/game/component"
	"webscape/server/game/gameevent"
	"webscape/server/game/model"
	"webscape/server/game/system"
	"webscape/server/math"
	"webscape/server/message"
	"webscape/server/util"
)

func newBankTestGame(t *testing.T) (*Game, model.EntityId, model.EntityId, *component.CInventory) {
	t.Helper()
	g, player, inventory := newDropTestGame(t)
	target := g.componentManager.CreateNewEntity(component.NewCPosition(math.Vec2{X: 1, Y: 0}), &component.CBanker{})
	g.update()
	return g, player, target, inventory
}

func bankDeltas(t *testing.T, messages []message.Message) map[string]any {
	t.Helper()
	result := map[string]any{}
	for _, msg := range messages {
		if msg.Metadata.Type != message.MessageTypeGameUpdate {
			continue
		}
		var wire struct {
			Data struct {
				Entities []struct {
					ComponentId string
					Data        any
				}
			}
		}
		if err := json.Unmarshal([]byte(msg.Marshal()), &wire); err != nil {
			t.Fatal(err)
		}
		for _, delta := range wire.Data.Entities {
			if delta.ComponentId == "bank" || delta.ComponentId == "banking" {
				result[delta.ComponentId] = delta.Data
			}
		}
	}
	return result
}

func TestBankLifecyclePrivacyAndStateBeforeResults(t *testing.T) {
	g, player, target, inventory := newBankTestGame(t)
	g.HandleRegister("other", model.NewEntityId(), "Other")
	sent := map[string][]message.Message{}
	g.RegisterSender(func(client string, msg message.Message) { sent[client] = append(sent[client], msg) })
	g.update()
	if len(bankDeltas(t, sent["player"])) != 0 {
		t.Fatal("bank exposed outside session")
	}
	g.HandleInteract("player", target, component.InteractionOptionBank)
	g.update()
	if err := system.ValidateEntityState(g.componentManager, player); err != nil {
		t.Fatal(err)
	}
	for a, b := range map[model.EntityId]model.EntityId{player: target, target: player} {
		facing := g.componentManager.GetEntityComponent(component.ComponentIdFacing, a)
		if facing == nil || facing.(*component.CFacing).GetTargetEntityId() != b {
			t.Fatal("bank participants not facing")
		}
	}
	if len(bankDeltas(t, sent["player"])) != 2 || len(bankDeltas(t, sent["other"])) != 0 {
		t.Fatal("incorrect bank visibility")
	}
	collected := 0
	g.RegisterGameEventHandler(gameevent.HandlerFunc(func(event gameevent.Event) {
		if strings.HasPrefix(event.Id, "collect:") {
			collected++
		}
	}))
	item := inventory.GetAllItems()[0]
	sent = map[string][]message.Message{}
	g.HandleBankTransfer("player", target, true, item.Id, 1)
	g.update()
	types := messageTypes(sent["player"])
	state, result := indexOfMessageType(types, message.MessageTypeGameUpdate), indexOfMessageType(types, message.MessageTypeBankResult)
	if inventory.HasItem(item.Id) || state < 0 || result <= state {
		t.Fatal("deposit failed or result preceded state")
	}
	if indexOfMessageType(messageTypes(sent["other"]), message.MessageTypeBankResult) >= 0 || len(bankDeltas(t, sent["other"])) != 0 {
		t.Fatal("bank result or contents leaked")
	}
	g.HandleBankTransfer("player", target, false, item.Id, 1)
	if !inventory.HasItem(item.Id) || collected != 0 {
		t.Fatal("withdrawal lost item or counted as loot")
	}
	g.HandleBankClose("player", model.NewEntityId())
	if g.componentManager.GetEntityComponent(component.ComponentIdBanking, player) == nil {
		t.Fatal("stale close ended session")
	}
	sent = map[string][]message.Message{}
	g.HandleMove("player", 0, 1)
	g.update()
	deltas := bankDeltas(t, sent["player"])
	if len(deltas) != 2 || deltas["bank"] != nil || deltas["banking"] != nil {
		t.Fatal("leaving bank did not remove private replicated state")
	}
	if g.componentManager.GetEntityComponent(component.ComponentIdFacing, target) != nil {
		t.Fatal("banker still facing departed player")
	}
}

func TestBankRejectsInvalidAccess(t *testing.T) {
	for _, reason := range []string{"closed", "distance", "visibility", "removed", "combat", "death", "other banker", "equipped", "unknown player"} {
		t.Run(reason, func(t *testing.T) {
			g, player, target, inventory := newBankTestGame(t)
			if !g.StartBankingFor(player, target) {
				t.Fatal("cannot bank")
			}
			item := inventory.GetAllItems()[0]
			client := "player"
			switch reason {
			case "closed":
				g.HandleBankClose(client, target)
			case "distance":
				movePlayerForTest(g, player, 0, 3)
			case "visibility":
				g.clients[client].loadedChunks = nil
			case "removed":
				g.componentManager.RemoveEntity(target)
			case "combat":
				g.componentManager.SetEntityComponent(target, component.NewCCombatState(player))
			case "death":
				g.componentManager.SetEntityComponent(player, component.NewCHealth(100, 0))
			case "other banker":
				target = g.componentManager.CreateNewEntity(component.NewCPosition(math.Vec2{X: 0, Y: 1}), &component.CBanker{})
			case "equipped":
				g.HandleEquip(client, item.Id)
			case "unknown player":
				client = "stranger"
			}
			bank := g.componentManager.GetEntityComponent(component.ComponentIdBank, player).(*component.CBank)
			beforeInventory, beforeBank := inventory.Serialize(), bank.Serialize()
			g.HandleBankTransfer(client, target, true, item.Id, 1)
			if !util.JsonEqual(beforeInventory, inventory.Serialize()) || !util.JsonEqual(beforeBank, bank.Serialize()) {
				t.Fatal("unauthorized transfer changed items")
			}
			switch reason {
			case "distance", "visibility", "removed", "combat", "death":
				validator := system.BankingSystem{SystemBase: system.SystemBase{ComponentManager: g.componentManager}, Validator: g}
				validator.Update()
				if g.componentManager.GetEntityComponent(component.ComponentIdBanking, player) != nil {
					t.Fatal("invalid session survived")
				}
			}
		})
	}
}

func TestBankSharedAcrossBankersAndIsolatedBetweenPlayers(t *testing.T) {
	g, player, first, inventory := newBankTestGame(t)
	second := g.componentManager.CreateNewEntity(component.NewCPosition(math.Vec2{X: 0, Y: 1}), &component.CBanker{})
	g.HandleRegister("other", model.NewEntityId(), "Other")
	g.update()
	g.StartBankingFor(player, first)
	item := inventory.GetAllItems()[0]
	g.HandleBankTransfer("player", first, true, item.Id, 1)
	g.HandleBankClose("player", first)
	g.HandleInteract("other", first, component.InteractionOptionBank)
	g.update()
	g.HandleBankTransfer("other", first, false, item.Id, 1)
	if !g.StartBankingFor(player, second) {
		t.Fatal("second banker unavailable")
	}
	g.HandleBankTransfer("player", second, false, item.Id, 1)
	if !inventory.HasItem(item.Id) {
		t.Fatal("bank not global to character or stolen by another player")
	}
	before := inventory.Serialize()
	g.HandleBankTransfer("player", second, false, item.Id, 1)
	if !util.JsonEqual(before, inventory.Serialize()) {
		t.Fatal("repeat withdrawal duplicated item")
	}
}

func TestBankPersistsWithoutSessionAndMigratesOldPlayers(t *testing.T) {
	g := snapshotGame(t)
	player := model.NewEntityId()
	g.HandleRegister("player", player, "Player")
	inventory := g.componentManager.GetEntityComponent(component.ComponentIdInventory, player).(*component.CInventory)
	bank := g.componentManager.GetEntityComponent(component.ComponentIdBank, player).(*component.CBank)
	item := inventory.GetAllItems()[0]
	item.Properties = &model.ItemProperties{CustomName: "Saved heirloom"}
	bank.Transfer(inventory, true, item.Id, 1)
	g.componentManager.SetEntityComponent(player, &component.CBanking{TargetEntityId: model.NewEntityId()})
	g.HandleLeave("player")
	saved := savedBytes(t, g)
	restored := snapshotGame(t)
	if err := restored.RestoreSnapshot(saved); err != nil {
		t.Fatal(err)
	}
	restored.HandleRegister("returning", player, "Player")
	got := restored.componentManager.GetEntityComponent(component.ComponentIdBank, player).(*component.CBank)
	if len(got.GetAllItems()) != 1 || !util.JsonEqual(component.SerializeItem(item), component.SerializeItem(got.GetAllItems()[0])) {
		t.Fatal("saved bank changed")
	}
	if restored.componentManager.GetEntityComponent(component.ComponentIdBanking, player) != nil {
		t.Fatal("bank session restored")
	}
	var snapshot gameSnapshot
	if err := json.Unmarshal(saved, &snapshot); err != nil {
		t.Fatal(err)
	}
	delete(snapshot.Entities[player.String()], "bank")
	legacy, _ := json.Marshal(snapshot)
	restored = snapshotGame(t)
	if err := restored.RestoreSnapshot(legacy); err != nil {
		t.Fatal(err)
	}
	restored.HandleRegister("legacy", player, "Player")
	if got := restored.componentManager.GetEntityComponent(component.ComponentIdBank, player); got == nil || len(got.(*component.CBank).GetAllItems()) != 0 {
		t.Fatal("old player did not get empty bank")
	}
	// A save must never claim the same item in both containers.
	record, err := (&component.CInventory{}).Save()
	if err != nil {
		t.Fatal(err)
	}
	record.Data, _ = json.Marshal(map[string]any{"items": []*model.Item{item}})
	snapshot.Entities[player.String()]["inventory"] = record
	bankRecord, _ := bank.Save()
	snapshot.Entities[player.String()]["bank"] = bankRecord
	duplicated, _ := json.Marshal(snapshot)
	if err := snapshotGame(t).RestoreSnapshot(duplicated); err == nil {
		t.Fatal("duplicated item in save accepted")
	}
}

func TestWillowbrookBankCanBeReachedFromSpawn(t *testing.T) {
	g := snapshotGame(t)
	bankers := g.componentManager.GetComponent(component.ComponentIdBanker)
	if len(bankers) != 1 {
		t.Fatalf("expected one banker, got %d", len(bankers))
	}
	var banker model.EntityId
	for id := range bankers {
		banker = id
	}
	spawn := g.world.GetPlayerSpawn()
	spawnChunk, _ := g.world.GlobalToChunk(spawn.X, spawn.Y)
	position := g.componentManager.GetEntityComponent(component.ComponentIdPosition, banker).(*component.CPosition).GetPosition()
	bankChunk, _ := g.world.GlobalToChunk(position.X, position.Y)
	if bankChunk.X != spawnChunk.X || bankChunk.Y != spawnChunk.Y-1 {
		t.Fatalf("banker is in chunk %v, expected directly north of spawn chunk %v", bankChunk, spawnChunk)
	}
	player := model.NewEntityId()
	g.HandleRegister("bank-visitor", player, "Bank Visitor")
	g.HandleInteract("bank-visitor", banker, component.InteractionOptionBank)
	// Exercise actual pathing, doors, scenery collision, and chunk streaming.
	for range 100 {
		g.update()
		if active := g.componentManager.GetEntityComponent(component.ComponentIdBanking, player); active != nil {
			if active.(*component.CBanking).TargetEntityId != banker {
				t.Fatal("opened the wrong bank")
			}
			return
		}
	}
	t.Fatal("player could not walk from spawn through the bank entrance and reach the banker")
}
