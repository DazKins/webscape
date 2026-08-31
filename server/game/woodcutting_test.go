package game

import (
	"encoding/json"
	"strings"
	"testing"
	"testing/fstest"
	"webscape/server/game/component"
	"webscape/server/game/gameevent"
	"webscape/server/game/model"
	"webscape/server/game/world"
	"webscape/server/message"
)

func TestWoodcuttingRollOutcomesFollowPhaseTimeline(t *testing.T) {
	tests := []struct {
		name       string
		roll       int
		wantDamage int
		wantLog    string
	}{
		{name: "miss", roll: 0, wantDamage: 0, wantLog: "miss the tree"},
		{name: "bad chop", roll: 25, wantDamage: 1, wantLog: "bad chop"},
		{name: "good chop", roll: 75, wantDamage: 2, wantLog: "good chop"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			game, treeId := setupWoodcuttingGame(t)
			events := recordGameEvents(game)
			playerId := joinWoodcutter(t, game, "client-1", "One")
			game.woodcuttingSystem.RollSource = func() int { return test.roll }

			game.HandleInteract("client-1", treeId, component.InteractionOptionChop)
			game.update()

			tree := woodcuttableFor(t, game, treeId)
			assertWoodcuttingPhase(t, game, playerId, component.WoodcuttingPhaseSwinging, 1)
			if got := tree.GetCurrentDurability(); got != 5 {
				t.Fatalf("durability on initial swinging tick = %d, want 5", got)
			}
			if got := countGameEvents(*events, gameevent.EventIdWoodcuttingSwing); got != 0 {
				t.Fatalf("swing events on initial swinging tick = %d, want 0", got)
			}

			game.update()
			assertWoodcuttingPhase(t, game, playerId, component.WoodcuttingPhaseRecovering, 2)
			if got := tree.GetCurrentDurability(); got != 5-test.wantDamage {
				t.Fatalf("durability after resolved swing = %d, want %d", got, 5-test.wantDamage)
			}
			if got := countGameEvents(*events, gameevent.EventIdWoodcuttingSwing); got != 1 {
				t.Fatalf("resolved swing events = %d, want 1", got)
			}
			assertActivityLogContains(t, game, playerId, test.wantLog)

			game.update()
			assertWoodcuttingPhase(t, game, playerId, component.WoodcuttingPhaseSwinging, 3)
			if got := tree.GetCurrentDurability(); got != 5-test.wantDamage {
				t.Fatalf("durability on next swinging tick = %d, want unchanged %d", got, 5-test.wantDamage)
			}
			if got := countGameEvents(*events, gameevent.EventIdWoodcuttingSwing); got != 1 {
				t.Fatalf("swing events during next wind-up = %d, want 1 total", got)
			}

			game.update()
			if got := tree.GetCurrentDurability(); got != 5-2*test.wantDamage {
				t.Fatalf("durability after second resolved swing = %d, want %d", got, 5-2*test.wantDamage)
			}
			if got := countGameEvents(*events, gameevent.EventIdWoodcuttingSwing); got != 2 {
				t.Fatalf("events after second resolved swing = %d, want 2 total", got)
			}
		})
	}
}

