package game

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"
	"webscape/server/config"
	"webscape/server/game/component"
	"webscape/server/game/entity"
	"webscape/server/game/gameevent"
	"webscape/server/game/model"
	"webscape/server/math"
	"webscape/server/message"
)

func TestAdminChatAuthorizationAndPrivacy(t *testing.T) {
	enabled, disabled := true, false
	for _, tt := range []struct {
		name     string
		settings config.AdminCommandsConfig
		dev      bool
		command  string
		allowed  bool
	}{
		{"default disabled", config.AdminCommandsConfig{}, false, "/reset", false},
		{"default dev", config.AdminCommandsConfig{}, true, "/reset", true},
		{"explicit disabled", config.AdminCommandsConfig{Enabled: &disabled}, true, "/reset", false},
		{"prod empty", config.AdminCommandsConfig{Enabled: &enabled}, false, "/reset", false},
		{"not listed", config.AdminCommandsConfig{Enabled: &enabled, PlayerIDs: []string{model.NewEntityId().String()}}, true, "/reset", false},
		{"unknown", config.AdminCommandsConfig{}, true, "/unknown", false},
		{"arguments", config.AdminCommandsConfig{}, true, "/reset someone", false},
		{"whitespace", config.AdminCommandsConfig{}, true, "  /reset \n", true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGameWithWorld(loadInterestWorld(t))
			sent := map[string][]message.Message{}
			g.RegisterSender(func(client string, msg message.Message) { sent[client] = append(sent[client], msg) })
			id := model.NewEntityId()
			g.HandleRegister("owner", id, "Owner")
			g.HandleRegister("other", model.NewEntityId(), "Other")
			g.ConfigureAdminCommands(tt.settings, tt.dev)
			health := g.componentManager.GetEntityComponent(component.ComponentIdHealth, id).(*component.CHealth)
			health.SetCurrentHealth(1)
			speech := 0
			g.RegisterGameEventHandlerFor(gameevent.EventIdChatSpoken, gameevent.HandlerFunc(func(gameevent.Event) { speech++ }))
			sent = map[string][]message.Message{}
			g.HandleChat("unregistered", "/reset")
			g.HandleChat("owner", tt.command)
			got := g.componentManager.GetEntityComponent(component.ComponentIdHealth, id).(*component.CHealth).GetCurrentHealth()
			if (got == 100) != tt.allowed {
				t.Fatalf("health=%d allowed=%v", got, tt.allowed)
			}
			g.update()
			if speech != 0 {
				t.Fatal("slash command emitted public speech")
			}
			types := messageTypes(sent["owner"])
			stateIndex := indexOfMessageType(types, message.MessageTypeGameUpdate)
			replyIndex := indexOfMessageType(types, message.MessageTypeAdminCommandResult)
			if stateIndex < 0 || replyIndex < 0 || stateIndex >= replyIndex {
				t.Fatalf("response precedes state: %v", types)
			}
			var response struct {
				Data struct {
					Success bool
					Message string
				}
			}
			if err := json.Unmarshal([]byte(sent["owner"][replyIndex].Marshal()), &response); err != nil {
				t.Fatal(err)
			}
			if response.Data.Success != tt.allowed || response.Data.Message == "" {
				t.Fatalf("invalid reply: %+v", response)
			}
			if len(sent["unregistered"]) != 0 {
				t.Fatal("unregistered command received a reply")
			}
			for _, msg := range sent["other"] {
				if msg.Metadata.Type == message.MessageTypeAdminCommandResult {
					t.Fatal("private reply leaked")
				}
			}
		})
	}
}

