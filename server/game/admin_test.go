package game

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"webscape/server/config"
	"webscape/server/game/component"
	"webscape/server/game/entity"
	"webscape/server/game/gameevent"
	"webscape/server/game/model"
	"webscape/server/math"
	"webscape/server/message"
)

func TestGiveAdminCommand(t *testing.T) {
	disabled, enabled := false, true
	for _, tt := range []struct {
		name, command, wantID, reply string
		full, overflow, prod         bool
		empty                        bool
		quantity                     int
		settings                     config.AdminCommandsConfig
	}{
		{name: "equipment", command: "/give chainmailChestplate", wantID: "chainmailChestplate", reply: "Added Chainmail Chestplate"},
		{name: "existing item", command: "/give ironSword", wantID: "ironSword", reply: "Added Iron Sword"},
		{name: "whitespace", command: "  /give\tapple  ", wantID: "apple", reply: "Added Apple"},
		{name: "stack in full inventory", command: "/give gold", wantID: "gold", full: true, reply: "Added Gold"},
		{name: "full inventory", command: "/give apple", full: true, reply: "full"},
		{name: "overflow", command: "/give gold", overflow: true, reply: "full"},
		{name: "unknown", command: "/give missing", reply: "Unknown item definition"},
		{name: "case sensitive", command: "/give IronSword", reply: "Unknown item definition"},
		{name: "missing argument", command: "/give", reply: "Usage:"},
		{name: "extra argument", command: "/give ironSword 2 extra", reply: "Usage:"},
		{name: "multiple non-stackables", command: "/give ironSword 3", wantID: "ironSword", quantity: 3, reply: "Added 3 x Iron Sword"},
		{name: "exact available slots", command: "/give ironSword 9", wantID: "ironSword", quantity: 9, reply: "Added 9 x Iron Sword"},
		{name: "no partial grant", command: "/give ironSword 10", reply: "full"},
		{name: "large non-stackable request", command: "/give ironSword 2147483647", reply: "full"},
		{name: "merge quantity", command: "/give gold 1000", wantID: "gold", quantity: 1000, full: true, reply: "Added 1000 x Gold"},
		{name: "new stack", command: "/give arrow 250", wantID: "arrow", quantity: 250, empty: true, reply: "Added 250 x Arrow"},
		{name: "new stack cannot fit", command: "/give gold 2", empty: true, full: true, reply: "full"},
		{name: "stack limit", command: "/give gold 2147483547", wantID: "gold", quantity: 2147483547, reply: "Added 2147483547 x Gold"},
		{name: "batch exceeds stack limit", command: "/give gold 2147483548", reply: "full"},
		{name: "zero quantity", command: "/give gold 0", reply: "Quantity must"},
		{name: "negative quantity", command: "/give gold -2", reply: "Quantity must"},
		{name: "fractional quantity", command: "/give gold 1.5", reply: "Quantity must"},
		{name: "non-numeric quantity", command: "/give gold lots", reply: "Quantity must"},
		{name: "quantity exceeds limit", command: "/give gold 2147483648", reply: "Quantity must"},
		{name: "quantity parse overflow", command: "/give gold 99999999999999999999", reply: "Quantity must"},
		{name: "disabled", command: "/give apple", settings: config.AdminCommandsConfig{Enabled: &disabled}, reply: "disabled"},
		{name: "not allowlisted", command: "/give apple", settings: config.AdminCommandsConfig{PlayerIDs: []string{model.NewEntityId().String()}}, reply: "permission"},
		{name: "production default", command: "/give apple", prod: true, reply: "disabled"},
		{name: "production empty allowlist", command: "/give apple", prod: true, settings: config.AdminCommandsConfig{Enabled: &enabled}, reply: "permission"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGameWithWorld(loadInterestWorld(t))
			sent := map[string][]message.Message{}
			g.RegisterSender(func(client string, msg message.Message) { sent[client] = append(sent[client], msg) })
			id, other := model.NewEntityId(), model.NewEntityId()
			g.HandleRegister("owner", id, "Owner")
			g.HandleRegister("other", other, "Other")
			g.ConfigureAdminCommands(tt.settings, !tt.prod)
			inventory := g.componentManager.GetEntityComponent(component.ComponentIdInventory, id).(*component.CInventory)
			if tt.empty {
				inventory = component.NewCInventory()
				g.componentManager.SetEntityComponent(id, inventory)
			}
			if tt.full {
				for !inventory.IsFull() {
					if !inventory.AddItem(model.NewItem("bread")) {
						t.Fatal("could not fill inventory")
					}
				}
			}
			if tt.overflow {
				inventory.FindByDefinition("gold").Quantity = model.MaxStackQuantity
			}
			before := inventory.Clone()
			otherInventory := g.componentManager.GetEntityComponent(component.ComponentIdInventory, other).(*component.CInventory)
			otherBefore := otherInventory.Clone()
			speech := 0
			g.RegisterGameEventHandlerFor(gameevent.EventIdChatSpoken, gameevent.HandlerFunc(func(gameevent.Event) { speech++ }))
			sent = map[string][]message.Message{}
			g.HandleChat("unregistered", tt.command)
			g.HandleChat("owner", tt.command)
			if tt.wantID == "" {
				if !reflect.DeepEqual(before, inventory) {
					t.Fatal("failed command changed inventory")
				}
			} else {
				counts := func(inv *component.CInventory) map[string]int {
					result := map[string]int{}
					for _, item := range inv.GetAllItems() {
						result[item.DefinitionID] += item.Quantity
					}
					return result
				}
				want := counts(before)
				quantity := tt.quantity
				if quantity == 0 {
					quantity = 1
				}
				want[tt.wantID] += quantity
				if !reflect.DeepEqual(want, counts(inventory)) {
					t.Fatal("give did not add exactly the requested quantity")
				}
				seenIDs := map[model.ItemId]bool{}
				stacks := map[string]int{}
				for _, item := range inventory.GetAllItems() {
					if seenIDs[item.Id] || item.ValidateSaved() != nil {
						t.Fatal("invalid item quantity or duplicate instance UUID")
					}
					seenIDs[item.Id] = true
					if item.IsStackable() {
						stacks[item.DefinitionID]++
						if stacks[item.DefinitionID] > 1 {
							t.Fatal("stackable items were not merged")
						}
					}
				}
				for _, old := range before.GetAllItems() {
					if !seenIDs[old.Id] {
						t.Fatal("give replaced an existing instance UUID")
					}
				}
			}
			if !reflect.DeepEqual(otherBefore, otherInventory) {
				t.Fatal("give changed another player's inventory")
			}
			g.update()
			if speech != 0 || len(sent["unregistered"]) != 0 {
				t.Fatal("command emitted speech or replied to an unregistered connection")
			}
			types := messageTypes(sent["owner"])
			stateIndex := indexOfMessageType(types, message.MessageTypeGameUpdate)
			replyIndex := indexOfMessageType(types, message.MessageTypeAdminCommandResult)
			if stateIndex < 0 || replyIndex <= stateIndex {
				t.Fatalf("missing reply or reply precedes state: %v", types)
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
			if response.Data.Success != (tt.wantID != "") || !strings.Contains(response.Data.Message, tt.reply) {
				t.Fatalf("unexpected response: %+v", response)
			}
			for _, msg := range sent["other"] {
				if msg.Metadata.Type == message.MessageTypeAdminCommandResult {
					t.Fatal("private reply leaked")
				}
			}
		})
	}
}