func TestWoodcuttingRequiresEquippedAxeAndInventorySpace(t *testing.T) {
	game, treeId := setupWoodcuttingGame(t)
	playerId := joinPlayer(t, game, "client-1", "One")
	game.woodcuttingSystem.RollSource = func() int { return 75 }

	game.HandleInteract("client-1", treeId, component.InteractionOptionChop)
	game.update()
	if game.componentManager.GetEntityComponent(component.ComponentIdWoodcutting, playerId) != nil {
		t.Fatal("woodcutting started without an equipped axe")
	}
	if woodcuttableFor(t, game, treeId).GetCurrentDurability() != 5 {
		t.Fatal("tree lost durability without an equipped axe")
	}
	assertActivityLogContains(t, game, playerId, "equipped axe")

	equipAxe(t, game, "client-1", playerId)
	inventory := game.componentManager.GetEntityComponent(component.ComponentIdInventory, playerId).(*component.CInventory)
	for !inventory.IsFull() {
		inventory.AddItem(model.NewItem("Filler", "test"))
	}
	game.componentManager.SetEntityComponent(playerId, inventory)

	game.HandleInteract("client-1", treeId, component.InteractionOptionChop)
	game.update()
	if woodcuttableFor(t, game, treeId).GetCurrentDurability() != 5 {
		t.Fatal("full inventory consumed tree durability")
	}
	if game.componentManager.GetEntityComponent(component.ComponentIdWoodcutting, playerId) != nil {
		t.Fatal("woodcutting remained active with a full inventory")
	}
	assertActivityLogContains(t, game, playerId, "inventory is full")
}

func TestWoodcuttingStateReplicatesToObserversAndLateInterest(t *testing.T) {
	game, treeId := setupWoodcuttingGame(t)
	sent := map[string][]message.Message{}
	game.RegisterSender(func(clientId string, msg message.Message) {
		sent[clientId] = append(sent[clientId], msg)
	})
	playerId := joinWoodcutter(t, game, "client-1", "One")
	joinPlayer(t, game, "client-2", "Two")
	sent = map[string][]message.Message{}

	game.HandleInteract("client-1", treeId, component.InteractionOptionChop)
	game.update()
	for _, clientId := range []string{"client-1", "client-2"} {
		state, removed, serverTick := woodcuttingUpdateFor(t, sent[clientId], playerId)
		if removed || state == nil || state.TargetEntityId != treeId.String() ||
			state.Phase != string(component.WoodcuttingPhaseSwinging) ||
			state.PhaseStartedTick != 1 || serverTick != 1 {
			t.Fatalf("woodcutting for %s = %#v removed=%v at tick %d", clientId, state, removed, serverTick)
		}
	}

	sent["client-3"] = nil
	joinPlayer(t, game, "client-3", "Three")
	state, removed, serverTick := woodcuttingUpdateFor(t, sent["client-3"], playerId)
	if removed || state == nil || state.Phase != string(component.WoodcuttingPhaseSwinging) ||
		state.PhaseStartedTick != 1 || serverTick != 1 {
		t.Fatalf("late woodcutting snapshot = %#v removed=%v at tick %d", state, removed, serverTick)
	}

	game.woodcuttingSystem.RollSource = func() int { return 0 }
	for clientId := range sent {
		sent[clientId] = nil
	}
	game.update()
	for _, clientId := range []string{"client-1", "client-2", "client-3"} {
		state, removed, serverTick := woodcuttingUpdateFor(t, sent[clientId], playerId)
		if removed || state == nil || state.Phase != string(component.WoodcuttingPhaseRecovering) ||
			state.PhaseStartedTick != 2 || serverTick != 2 {
			t.Fatalf("recovery for %s = %#v removed=%v at tick %d", clientId, state, removed, serverTick)
		}
		for _, msg := range sent[clientId] {
			if msg.Metadata.Type.String() == "woodcuttingSwing" {
				t.Fatalf("woodcutting swing was projected to %s", clientId)
			}
		}
	}
}

func TestWoodcuttingCancellationReplicatesTombstoneFromEveryPhase(t *testing.T) {
	for _, phase := range []component.WoodcuttingPhase{
		component.WoodcuttingPhaseSwinging,
		component.WoodcuttingPhaseRecovering,
	} {
		t.Run(string(phase), func(t *testing.T) {
			game, treeId := setupWoodcuttingGame(t)
			sent := map[string][]message.Message{}
			game.RegisterSender(func(clientId string, msg message.Message) {
				sent[clientId] = append(sent[clientId], msg)
			})
			playerId := joinWoodcutter(t, game, "client-1", "One")
			game.woodcuttingSystem.RollSource = func() int { return 0 }
			game.HandleInteract("client-1", treeId, component.InteractionOptionChop)
			game.update()
			if phase == component.WoodcuttingPhaseRecovering {
				game.update()
			}
			assertWoodcuttingPhase(t, game, playerId, phase, game.CurrentTick())

			sent["client-1"] = nil
			game.HandleMove("client-1", 0, 0)
			game.update()
			state, removed, _ := woodcuttingUpdateFor(t, sent["client-1"], playerId)
			if state != nil || !removed {
				t.Fatalf("cancellation update = %#v removed=%v, want tombstone", state, removed)
			}
		})
	}
}

