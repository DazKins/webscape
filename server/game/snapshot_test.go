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

func TestSnapshotRestoresPlayersWorldAndReferences(t *testing.T) {
	g := snapshotGame(t)
	id := model.NewEntityId()
	g.HandleRegister("first", id, "Saved Player")
	for range 3 {
		g.update()
	}
	inventory := g.componentManager.GetEntityComponent(component.ComponentIdInventory, id).(*component.CInventory)
	sword := inventory.GetAllItems()[0]
	g.HandleEquip("first", sword.Id)
	// Mutate fields which replication does not fully encode, including item IDs,
	// loot contents and the spawner's child reference.
	g.componentManager.SetEntityComponent(id, component.NewCPosition(math.Vec2{X: 4, Y: 5}))
	g.componentManager.GetEntityComponent(component.ComponentIdHealth, id).(*component.CHealth).SetCurrentHealth(42)
	quests := g.componentManager.GetEntityComponent(component.ComponentIdQuestLog, id).(*component.CQuestLog)
	for _, q := range g.world.GetQuestRegistry().All() {
		quests.SetProgress(q.Id, 0, q.Steps[0].Id, 0)
		break
	}
	for _, c := range g.componentManager.GetComponent(component.ComponentIdLootable) {
		c.(*component.CLootable).SetLooted(true)
		break
	}
	for _, c := range g.componentManager.GetComponent(component.ComponentIdOpenable) {
		c.(*component.COpenable).SetOpen(true)
		break
	}
	for _, c := range g.componentManager.GetComponent(component.ComponentIdWoodcuttable) {
		tree := c.(*component.CWoodcuttable)
		tree.SetCurrentDurability(0)
		tree.SetDepleted(true)
		tree.SetRemainingRespawnTicks(7)
		break
	}
	drop := model.CreateGold(17)
	droppedID := g.componentManager.CreateNewEntity(component.NewCPosition(math.Vec2{X: 1, Y: 1}), &component.CDroppedItem{Item: drop})
	var deleted model.EntityId
	for entityID := range g.componentManager.GetComponent(component.ComponentIdShop) {
		deleted = entityID
		g.componentManager.RemoveEntity(entityID)
		break
	}
	before := savedBytes(t, g)
	restored := snapshotGame(t)
	if err := restored.RestoreSnapshot(before); err != nil {
		t.Fatal(err)
	}
	if restored.componentManager.HasEntity(id) {
		t.Fatal("offline player entered active simulation")
	}
	if deleted != (model.EntityId{}) && restored.componentManager.HasEntity(deleted) {
		t.Fatal("deleted entity resurrected")
	}
	if got := restored.componentManager.GetEntityComponent(component.ComponentIdDroppedItem, droppedID).(*component.CDroppedItem).Item; got.Id != drop.Id || got.Quantity != 17 {
		t.Fatal("dropped item lost identity or quantity")
	}
	if after := savedBytes(t, restored); !bytes.Equal(before, after) {
		t.Fatal("snapshot round-trip changed durable state")
	}
	restored.HandleRegister("second", id, "Replacement Name")
	if !restored.IsRegistered("second") {
		t.Fatal("returning player rejected")
	}
	if got := restored.componentManager.GetEntityComponent(component.ComponentIdPlayer, id).(*component.CPlayer).GetName(); got != "Saved Player" {
		t.Fatal("saved identity overwritten")
	}
	if got := restored.componentManager.GetEntityComponent(component.ComponentIdEquipped, id).(*component.CEquipped).GetEquippedItem(model.SlotWeapon); got == nil || got.Id != sword.Id {
		t.Fatal("equipped item not restored")
	}
	// A second restart must not create a second child for an existing spawner.
	children := map[model.EntityId]model.EntityId{}
	for parent, c := range restored.componentManager.GetComponent(component.ComponentIdSpawn) {
		spawn := c.(*component.CSpawn)
		if spawn.HasChildEntityId() {
			children[parent] = spawn.GetChildEntityId()
		}
	}
	restored.update()
	for parent, child := range children {
		if got := restored.componentManager.GetEntityComponent(component.ComponentIdSpawn, parent).(*component.CSpawn).GetChildEntityId(); got != child {
			t.Fatal("spawn child duplicated")
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
	if g.componentManager.GetEntityComponent(component.ComponentIdFacing, id) != nil {
		t.Fatal("stale target restored")
	}
	if got := g.componentManager.GetEntityComponent(component.ComponentIdLocomotion, id).(*component.CLocomotion).GetPhase(); got != component.LocomotionPhaseIdle {
		t.Fatal("movement restored")
	}
}

func TestSnapshotRestoresAcrossContentChanges(t *testing.T) {
	g := snapshotGame(t)
	g.HandleRegister("one", model.NewEntityId(), "Saved Player")
	before := savedBytes(t, g)
	var saved gameSnapshot
	if err := json.Unmarshal(before, &saved); err != nil {
		t.Fatal(err)
	}
	saved.ContentHash = "previous-authored-content"
	data, err := json.Marshal(saved)
	if err != nil {
		t.Fatal(err)
	}
	restored := snapshotGame(t)
	if err := restored.RestoreSnapshot(data); err != nil {
		t.Fatal(err)
	}
	// Player and world state survive; the next save records the current content hash.
	if after := savedBytes(t, restored); !bytes.Equal(before, after) {
		t.Fatal("content change altered saved state or retained the old hash")
	}
}

func TestInvalidSnapshotsDoNotPartiallyReplaceWorld(t *testing.T) {
	g := snapshotGame(t)
	id := model.NewEntityId()
	g.HandleRegister("one", id, "Player")
	valid := savedBytes(t, g)
	for _, change := range []func(*gameSnapshot){
		func(s *gameSnapshot) { s.Version = 99 },
		func(s *gameSnapshot) { s.Entities = nil },
		func(s *gameSnapshot) { delete(s.Entities[id.String()], string(component.ComponentIdInventory)) },
		func(s *gameSnapshot) {
			s.Entities[id.String()]["unknown"] = component.SavedComponent{Version: 1, Data: json.RawMessage(`{}`)}
		},
		func(s *gameSnapshot) {
			s.Entities[id.String()][string(component.ComponentIdInventory)] = component.SavedComponent{Version: 1, Data: json.RawMessage(`{"items":[null]}`)}
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
	if _, ok := s.Entities[id.String()]; ok {
		t.Fatal("none mode retained player")
	}
}