func TestGiveSupportsEveryDefinitionForAllowlistedAdmin(t *testing.T) {
	g := NewGameWithWorld(loadInterestWorld(t))
	g.RegisterSender(func(string, message.Message) {})
	id := model.NewEntityId()
	g.HandleRegister("owner", id, "Owner")
	enabled := true
	g.ConfigureAdminCommands(config.AdminCommandsConfig{Enabled: &enabled, PlayerIDs: []string{id.String()}}, false)
	seen := map[model.ItemId]bool{}
	for _, definitionID := range model.ItemDefinitionIDs() {
		inventory := component.NewCInventory()
		g.componentManager.SetEntityComponent(id, inventory)
		g.HandleChat("owner", "/give "+definitionID)
		items := inventory.GetAllItems()
		if len(items) != 1 {
			t.Fatalf("%s: expected one item, got %d", definitionID, len(items))
		}
		item := items[0]
		if item.DefinitionID != definitionID || item.Quantity != 1 || item.HasProperties() || seen[item.Id] || item.ValidateSaved() != nil {
			t.Fatalf("%s: invalid instance: %+v", definitionID, item)
		}
		seen[item.Id] = true
	}
}

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
	defaults := entity.CreatePlayerEntity(id, "Owner", g.world.GetPlayerSpawn(), g.currentTick, g.world.GetManaSettings().MaxMana)
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
	delete(worldBefore.Players, id.String())
	delete(worldAfter.Players, id.String())
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