func TestResetReplacesAllCharacterStateAndPersists(t *testing.T) {
	g := snapshotGame(t)
	id, other := model.NewEntityId(), model.NewEntityId()
	g.HandleRegister("owner", id, "Owner")
	g.HandleRegister("other", other, "Other")
	enabled := true
	g.ConfigureAdminCommands(config.AdminCommandsConfig{Enabled: &enabled, PlayerIDs: []string{id.String()}}, false)
	inventory := g.componentManager.GetEntityComponent(component.ComponentIdInventory, id).(*component.CInventory)
	oldItems := map[model.ItemId]bool{}
	for _, item := range inventory.GetAllItems() {
		oldItems[item.Id] = true
	}
	g.HandleEquip("owner", inventory.GetAllItems()[0].Id)
	g.componentManager.SetEntityComponent(id, component.NewCPosition(math.Vec2{X: 4, Y: 5}))
	g.componentManager.SetEntityComponent(id, component.NewCHealth(3, 200))
	quests := g.componentManager.GetEntityComponent(component.ComponentIdQuestLog, id).(*component.CQuestLog)
	quests.CompleteQuest("old-quest")
	quests.StartQuest("active-quest", "step")
	g.componentManager.GetEntityComponent(component.ComponentIdCombatLog, id).(*component.CCombatLog).AddEntry(component.NewCombatLogEntry("old", "hit"))
	g.componentManager.SetEntityComponent(id, component.NewCActiveConversation("old", other, "node"))
	g.componentManager.SetEntityComponent(id, component.NewCCombatState(other))
	g.componentManager.SetEntityComponent(other, component.NewCCombatState(id))
	g.componentManager.SetEntityComponent(id, component.NewCRewardDrop()) // arbitrary extra component must disappear
	g.syncClient("owner")
	beforeWorld := savedBytes(t, g)
	g.HandleChat("owner", "/reset")
	if !g.IsRegistered("owner") {
		t.Fatal("reset lost registration")
	}
	defaults := entity.CreatePlayerEntity(id, "Owner", g.world.GetPlayerSpawn(), g.currentTick)
	expected := map[component.ComponentId]bool{}
	for _, want := range defaults {
		expected[want.GetId()] = true
		got := g.componentManager.GetEntityComponent(want.GetId(), id)
		if got == nil {
			t.Fatalf("missing %s", want.GetId())
		}
		switch want.GetId() {
		case component.ComponentIdAppearance:
			if err := component.ValidateAppearance(got.(*component.CAppearance).GetAppearance()); err != nil {
				t.Fatal(err)
			}
		case component.ComponentIdInventory:
			counts := func(inv *component.CInventory) map[string]int {
				m := map[string]int{}
				for _, item := range inv.GetAllItems() {
					m[item.Type()] += item.Quantity
				}
				return m
			}
			if !reflect.DeepEqual(counts(got.(*component.CInventory)), counts(want.(*component.CInventory))) {
				t.Fatal("starter inventory differs")
			}
			for _, item := range got.(*component.CInventory).GetAllItems() {
				if oldItems[item.Id] {
					t.Fatal("old item survived")
				}
			}
		default:
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("%s did not reset: %#v", want.GetId(), got)
			}
		}
	}
	for cid, entities := range g.componentManager.GetAllComponents() {
		if entities[id] != nil && !expected[cid] {
			t.Fatalf("extra component survived: %s", cid)
		}
	}
	if g.componentManager.GetEntityComponent(component.ComponentIdCombatState, other) != nil {
		t.Fatal("opponent still attacking")
	}
	// Existing replication baselines must produce removals for old action state.
	sent := []message.Message{}
	g.RegisterSender(func(client string, msg message.Message) {
		if client == "owner" {
			sent = append(sent, msg)
		}
	})
	g.syncClient("owner")
	foundRemoval := false
	for _, msg := range sent {
		if msg.Metadata.Type != message.MessageTypeGameUpdate {
			continue
		}
		var payload struct {
			Data struct {
				Entities []struct {
					EntityID    string          `json:"entityId"`
					ComponentID string          `json:"componentId"`
					Data        json.RawMessage `json:"data"`
				}
			}
		}
		if err := json.Unmarshal([]byte(msg.Marshal()), &payload); err != nil {
			t.Fatal(err)
		}
		for _, update := range payload.Data.Entities {
			if update.EntityID == id.String() && update.ComponentID == string(component.ComponentIdCombatState) && string(update.Data) == "null" {
				foundRemoval = true
			}
		}
	}
	if !foundRemoval {
		t.Fatal("reset did not replicate removed components")
	}
	before := savedBytes(t, g)
	var worldBefore, worldAfter gameSnapshot
	if err := json.Unmarshal(beforeWorld, &worldBefore); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(before, &worldAfter); err != nil {
		t.Fatal(err)
	}
	delete(worldBefore.Entities, id.String())
	delete(worldAfter.Entities, id.String())
	if !reflect.DeepEqual(worldBefore, worldAfter) {
		t.Fatal("reset changed another character or the world")
	}
	g.HandleLeave("owner")
	g.HandleRegister("rejoined", id, "Ignored")
	if !bytes.Equal(before, savedBytes(t, g)) {
		t.Fatal("reconnect changed reset state")
	}
	restored := snapshotGame(t)
	if err := restored.RestoreSnapshot(before); err != nil {
		t.Fatal(err)
	}
	restored.HandleRegister("restored", id, "Ignored")
	var want, got any
	if err := json.Unmarshal(before, &want); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(savedBytes(t, restored), &got); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(want, got) {
		t.Fatal("reset state did not survive snapshot restore")
	}
}