func TestRestartingWoodcuttingStartsFreshSwingingPhase(t *testing.T) {
	game, treeId := setupWoodcuttingGame(t)
	playerId := joinWoodcutter(t, game, "client-1", "One")
	events := recordGameEvents(game)
	game.woodcuttingSystem.RollSource = func() int { return 0 }

	game.HandleInteract("client-1", treeId, component.InteractionOptionChop)
	game.update()
	assertWoodcuttingPhase(t, game, playerId, component.WoodcuttingPhaseSwinging, 1)

	game.HandleInteract("client-1", treeId, component.InteractionOptionChop)
	if game.componentManager.GetEntityComponent(component.ComponentIdWoodcutting, playerId) != nil {
		t.Fatal("new interaction did not cancel the existing woodcutting state immediately")
	}
	game.update()
	assertWoodcuttingPhase(t, game, playerId, component.WoodcuttingPhaseSwinging, 2)
	if got := countGameEvents(*events, gameevent.EventIdWoodcuttingSwing); got != 0 {
		t.Fatalf("restart emitted %d swing events, want 0", got)
	}
}

func TestWoodcuttingCancelsOnMoveUnequipAndInvalidTarget(t *testing.T) {
	t.Run("move", func(t *testing.T) {
		game, treeId := setupWoodcuttingGame(t)
		playerId := joinWoodcutter(t, game, "client-1", "One")
		game.woodcuttingSystem.RollSource = func() int { return 0 }
		game.HandleInteract("client-1", treeId, component.InteractionOptionChop)
		game.update()
		game.HandleMove("client-1", 0, 0)
		if game.componentManager.GetEntityComponent(component.ComponentIdWoodcutting, playerId) != nil {
			t.Fatal("move did not cancel woodcutting")
		}
	})

	t.Run("another interaction", func(t *testing.T) {
		game, treeId := setupWoodcuttingGame(t)
		playerId := joinWoodcutter(t, game, "client-1", "One")
		game.woodcuttingSystem.RollSource = func() int { return 0 }
		game.HandleInteract("client-1", treeId, component.InteractionOptionChop)
		game.update()
		chestId, ok := firstEntityWithComponent(game, component.ComponentIdLootable)
		if !ok {
			t.Fatal("test world has no lootable chest")
		}
		game.HandleInteract("client-1", chestId, component.InteractionOptionLoot)
		if game.componentManager.GetEntityComponent(component.ComponentIdWoodcutting, playerId) != nil {
			t.Fatal("starting another interaction did not cancel woodcutting")
		}
	})

	t.Run("unequip", func(t *testing.T) {
		game, treeId := setupWoodcuttingGame(t)
		playerId := joinWoodcutter(t, game, "client-1", "One")
		game.woodcuttingSystem.RollSource = func() int { return 0 }
		game.HandleInteract("client-1", treeId, component.InteractionOptionChop)
		game.update()
		game.HandleUnequip("client-1", model.SlotWeapon)
		game.update()
		if game.componentManager.GetEntityComponent(component.ComponentIdWoodcutting, playerId) != nil {
			t.Fatal("unequipping axe did not cancel woodcutting")
		}
		assertActivityLogContains(t, game, playerId, "equipped axe")
	})

	t.Run("death", func(t *testing.T) {
		game, treeId := setupWoodcuttingGame(t)
		playerId := joinWoodcutter(t, game, "client-1", "One")
		game.woodcuttingSystem.RollSource = func() int { return 75 }
		game.HandleInteract("client-1", treeId, component.InteractionOptionChop)
		game.update()
		durability := woodcuttableFor(t, game, treeId).GetCurrentDurability()
		health := game.componentManager.GetEntityComponent(component.ComponentIdHealth, playerId).(*component.CHealth)
		health.SetCurrentHealth(0)
		game.componentManager.SetEntityComponent(playerId, health)
		game.woodcuttingSystem.Update()
		if game.componentManager.GetEntityComponent(component.ComponentIdWoodcutting, playerId) != nil {
			t.Fatal("death did not cancel woodcutting")
		}
		if woodcuttableFor(t, game, treeId).GetCurrentDurability() != durability {
			t.Fatal("dead player made another woodcutting attempt")
		}
	})

	t.Run("target removed", func(t *testing.T) {
		game, treeId := setupWoodcuttingGame(t)
		playerId := joinWoodcutter(t, game, "client-1", "One")
		game.woodcuttingSystem.RollSource = func() int { return 0 }
		game.HandleInteract("client-1", treeId, component.InteractionOptionChop)
		game.update()
		game.componentManager.RemoveEntity(treeId)
		game.update()
		if game.componentManager.GetEntityComponent(component.ComponentIdWoodcutting, playerId) != nil {
			t.Fatal("removed target did not cancel woodcutting")
		}
	})
}

