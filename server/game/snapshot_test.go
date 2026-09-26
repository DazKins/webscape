package game

import (
	"bytes"
	"encoding/json"
	"testing"
	"webscape/server/game/component"
	"webscape/server/game/model"
	"webscape/server/game/world"
	"webscape/server/math"
	"webscape/server/message"
)

func snapshotGame(t *testing.T) *Game {
	t.Helper()
	w, err := world.LoadFromGameFolder("../../game-project")
	if err != nil {
		t.Fatal(err)
	}
	g := NewGameWithWorld(w)
	g.RegisterSender(func(string, message.Message) {})
	g.RegisterBroadcaster(func(message.Message) {})
	g.RetainOfflinePlayers()
	return g
}
func savedBytes(t *testing.T, g *Game) []byte {
	t.Helper()
	data, err := g.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestSnapshotRestoresOnlyPlayers(t *testing.T) {
	g := snapshotGame(t)
	id := model.NewEntityId()
	g.HandleRegister("first", id, "Saved Player")
	for range 3 {
		g.update()
	}
	inventory := g.componentManager.GetEntityComponent(component.ComponentIdInventory, id).(*component.CInventory)
	sword := inventory.GetAllItems()[0]
	g.HandleEquip("first", sword.Id)
	g.HandleInventoryMove("first", inventory.GetAllItems()[0].Id, 19, 1)
	g.componentManager.GetEntityComponent(component.ComponentIdHealth, id).(*component.CHealth).SetCurrentHealth(42)
	dropID := g.componentManager.CreateNewEntity(component.NewCPosition(g.world.GetPlayerSpawn()), &component.CDroppedItem{Item: model.CreateGold(17)})
	for banker := range g.componentManager.GetComponent(component.ComponentIdBanker) {
		g.componentManager.RemoveEntity(banker)
	}
	before := savedBytes(t, g)
	var saved gameSnapshot
	json.Unmarshal(before, &saved)
	if len(saved.Players) != 1 {
		t.Fatal("saved non-player entities")
	}
	restored := snapshotGame(t)
	if err := restored.RestoreSnapshot(before); err != nil {
		t.Fatal(err)
	}
	if restored.componentManager.HasEntity(id) || restored.componentManager.HasEntity(dropID) {
		t.Fatal("restored active player or ground drop")
	}
	if len(restored.componentManager.GetComponent(component.ComponentIdBanker)) != 1 {
		t.Fatal("authored banker was not rebuilt")
	}
	if restored.currentTick != 0 {
		t.Fatal("clock did not reset")
	}
	if after := savedBytes(t, restored); !bytes.Equal(before, after) {
		t.Fatal("player state changed")
	}
	restored.HandleRegister("second", id, "Replacement")
	if got := restored.componentManager.GetEntityComponent(component.ComponentIdEquipped, id).(*component.CEquipped).GetEquippedItem(model.SlotWeapon); got == nil || got.Id != sword.Id {
		t.Fatal("equipment lost")
	}
	if got := restored.componentManager.GetEntityComponent(component.ComponentIdPlayer, id).(*component.CPlayer).GetName(); got != "Saved Player" {
		t.Fatal("identity lost")
	}
}

func TestSavedPlayerBlockedPositionFallsBackToSpawn(t *testing.T) {
	for _, pos := range []math.Vec2{{X: 99999, Y: 99999}, {X: 8, Y: -23}} {
		g := snapshotGame(t)
		id := model.NewEntityId()
		g.HandleRegister("one", id, "Player")
		g.componentManager.SetEntityComponent(id, component.NewCPosition(pos))
		restored := snapshotGame(t)
		if err := restored.RestoreSnapshot(savedBytes(t, g)); err != nil {
			t.Fatal(err)
		}
		restored.HandleRegister("two", id, "Player")
		got := restored.componentManager.GetEntityComponent(component.ComponentIdPosition, id).(*component.CPosition).GetPosition()
		if got != restored.world.GetPlayerSpawn() {
			t.Fatalf("blocked position retained: %v", got)
		}
	}
}

func TestDisconnectRetainsProgressAndResetsActivity(t *testing.T) {
	g := snapshotGame(t)
	id := model.NewEntityId()
	g.HandleRegister("one", id, "Player")
	gold := model.CreateGold(19)
	g.componentManager.SetEntityComponents(id, component.NewCInventory(), component.NewCLocomotion(component.LocomotionPhaseMoving, 9), component.NewCFacing(model.NewEntityId()))
	g.componentManager.GetEntityComponent(component.ComponentIdInventory, id).(*component.CInventory).AddItem(gold)
	g.HandleInventoryMove("one", gold.Id, 19, 1)
	beforeInventory, _ := json.Marshal(g.componentManager.GetEntityComponent(component.ComponentIdInventory, id).(*component.CInventory).Serialize())
	g.HandleLeave("one")
	if g.componentManager.HasEntity(id) {
		t.Fatal("disconnected entity still simulated")
	}
	for range 3 {
		g.update()
	}
	g.HandleRegister("two", id, "Player")
	inv := g.componentManager.GetEntityComponent(component.ComponentIdInventory, id).(*component.CInventory)
	if items := inv.GetAllItems(); len(items) != 1 || items[0].Id != gold.Id {
		t.Fatal("reconnect reset inventory")
	}
	afterInventory, _ := json.Marshal(inv.Serialize())
	if !bytes.Equal(beforeInventory, afterInventory) {
		t.Fatal("reconnect lost inventory positions")
	}
	if g.componentManager.GetEntityComponent(component.ComponentIdFacing, id) != nil {
		t.Fatal("stale target restored")
	}
	if got := g.componentManager.GetEntityComponent(component.ComponentIdLocomotion, id).(*component.CLocomotion).GetPhase(); got != component.LocomotionPhaseIdle {
		t.Fatal("movement restored")
	}
}

func TestInvalidSnapshotsDoNotPartiallyReplaceWorld(t *testing.T) {
	g := snapshotGame(t)
	id := model.NewEntityId()
	g.HandleRegister("one", id, "Player")
	valid := savedBytes(t, g)
	for _, change := range []func(*gameSnapshot){
		func(s *gameSnapshot) { s.Version = 99 },
		func(s *gameSnapshot) { s.Players = nil },
		func(s *gameSnapshot) { delete(s.Players[id.String()], string(component.ComponentIdInventory)) },
		func(s *gameSnapshot) {
			s.Players[id.String()]["unknown"] = component.SavedComponent{Version: 1, Data: json.RawMessage(`{}`)}
		},
		func(s *gameSnapshot) {
			s.Players[id.String()][string(component.ComponentIdInventory)] = component.SavedComponent{Version: 1, Data: json.RawMessage(`{"items":[null]}`)}
		},
	} {
		var s gameSnapshot
		if err := json.Unmarshal(valid, &s); err != nil {
			t.Fatal(err)
		}
		change(&s)
		data, _ := json.Marshal(s)
		target := snapshotGame(t)
		before := savedBytes(t, target)
		if err := target.RestoreSnapshot(data); err == nil {
			t.Fatal("accepted invalid snapshot")
		}
		if after := savedBytes(t, target); !bytes.Equal(before, after) {
			t.Fatal("invalid save partially replaced world")
		}
	}
}

func TestNoStorageDoesNotRetainDisconnectedPlayers(t *testing.T) {
	g := snapshotGame(t)
	g.offlinePlayers = nil
	id := model.NewEntityId()
	g.HandleRegister("one", id, "Player")
	g.HandleLeave("one")
	var s gameSnapshot
	json.Unmarshal(savedBytes(t, g), &s)
	if _, ok := s.Players[id.String()]; ok {
		t.Fatal("none mode retained player")
	}
}

func TestRestartResetsWorldInteractions(t *testing.T) {
	g := snapshotGame(t)
	// Every door is flipped, every loot container looted and every tree depleted.
	openCount := func(game *Game) int {
		count := 0
		for _, c := range game.componentManager.GetComponent(component.ComponentIdOpenable) {
			if c.(*component.COpenable).IsOpen() {
				count++
			}
		}
		return count
	}
	defaultOpen := openCount(g)
	for _, c := range g.componentManager.GetComponent(component.ComponentIdOpenable) {
		door := c.(*component.COpenable)
		door.SetOpen(!door.IsOpen())
	}
	for _, c := range g.componentManager.GetComponent(component.ComponentIdLootable) {
		c.(*component.CLootable).SetLooted(true)
	}
	for _, c := range g.componentManager.GetComponent(component.ComponentIdWoodcuttable) {
		tree := c.(*component.CWoodcuttable)
		tree.SetDepleted(true)
		tree.SetCurrentDurability(0)
	}
	restored := snapshotGame(t)
	if err := restored.RestoreSnapshot(savedBytes(t, g)); err != nil {
		t.Fatal(err)
	}
	if openCount(restored) != defaultOpen {
		t.Fatal("doors did not reset")
	}
	for _, c := range restored.componentManager.GetComponent(component.ComponentIdLootable) {
		if c.(*component.CLootable).IsLooted() {
			t.Fatal("loot did not reset")
		}
	}
	for _, c := range restored.componentManager.GetComponent(component.ComponentIdWoodcuttable) {
		if c.(*component.CWoodcuttable).IsDepleted() {
			t.Fatal("tree did not reset")
		}
	}
}

func TestGuestProgressExcludedFromSnapshots(t *testing.T) {
	g := snapshotGame(t)
	guest, account := model.NewEntityId(), model.NewEntityId()
	g.HandleRegisterGuest("guest", guest, "Guest")
	g.HandleRegister("account", account, "Account")
	g.componentManager.GetEntityComponent(component.ComponentIdHealth, guest).(*component.CHealth).SetCurrentHealth(42)
	check := func() {
		t.Helper()
		var saved gameSnapshot
		if err := json.Unmarshal(savedBytes(t, g), &saved); err != nil {
			t.Fatal(err)
		}
		if len(saved.Players) != 1 || saved.Players[account.String()] == nil || saved.Players[guest.String()] != nil {
			t.Fatal("snapshot must contain only the account")
		}
	}
	check()
	g.HandleLeave("guest")
	g.HandleLeave("account")
	check()
	g.HandleRegisterGuest("reconnected", guest, "Guest")
	if got := g.componentManager.GetEntityComponent(component.ComponentIdHealth, guest).(*component.CHealth).GetCurrentHealth(); got != 42 {
		t.Fatal("guest lost progress within session", got)
	}
	check()
}
