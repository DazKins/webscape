package game

import (
	"encoding/json"
	"strings"
	"testing"
	"webscape/server/game/component"
	"webscape/server/game/model"
	"webscape/server/game/world"
	"webscape/server/message"
	"webscape/server/util"
)

func TestManaReplicationAndOfflinePersistence(t *testing.T) {
	g := snapshotGame(t)
	var sent []message.Message
	g.RegisterSender(func(_ string, msg message.Message) { sent = append(sent, msg) })
	id := model.NewEntityId()
	g.HandleRegister("player", id, "Player")
	ids, _ := streamedComponentIDs(t, sent)
	if !ids["health"] || !ids["mana"] {
		t.Fatal("initial snapshot omitted vitals")
	}
	sent = nil
	g.syncClient("player")
	ids, _ = streamedComponentIDs(t, sent)
	if ids["mana"] {
		t.Fatal("unchanged mana produced a redundant delta")
	}
	mana := g.componentManager.GetEntityComponent(component.ComponentIdMana, id).(*component.CMana)
	mana.Spend(37)
	sent = nil
	g.syncClient("player")
	ids, _ = streamedComponentIDs(t, sent)
	if !ids["mana"] {
		t.Fatal("spent mana was not replicated")
	}
	data := g.clients["player"].baseline[component.ComponentIdMana][id].(util.JObject)
	if data["currentMana"] != util.JNumber(63) || data["maxMana"] != util.JNumber(100) {
		t.Fatalf("replicated mana = %v", data)
	}
	g.HandleLeave("player")
	for range 10 {
		g.update()
	}
	saved := savedBytes(t, g)
	restored := snapshotGame(t)
	if err := restored.RestoreSnapshot(saved); err != nil {
		t.Fatal(err)
	}
	restored.HandleRegister("returning", id, "Player")
	mana = restored.componentManager.GetEntityComponent(component.ComponentIdMana, id).(*component.CMana)
	if mana.GetCurrentMana() != 63 {
		t.Fatalf("offline/restore changed mana: %d", mana.GetCurrentMana())
	}
	data = restored.clients["returning"].baseline[component.ComponentIdMana][id].(util.JObject)
	if data["currentMana"] != util.JNumber(63) {
		t.Fatal("reconnect omitted current mana")
	}
}

func TestLegacyPlayerSaveReceivesMana(t *testing.T) {
	g := snapshotGame(t)
	id := model.NewEntityId()
	g.HandleRegister("player", id, "Player")
	var save gameSnapshot
	if err := json.Unmarshal(savedBytes(t, g), &save); err != nil {
		t.Fatal(err)
	}
	delete(save.Players[id.String()], "mana")
	data, err := json.Marshal(save)
	if err != nil {
		t.Fatal(err)
	}
	restored := snapshotGame(t)
	if err := restored.RestoreSnapshot(data); err != nil {
		t.Fatal(err)
	}
	restored.HandleRegister("returning", id, "Player")
	mana := restored.componentManager.GetEntityComponent(component.ComponentIdMana, id).(*component.CMana)
	if mana.GetMaxMana() != 100 || mana.GetCurrentMana() != 100 {
		t.Fatal("legacy player missing full default mana")
	}
	// The migrated component must itself survive another restart.
	reloaded := snapshotGame(t)
	if err := reloaded.RestoreSnapshot(savedBytes(t, restored)); err != nil {
		t.Fatal(err)
	}
}

func TestAuthoredManaSettingsControlNewPlayersAndCombat(t *testing.T) {
	fs := interestWorldFS()
	fs["game.json"].Data = []byte(strings.Replace(string(fs["game.json"].Data), `{`, `{"mana":{"maxMana":40,"staffCastCost":7,"regenAmount":2,"regenIntervalTicks":10},`, 1))
	w, err := world.LoadFromGameFS(fs)
	if err != nil {
		t.Fatal(err)
	}
	g := NewGameWithWorld(w)
	g.RegisterSender(func(string, message.Message) {})
	id, target := model.NewEntityId(), model.NewEntityId()
	g.HandleRegister("player", id, "Player")
	g.HandleRegister("target", target, "Target")
	movePlayerForTest(g, target, 1, 0)
	mana := g.componentManager.GetEntityComponent(component.ComponentIdMana, id).(*component.CMana)
	if mana.GetCurrentMana() != 40 || mana.GetMaxMana() != 40 {
		t.Fatal("new player ignored authored mana maximum")
	}
	inventory := g.componentManager.GetEntityComponent(component.ComponentIdInventory, id).(*component.CInventory)
	for _, item := range inventory.GetAllItems() {
		if item.DefinitionID == "magicStaff" {
			g.HandleEquip("player", item.Id)
			break
		}
	}
	g.stateTransitions.BeginCombat(id, component.NewCCombatState(target))
	for range 3 {
		g.update()
	}
	if mana.GetCurrentMana() != 33 {
		t.Fatalf("staff ignored authored cost: mana=%d", mana.GetCurrentMana())
	}
	g.stateTransitions.EndCombat(id)
	for range 7 {
		g.update()
	}
	if mana.GetCurrentMana() != 35 {
		t.Fatalf("regeneration ignored authored settings: mana=%d", mana.GetCurrentMana())
	}
}