func TestWoodcuttingSharesDurabilityAttributesFellingAndRegrowsExactly(t *testing.T) {
	game, treeId := setupWoodcuttingGame(t)
	events := recordGameEvents(game)
	playerOne := joinWoodcutter(t, game, "client-1", "One")
	playerTwo := joinWoodcutter(t, game, "client-2", "Two")
	game.woodcuttingSystem.RollSource = func() int { return 25 }

	tree := woodcuttableFor(t, game, treeId)
	tree.SetCurrentDurability(1)
	game.componentManager.SetEntityComponent(treeId, tree)
	game.HandleInteract("client-1", treeId, component.InteractionOptionChop)
	game.HandleInteract("client-2", treeId, component.InteractionOptionChop)
	game.update()
	game.update()

	if got := countGameEvents(*events, gameevent.EventIdWoodcuttingSwing); got != 1 {
		t.Fatalf("felling swing signals = %d, want 1", got)
	}
	if !tree.IsDepleted() || tree.GetCurrentDurability() != 0 || tree.GetRemainingRespawnTicks() != 60 {
		t.Fatalf("felled tree state = depleted %v durability %d respawn %d", tree.IsDepleted(), tree.GetCurrentDurability(), tree.GetRemainingRespawnTicks())
	}
	if got := countInventoryItemsByName(game, playerOne, "Logs") + countInventoryItemsByName(game, playerTwo, "Logs"); got != 1 {
		t.Fatalf("woodcutters received %d Logs items in total, want exactly one", got)
	}
	if game.componentManager.GetEntityComponent(component.ComponentIdWoodcutting, playerOne) != nil ||
		game.componentManager.GetEntityComponent(component.ComponentIdWoodcutting, playerTwo) != nil {
		t.Fatal("competing woodcutting states remained after felling")
	}
	assertNoInteractionOption(t, game.getInteractionOptionsForEntity(treeId), component.InteractionOptionChop)

	for tick := 1; tick < 60; tick++ {
		game.woodcuttingSystem.Update()
		if !tree.IsDepleted() {
			t.Fatalf("tree regrew after %d subsequent ticks, want 60", tick)
		}
	}
	game.woodcuttingSystem.Update()
	if tree.IsDepleted() || tree.GetCurrentDurability() != 5 || tree.GetRemainingRespawnTicks() != 0 {
		t.Fatalf("tree did not regrow at tick 60: %#v", tree.Serialize())
	}
	assertHasInteractionOption(t, game.getInteractionOptionsForEntity(treeId), component.InteractionOptionChop)
	if countInventoryItemsByName(game, playerOne, "Logs") == 1 {
		assertActivityLogContains(t, game, playerOne, "tree has regrown")
	} else {
		assertActivityLogContains(t, game, playerTwo, "tree has regrown")
	}
}

func recordGameEvents(game *Game) *[]gameevent.Event {
	events := []gameevent.Event{}
	game.RegisterGameEventHandler(gameevent.HandlerFunc(func(event gameevent.Event) {
		events = append(events, event)
	}))
	return &events
}

func countGameEvents(events []gameevent.Event, eventId string) int {
	count := 0
	for _, event := range events {
		if event.Id == eventId {
			count++
		}
	}
	return count
}

func setupWoodcuttingGame(t *testing.T) (*Game, model.EntityId) {
	t.Helper()
	gameFS := fstest.MapFS{
		"game.json": {Data: []byte(`{"formatVersion":2,"id":"woodcutting_test","world":{"chunkSize":{"x":3,"y":3}},"files":{"chunks":["chunks/test.json"],"conversations":[],"quests":[]}}`)},
		"chunks/test.json": {Data: []byte(`{
			"formatVersion":2,"id":"test","coordinate":{"x":0,"y":0},
			"terrain":["grass","grass","grass","grass","grass","grass","grass","grass","grass"],
			"heights":[0,0,0,0,0,0,0,0,0],
			"blockers":[false,false,false,false,false,false,false,false,false],"walls":[],
			"entities":[
				{"id":"player_spawn","components":{"position":{"x":0,"y":1},"playerSpawn":{}}},
				{"id":"tree_001","components":{"position":{"x":1,"y":1},"metadata":{"objectId":"tree_001","name":"Tree","type":"tree","width":1,"height":1,"blocksMovement":true},"renderable":{"type":"tree"},"woodcuttable":{"maxDurability":5,"respawnTicks":60,"yield":{"name":"Logs","type":"material","count":1}}}},
				{"id":"chest_001","components":{"position":{"x":0,"y":2},"metadata":{"objectId":"chest_001","name":"Chest","type":"chest","width":1,"height":1,"blocksMovement":true},"renderable":{"type":"chest"},"lootable":{"once":true,"items":[]}}}
			]
		}`)},
	}
	loaded, err := world.LoadFromGameFS(gameFS)
	if err != nil {
		t.Fatalf("LoadFromGameFS: %v", err)
	}
	game := NewGameWithWorld(loaded)
	game.RegisterBroadcaster(func(message.Message) {})
	game.RegisterSender(func(string, message.Message) {})
	treeId, ok := firstEntityWithComponent(game, component.ComponentIdWoodcuttable)
	if !ok {
		t.Fatal("woodcuttable tree was not loaded")
	}
	return game, treeId
}

func joinPlayer(t *testing.T, game *Game, clientId string, name string) model.EntityId {
	t.Helper()
	playerId := model.NewEntityId()
	game.HandleRegister(clientId, playerId, name)
	return playerId
}

func joinWoodcutter(t *testing.T, game *Game, clientId string, name string) model.EntityId {
	t.Helper()
	playerId := joinPlayer(t, game, clientId, name)
	equipAxe(t, game, clientId, playerId)
	return playerId
}

func equipAxe(t *testing.T, game *Game, clientId string, playerId model.EntityId) {
	t.Helper()
	inventory := game.componentManager.GetEntityComponent(component.ComponentIdInventory, playerId).(*component.CInventory)
	for _, item := range inventory.GetAllItems() {
		if item.Type == "axe" {
			game.HandleEquip(clientId, item.Id)
			return
		}
	}
	t.Fatal("starter inventory has no axe")
}

func woodcuttableFor(t *testing.T, game *Game, treeId model.EntityId) *component.CWoodcuttable {
	t.Helper()
	value := game.componentManager.GetEntityComponent(component.ComponentIdWoodcuttable, treeId)
	if value == nil {
		t.Fatal("tree has no woodcuttable component")
	}
	return value.(*component.CWoodcuttable)
}

type serializedWoodcuttingState struct {
	TargetEntityId   string `json:"targetEntityId"`
	Phase            string `json:"phase"`
	PhaseStartedTick uint64 `json:"phaseStartedTick"`
}

func woodcuttingUpdateFor(
	t *testing.T,
	messages []message.Message,
	entityId model.EntityId,
) (*serializedWoodcuttingState, bool, uint64) {
	t.Helper()
	var result *serializedWoodcuttingState
	removed := false
	var serverTick uint64
	for _, msg := range messages {
		if msg.Metadata.Type != message.MessageTypeGameUpdate {
			continue
		}
		var payload struct {
			Data struct {
				ServerTick uint64 `json:"serverTick"`
				Entities   []struct {
					EntityId    string          `json:"entityId"`
					ComponentId string          `json:"componentId"`
					Data        json.RawMessage `json:"data"`
				} `json:"entities"`
			} `json:"data"`
		}
		if err := json.Unmarshal([]byte(msg.Marshal()), &payload); err != nil {
			t.Fatal(err)
		}
		for _, update := range payload.Data.Entities {
			if update.EntityId != entityId.String() ||
				update.ComponentId != component.ComponentIdWoodcutting.String() {
				continue
			}
			serverTick = payload.Data.ServerTick
			if string(update.Data) == "null" {
				result = nil
				removed = true
				continue
			}
			state := &serializedWoodcuttingState{}
			if err := json.Unmarshal(update.Data, state); err != nil {
				t.Fatal(err)
			}
			result = state
			removed = false
		}
	}
	return result, removed, serverTick
}

func assertWoodcuttingPhase(
	t *testing.T,
	game *Game,
	entityId model.EntityId,
	wantPhase component.WoodcuttingPhase,
	wantStartedTick uint64,
) {
	t.Helper()
	value := game.componentManager.GetEntityComponent(component.ComponentIdWoodcutting, entityId)
	if value == nil {
		t.Fatal("player has no woodcutting state")
	}
	woodcutting := value.(*component.CWoodcutting)
	if woodcutting.GetPhase() != wantPhase || woodcutting.GetPhaseStartedTick() != wantStartedTick {
		t.Fatalf(
			"woodcutting = %q at tick %d, want %q at tick %d",
			woodcutting.GetPhase(),
			woodcutting.GetPhaseStartedTick(),
			wantPhase,
			wantStartedTick,
		)
	}
}

func assertActivityLogContains(t *testing.T, game *Game, playerId model.EntityId, want string) {
	t.Helper()
	log := game.componentManager.GetEntityComponent(component.ComponentIdCombatLog, playerId).(*component.CCombatLog)
	for _, entry := range log.GetEntries() {
		if strings.Contains(strings.ToLower(entry.GetText()), strings.ToLower(want)) {
			return
		}
	}
	t.Fatalf("activity log does not contain %q", want)
}

func countInventoryItemsByName(game *Game, playerId model.EntityId, name string) int {
	inventory := game.componentManager.GetEntityComponent(component.ComponentIdInventory, playerId).(*component.CInventory)
	count := 0
	for _, item := range inventory.GetAllItems() {
		if item.Name == name {
			count++
		}
	}
	return count
}

func assertHasInteractionOption(t *testing.T, options []component.InteractionOption, want component.InteractionOption) {
	t.Helper()
	for _, option := range options {
		if option == want {
			return
		}
	}
	t.Fatalf("interaction options %#v do not contain %q", options, want)
}

func assertNoInteractionOption(t *testing.T, options []component.InteractionOption, unwanted component.InteractionOption) {
	t.Helper()
	for _, option := range options {
		if option == unwanted {
			t.Fatalf("interaction options %#v contain %q", options, unwanted)
		}
	}
}
