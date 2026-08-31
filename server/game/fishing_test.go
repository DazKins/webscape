package game

import (
	"encoding/json"
	"strings"
	"testing"
	"testing/fstest"
	"webscape/server/game/component"
	"webscape/server/game/gameevent"
	"webscape/server/game/model"
	gamesystem "webscape/server/game/system"
	"webscape/server/game/world"
	"webscape/server/math"
	"webscape/server/message"
	"webscape/server/util"
)

func TestFishingPhaseTimelineMissesAndCatches(t *testing.T) {
	game, spotId := setupFishingGame(t, false)
	events := recordGameEvents(game)
	playerId := joinFisher(t, game, "client-1", "One")
	rolls := []int{99, 0, 0}
	rollCount := 0
	game.fishingSystem.RollSource = func() int {
		rollCount++
		roll := rolls[0]
		rolls = rolls[1:]
		return roll
	}

	assertHasInteractionOption(t, game.getInteractionOptionsForEntity(spotId), component.InteractionOptionFish)
	game.HandleInteract("client-1", spotId, component.InteractionOptionFish)

	game.update()
	assertFishingPhase(t, game, playerId, component.FishingPhaseCasting, 1)
	if rollCount != 0 {
		t.Fatalf("casting tick rolls = %d, want 0", rollCount)
	}

	game.update()
	assertFishingPhase(t, game, playerId, component.FishingPhaseWaiting, 2)
	if rollCount != 0 {
		t.Fatalf("waiting transition tick rolls = %d, want 0", rollCount)
	}

	missMessages := []message.Message{}
	game.RegisterSender(func(_ string, msg message.Message) {
		missMessages = append(missMessages, msg)
	})
	game.update()
	assertFishingPhase(t, game, playerId, component.FishingPhaseWaiting, 2)
	if rollCount != 1 || countInventoryItemsByName(game, playerId, "Raw Fish") != 0 {
		t.Fatalf("first waiting roll count = %d, fish = %d", rollCount, countInventoryItemsByName(game, playerId, "Raw Fish"))
	}
	if got := countGameEvents(*events, gameevent.EventIdFishingCatch); got != 0 {
		t.Fatalf("miss emitted %d fishing catch events", got)
	}
	if indexOfMessageType(messageTypes(missMessages), message.MessageTypeGameUpdate) >= 0 {
		t.Fatalf("miss emitted a game update: %v", messageTypes(missMessages))
	}
	assertActivityLogDoesNotContain(t, game, playerId, "catch nothing")

	game.update()
	assertFishingPhase(t, game, playerId, component.FishingPhaseReeling, 4)
	if rollCount != 2 || countInventoryItemsByName(game, playerId, "Raw Fish") != 1 {
		t.Fatalf("catch tick rolls = %d, fish = %d", rollCount, countInventoryItemsByName(game, playerId, "Raw Fish"))
	}

	game.update()
	assertFishingPhase(t, game, playerId, component.FishingPhaseWaiting, 5)
	if rollCount != 2 {
		t.Fatalf("reeling transition tick rolls = %d, want 2", rollCount)
	}

	game.update()
	assertFishingPhase(t, game, playerId, component.FishingPhaseReeling, 6)
	if rollCount != 3 || countInventoryItemsByName(game, playerId, "Raw Fish") != 2 {
		t.Fatalf("resumed waiting roll count = %d, fish = %d", rollCount, countInventoryItemsByName(game, playerId, "Raw Fish"))
	}
	if got := countGameEvents(*events, gameevent.EventIdFishingCatch); got != 2 {
		t.Fatalf("fishing catches = %d, want 2", got)
	}
	assertActivityLogContains(t, game, playerId, "catch 1 Raw Fish")
}

func TestFishingInteractionStartsAfterPathingFromAfar(t *testing.T) {
	game, spotId := setupDistantFishingGame(t)
	sent := []message.Message{}
	game.RegisterSender(func(_ string, msg message.Message) {
		sent = append(sent, msg)
	})
	playerId := joinFisher(t, game, "client-1", "One")
	rollCount := 0
	game.fishingSystem.RollSource = func() int {
		rollCount++
		return 99
	}
	sent = nil

	game.HandleInteract("client-1", spotId, component.InteractionOptionFish)
	for tick := uint64(1); tick <= 3; tick++ {
		game.update()
		assertEntityStateValid(t, game, playerId)
		if game.componentManager.GetEntityComponent(component.ComponentIdFishing, playerId) != nil {
			t.Fatalf("fishing started on movement tick %d", tick)
		}
		assertLocomotion(t, game, playerId, component.LocomotionPhaseMoving, 1)
	}
	if game.componentManager.GetEntityComponent(component.ComponentIdInteracting, playerId) == nil {
		t.Fatal("interaction was not retained after the final pathing step")
	}

	sent = nil
	game.update()
	assertEntityStateValid(t, game, playerId)
	assertFishingPhase(t, game, playerId, component.FishingPhaseCasting, 4)
	assertLocomotion(t, game, playerId, component.LocomotionPhaseIdle, 4)
	if game.componentManager.GetEntityComponent(component.ComponentIdInteracting, playerId) != nil {
		t.Fatal("interaction remained after fishing started")
	}
	if rollCount != 0 {
		t.Fatalf("casting tick rolls = %d, want 0", rollCount)
	}

	fishingState, fishingTick := fishingUpdateFor(t, sent, playerId)
	locomotionState, locomotionTick := locomotionUpdateFor(t, sent, playerId)
	if fishingState == nil || fishingState.Phase != string(component.FishingPhaseCasting) ||
		fishingState.PhaseStartedTick != 4 || fishingTick != 4 {
		t.Fatalf("replicated fishing state = %#v at tick %d", fishingState, fishingTick)
	}
	if locomotionState == nil || locomotionState.Phase != string(component.LocomotionPhaseIdle) ||
		locomotionState.PhaseStartedTick != 4 || locomotionTick != 4 {
		t.Fatalf("replicated locomotion state = %#v at tick %d", locomotionState, locomotionTick)
	}
}

func assertEntityStateValid(t *testing.T, game *Game, entityId model.EntityId) {
	t.Helper()
	if err := gamesystem.ValidateEntityState(game.componentManager, entityId); err != nil {
		t.Fatal(err)
	}
}

func TestFishingStartRejectsMovementFromCurrentTick(t *testing.T) {
	game, spotId := setupFishingGame(t, false)
	playerId := joinFisher(t, game, "client-1", "One")
	locomotion := game.componentManager.GetEntityComponent(
		component.ComponentIdLocomotion,
		playerId,
	).(*component.CLocomotion)
	locomotion.MarkMoving(game.CurrentTick())
	game.componentManager.SetEntityComponent(playerId, locomotion)

	if game.fishingSystem.StartFishingFor(playerId, spotId) {
		t.Fatal("fishing started after movement in the current tick")
	}
	if game.componentManager.GetEntityComponent(component.ComponentIdFishing, playerId) != nil {
		t.Fatal("rejected fishing start created fishing state")
	}
	assertLocomotion(t, game, playerId, component.LocomotionPhaseMoving, 0)
}

func setupDistantFishingGame(t *testing.T) (*Game, model.EntityId) {
	t.Helper()
	loaded, err := world.LoadFromGameFS(fstest.MapFS{
		"game.json": {Data: []byte(`{"formatVersion":2,"id":"distant_fishing_test","world":{"chunkSize":{"x":5,"y":3}},"files":{"chunks":["chunks/test.json"],"conversations":[],"quests":[]}}`)},
		"chunks/test.json": {Data: []byte(`{
			"formatVersion":2,"id":"test","coordinate":{"x":0,"y":0},
			"terrain":["grass","grass","grass","grass","grass","grass","grass","grass","grass","water","grass","grass","grass","grass","grass"],
			"heights":[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0],
			"blockers":[false,false,false,false,false,false,false,false,false,true,false,false,false,false,false],"walls":[],
			"entities":[
				{"id":"player_spawn","components":{"position":{"x":0,"y":1},"playerSpawn":{}}},
				{"id":"fishing_spot_001","components":{"position":{"x":4,"y":1},"metadata":{"objectId":"fishing_spot_001","name":"Fishing Spot","type":"fishingSpot","width":1,"height":1,"blocksMovement":true},"renderable":{"type":"fishingSpot"},"fishable":{"catchChancePercent":5,"yield":{"name":"Raw Fish","type":"fish","count":1}}}}
			]
		}`)},
	})
	if err != nil {
		t.Fatalf("LoadFromGameFS: %v", err)
	}
	game := NewGameWithWorld(loaded)
	game.RegisterBroadcaster(func(message.Message) {})
	game.RegisterSender(func(string, message.Message) {})
	spotId, ok := firstEntityWithComponent(game, component.ComponentIdFishable)
	if !ok {
		t.Fatal("fishable spot was not loaded")
	}
	return game, spotId
}

func TestFishingRequiresRodAndCapacity(t *testing.T) {
	game, spotId := setupFishingGame(t, false)
	playerId := joinPlayer(t, game, "client-1", "One")
	game.HandleInteract("client-1", spotId, component.InteractionOptionFish)
	game.update()
	if game.componentManager.GetEntityComponent(component.ComponentIdFishing, playerId) != nil {
		t.Fatal("fishing started without an equipped rod")
	}
	assertActivityLogContains(t, game, playerId, "equipped fishing rod")

	equipFishingRod(t, game, "client-1", playerId)
	inventory := game.componentManager.GetEntityComponent(component.ComponentIdInventory, playerId).(*component.CInventory)
	for !inventory.IsFull() {
		inventory.AddItem(model.NewItem("Filler", "test"))
	}
	game.componentManager.SetEntityComponent(playerId, inventory)
	game.HandleInteract("client-1", spotId, component.InteractionOptionFish)
	game.update()
	if game.componentManager.GetEntityComponent(component.ComponentIdFishing, playerId) != nil {
		t.Fatal("fishing started with a full inventory")
	}
	if countInventoryItemsByName(game, playerId, "Raw Fish") != 0 {
		t.Fatal("full inventory received a fish")
	}
	assertActivityLogContains(t, game, playerId, "inventory is full")
}

func TestFishingAllowsMultiplePlayersOnOneSpot(t *testing.T) {
	game, spotId := setupFishingGame(t, false)
	playerOne := joinFisher(t, game, "client-1", "One")
	playerTwo := joinFisher(t, game, "client-2", "Two")
	game.fishingSystem.RollSource = func() int { return 0 }
	game.HandleInteract("client-1", spotId, component.InteractionOptionFish)
	game.HandleInteract("client-2", spotId, component.InteractionOptionFish)
	game.update()
	game.update()
	game.update()

	if countInventoryItemsByName(game, playerOne, "Raw Fish") != 1 ||
		countInventoryItemsByName(game, playerTwo, "Raw Fish") != 1 {
		t.Fatal("both players did not catch from the shared spot")
	}
	if game.componentManager.GetEntityComponent(component.ComponentIdFishing, playerOne) == nil ||
		game.componentManager.GetEntityComponent(component.ComponentIdFishing, playerTwo) == nil {
		t.Fatal("shared fishing spot canceled one player's activity")
	}
}

func TestFishingCatchThatFillsInventoryAwardsThenCancels(t *testing.T) {
	game, spotId := setupFishingGame(t, false)
	playerId := joinFisher(t, game, "client-1", "One")
	inventory := game.componentManager.GetEntityComponent(component.ComponentIdInventory, playerId).(*component.CInventory)
	for inventory.AvailableSlots() > 1 {
		inventory.AddItem(model.NewItem("Filler", "test"))
	}
	game.componentManager.SetEntityComponent(playerId, inventory)
	game.fishingSystem.RollSource = func() int { return 0 }
	game.HandleInteract("client-1", spotId, component.InteractionOptionFish)
	game.update()
	game.update()
	game.update()

	if countInventoryItemsByName(game, playerId, "Raw Fish") != 1 || !inventory.IsFull() {
		t.Fatal("final available slot did not receive the catch")
	}
	stateValue := game.componentManager.GetEntityComponent(component.ComponentIdFishing, playerId)
	if stateValue == nil || stateValue.(*component.CFishing).GetPhase() != component.FishingPhaseReeling {
		t.Fatal("reeling state was not retained for a full tick after the final catch")
	}
	game.update()
	if game.componentManager.GetEntityComponent(component.ComponentIdFishing, playerId) != nil {
		t.Fatal("fishing did not cancel on the tick after the catch filled inventory")
	}
	assertActivityLogContains(t, game, playerId, "inventory is full")
}

func TestFishingCancellationConditions(t *testing.T) {
	tests := []struct {
		name   string
		cancel func(t *testing.T, game *Game, playerId model.EntityId, spotId model.EntityId)
	}{
		{name: "movement command", cancel: func(t *testing.T, game *Game, _ model.EntityId, _ model.EntityId) {
			game.HandleMove("client-1", 0, 0)
		}},
		{name: "another interaction", cancel: func(t *testing.T, game *Game, _ model.EntityId, _ model.EntityId) {
			chestId, _ := firstEntityWithComponent(game, component.ComponentIdLootable)
			game.HandleInteract("client-1", chestId, component.InteractionOptionLoot)
		}},
		{name: "combat", cancel: func(t *testing.T, game *Game, playerId model.EntityId, spotId model.EntityId) {
			game.componentManager.SetEntityComponent(playerId, component.NewCCombatState(spotId))
			game.fishingSystem.Update()
		}},
		{name: "pathing", cancel: func(t *testing.T, game *Game, playerId model.EntityId, spotId model.EntityId) {
			game.componentManager.SetEntityComponent(playerId, component.NewCPathing(component.PathingTarget{
				EntityId: util.OptionalSome(spotId),
			}))
			game.fishingSystem.Update()
		}},
		{name: "rod replacement", cancel: func(t *testing.T, game *Game, playerId model.EntityId, _ model.EntityId) {
			inventory := game.componentManager.GetEntityComponent(component.ComponentIdInventory, playerId).(*component.CInventory)
			for _, item := range inventory.GetAllItems() {
				if item.Name == "Iron Sword" {
					game.HandleEquip("client-1", item.Id)
					game.fishingSystem.Update()
					return
				}
			}
			t.Fatal("starter inventory has no sword")
		}},
		{name: "death", cancel: func(t *testing.T, game *Game, playerId model.EntityId, _ model.EntityId) {
			health := game.componentManager.GetEntityComponent(component.ComponentIdHealth, playerId).(*component.CHealth)
			health.SetCurrentHealth(0)
			game.componentManager.SetEntityComponent(playerId, health)
			game.fishingSystem.Update()
		}},
		{name: "inventory fills", cancel: func(t *testing.T, game *Game, playerId model.EntityId, _ model.EntityId) {
			inventory := game.componentManager.GetEntityComponent(component.ComponentIdInventory, playerId).(*component.CInventory)
			for !inventory.IsFull() {
				inventory.AddItem(model.NewItem("Filler", "test"))
			}
			game.componentManager.SetEntityComponent(playerId, inventory)
			game.update()
		}},
		{name: "target removed", cancel: func(t *testing.T, game *Game, _ model.EntityId, spotId model.EntityId) {
			game.componentManager.RemoveEntity(spotId)
			game.fishingSystem.Update()
		}},
		{name: "player position changes", cancel: func(t *testing.T, game *Game, playerId model.EntityId, _ model.EntityId) {
			position := game.componentManager.GetEntityComponent(component.ComponentIdPosition, playerId).(*component.CPosition)
			position.SetPosition(math.Vec2{X: 0, Y: 0})
			game.componentManager.SetEntityComponent(playerId, position)
			game.fishingSystem.Update()
		}},
		{name: "adjacency lost", cancel: func(t *testing.T, game *Game, _ model.EntityId, spotId model.EntityId) {
			position := game.componentManager.GetEntityComponent(component.ComponentIdPosition, spotId).(*component.CPosition)
			position.SetPosition(math.Vec2{X: 2, Y: 2})
			game.componentManager.SetEntityComponent(spotId, position)
			game.fishingSystem.Update()
		}},
	}

	phases := []struct {
		name         string
		advanceTicks int
	}{
		{name: "casting", advanceTicks: 1},
		{name: "waiting", advanceTicks: 2},
		{name: "reeling", advanceTicks: 3},
	}
	for _, phase := range phases {
		for _, test := range tests {
			t.Run(phase.name+"/"+test.name, func(t *testing.T) {
				game, spotId := setupFishingGame(t, false)
				playerId := joinFisher(t, game, "client-1", "One")
				game.fishingSystem.RollSource = func() int { return 0 }
				game.HandleInteract("client-1", spotId, component.InteractionOptionFish)
				for tick := 0; tick < phase.advanceTicks; tick++ {
					game.update()
				}
				if game.componentManager.GetEntityComponent(component.ComponentIdFishing, playerId) == nil {
					t.Fatal("fishing was not active before cancellation")
				}
				cancellationMessages := []message.Message{}
				game.RegisterSender(func(_ string, msg message.Message) {
					cancellationMessages = append(cancellationMessages, msg)
				})
				test.cancel(t, game, playerId, spotId)
				if game.componentManager.GetEntityComponent(component.ComponentIdFishing, playerId) != nil {
					t.Fatal("fishing activity was not canceled")
				}
				game.syncClient("client-1")
				if !hasFishingTombstoneAtTick(t, cancellationMessages, playerId, game.CurrentTick()) {
					t.Fatalf("messages = %v, want fishing tombstone", messageTypes(cancellationMessages))
				}
			})
		}
	}
}

func TestFishingCatchEventAdvancesQuestByAwardedCount(t *testing.T) {
	game, spotId := setupFishingGame(t, true)
	events := recordGameEvents(game)
	game.componentManager.SetEntityComponent(spotId, component.NewCFishable(5, component.LootItem{
		Name: "Raw Fish", Type: "fish", Count: 2,
	}))
	playerId := joinFisher(t, game, "client-1", "One")
	game.fishingSystem.RollSource = func() int { return 0 }
	game.HandleInteract("client-1", spotId, component.InteractionOptionFish)
	game.update()
	game.update()
	game.update()
	questLog := game.componentManager.GetEntityComponent(component.ComponentIdQuestLog, playerId).(*component.CQuestLog)
	if questLog.IsCompleted("catch_fish") {
		t.Fatal("quest completed after only one catch")
	}
	game.update()
	game.update()
	if !questLog.IsCompleted("catch_fish") {
		t.Fatal("fishing:catch events did not complete the two-catch quest")
	}
	if countInventoryItemsByName(game, playerId, "Raw Fish") != 4 {
		t.Fatal("two catches did not award four authored fish items")
	}
	for _, event := range *events {
		if event.Id == gameevent.EventIdFishingCatch && event.Count != 2 {
			t.Fatalf("fishing catch event count = %d, want awarded count 2", event.Count)
		}
	}
}

func TestFishingStateReplicatesPhaseAndServerTick(t *testing.T) {
	game, spotId := setupFishingGame(t, false)
	sent := map[string][]message.Message{}
	game.RegisterSender(func(clientID string, msg message.Message) {
		sent[clientID] = append(sent[clientID], msg)
	})
	playerId := joinFisher(t, game, "client-1", "One")
	joinPlayer(t, game, "client-2", "Two")
	sent = map[string][]message.Message{}
	game.HandleInteract("client-1", spotId, component.InteractionOptionFish)
	game.update()

	for _, clientId := range []string{"client-1", "client-2"} {
		var fishingUpdate *struct {
			TargetEntityId   string `json:"targetEntityId"`
			Phase            string `json:"phase"`
			PhaseStartedTick uint64 `json:"phaseStartedTick"`
		}
		var serverTick uint64
		for _, msg := range sent[clientId] {
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
			serverTick = payload.Data.ServerTick
			for _, update := range payload.Data.Entities {
				if update.EntityId != playerId.String() || update.ComponentId != component.ComponentIdFishing.String() {
					continue
				}
				var state struct {
					TargetEntityId   string `json:"targetEntityId"`
					Phase            string `json:"phase"`
					PhaseStartedTick uint64 `json:"phaseStartedTick"`
				}
				if err := json.Unmarshal(update.Data, &state); err != nil {
					t.Fatal(err)
				}
				fishingUpdate = &state
			}
		}
		if fishingUpdate == nil {
			t.Fatalf("messages for %s = %v, want replicated fishing state", clientId, messageTypes(sent[clientId]))
		}
		if serverTick != 1 || fishingUpdate.TargetEntityId != spotId.String() ||
			fishingUpdate.Phase != string(component.FishingPhaseCasting) || fishingUpdate.PhaseStartedTick != 1 {
			t.Fatalf("fishing state for %s = %#v", clientId, fishingUpdate)
		}
	}
}

func TestFishingRestartAtSameSpotUsesLaterCastingTick(t *testing.T) {
	game, spotId := setupFishingGame(t, false)
	playerId := joinFisher(t, game, "client-1", "One")
	game.HandleInteract("client-1", spotId, component.InteractionOptionFish)
	game.update()
	assertFishingPhase(t, game, playerId, component.FishingPhaseCasting, 1)

	game.HandleInteract("client-1", spotId, component.InteractionOptionFish)
	if game.componentManager.GetEntityComponent(component.ComponentIdFishing, playerId) != nil {
		t.Fatal("new interaction did not cancel the existing fishing state immediately")
	}
	game.update()
	assertFishingPhase(t, game, playerId, component.FishingPhaseCasting, 2)
}

func TestLateInterestedClientReceivesCurrentFishingPhase(t *testing.T) {
	game, spotId := setupFishingGame(t, false)
	playerId := joinFisher(t, game, "client-1", "One")
	game.fishingSystem.RollSource = func() int { return 0 }
	game.HandleInteract("client-1", spotId, component.InteractionOptionFish)
	game.update()
	game.update()
	game.update()
	assertFishingPhase(t, game, playerId, component.FishingPhaseReeling, 3)

	sent := []message.Message{}
	game.RegisterSender(func(clientID string, msg message.Message) {
		if clientID == "client-2" {
			sent = append(sent, msg)
		}
	})
	joinPlayer(t, game, "client-2", "Two")
	state, serverTick := fishingUpdateFor(t, sent, playerId)
	if state == nil || state.Phase != string(component.FishingPhaseReeling) ||
		state.PhaseStartedTick != 3 || serverTick != 3 {
		t.Fatalf("late fishing snapshot = %#v at server tick %d", state, serverTick)
	}
}

func TestServerTickStartsAtZeroAndIncrementsOncePerUpdate(t *testing.T) {
	game, spotId := setupFishingGame(t, false)
	sent := []message.Message{}
	game.RegisterSender(func(_ string, msg message.Message) { sent = append(sent, msg) })
	joinFisher(t, game, "client-1", "One")
	if game.CurrentTick() != 0 || !hasGameUpdateAtTick(t, sent, 0) {
		t.Fatalf("initial tick = %d, messages = %v", game.CurrentTick(), messageTypes(sent))
	}

	sent = sent[:0]
	game.HandleInteract("client-1", spotId, component.InteractionOptionFish)
	game.update()
	if game.CurrentTick() != 1 || !hasGameUpdateAtTick(t, sent, 1) {
		t.Fatalf("after one update tick = %d, messages = %v", game.CurrentTick(), messageTypes(sent))
	}

	sent = sent[:0]
	game.update()
	if game.CurrentTick() != 2 || !hasGameUpdateAtTick(t, sent, 2) {
		t.Fatalf("after two updates tick = %d, messages = %v", game.CurrentTick(), messageTypes(sent))
	}
}

func assertActivityLogDoesNotContain(t *testing.T, game *Game, playerId model.EntityId, unwanted string) {
	t.Helper()
	log := game.componentManager.GetEntityComponent(component.ComponentIdCombatLog, playerId).(*component.CCombatLog)
	for _, entry := range log.GetEntries() {
		if strings.Contains(strings.ToLower(entry.GetText()), strings.ToLower(unwanted)) {
			t.Fatalf("activity log unexpectedly contains %q", unwanted)
		}
	}
}

type serializedFishingState struct {
	TargetEntityId   string `json:"targetEntityId"`
	Phase            string `json:"phase"`
	PhaseStartedTick uint64 `json:"phaseStartedTick"`
}

func fishingUpdateFor(
	t *testing.T,
	messages []message.Message,
	playerId model.EntityId,
) (*serializedFishingState, uint64) {
	t.Helper()
	var result *serializedFishingState
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
			if update.EntityId != playerId.String() ||
				update.ComponentId != component.ComponentIdFishing.String() ||
				string(update.Data) == "null" {
				continue
			}
			state := &serializedFishingState{}
			if err := json.Unmarshal(update.Data, state); err != nil {
				t.Fatal(err)
			}
			result = state
			serverTick = payload.Data.ServerTick
		}
	}
	return result, serverTick
}

func hasFishingTombstoneAtTick(
	t *testing.T,
	messages []message.Message,
	playerId model.EntityId,
	wantTick uint64,
) bool {
	t.Helper()
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
			if update.EntityId == playerId.String() &&
				update.ComponentId == component.ComponentIdFishing.String() &&
				string(update.Data) == "null" && payload.Data.ServerTick == wantTick {
				return true
			}
		}
	}
	return false
}

func hasGameUpdateAtTick(t *testing.T, messages []message.Message, wantTick uint64) bool {
	t.Helper()
	for _, msg := range messages {
		if msg.Metadata.Type != message.MessageTypeGameUpdate {
			continue
		}
		var payload struct {
			Data struct {
				ServerTick uint64 `json:"serverTick"`
			} `json:"data"`
		}
		if err := json.Unmarshal([]byte(msg.Marshal()), &payload); err != nil {
			t.Fatal(err)
		}
		if payload.Data.ServerTick == wantTick {
			return true
		}
	}
	return false
}

func assertFishingPhase(
	t *testing.T,
	game *Game,
	playerId model.EntityId,
	wantPhase component.FishingPhase,
	wantStartedTick uint64,
) {
	t.Helper()
	value := game.componentManager.GetEntityComponent(component.ComponentIdFishing, playerId)
	if value == nil {
		t.Fatalf("fishing state is missing; want phase %q", wantPhase)
	}
	state := value.(*component.CFishing)
	if state.GetPhase() != wantPhase || state.GetPhaseStartedTick() != wantStartedTick {
		t.Fatalf(
			"fishing phase = %q at tick %d; want %q at tick %d",
			state.GetPhase(),
			state.GetPhaseStartedTick(),
			wantPhase,
			wantStartedTick,
		)
	}
}

func setupFishingGame(t *testing.T, withQuest bool) (*Game, model.EntityId) {
	t.Helper()
	questPaths := `[]`
	gameFS := fstest.MapFS{}
	if withQuest {
		questPaths = `["quests/fishing.json"]`
		gameFS["quests/fishing.json"] = &fstest.MapFile{Data: []byte(`{
			"formatVersion":2,"id":"fishing_quests","quests":[{
				"id":"catch_fish","startEventId":"fishing:catch",
				"steps":[{"id":"catch_four","description":"Catch four fish.","requirement":{"eventId":"fishing:catch","count":4}}],
				"rewards":{"items":[{"name":"Fishing Badge","type":"quest","count":1}]}
			}]
		}`)}
	}
	gameFS["game.json"] = &fstest.MapFile{Data: []byte(`{"formatVersion":2,"id":"fishing_test","world":{"chunkSize":{"x":3,"y":3}},"files":{"chunks":["chunks/test.json"],"conversations":[],"quests":` + questPaths + `}}`)}
	gameFS["chunks/test.json"] = &fstest.MapFile{Data: []byte(`{
		"formatVersion":2,"id":"test","coordinate":{"x":0,"y":0},
		"terrain":["grass","water","grass","grass","water","grass","grass","grass","grass"],
		"heights":[0,0,0,0,0,0,0,0,0],
		"blockers":[false,true,false,false,true,false,false,false,false],"walls":[],
		"entities":[
			{"id":"player_spawn","components":{"position":{"x":0,"y":1},"playerSpawn":{}}},
			{"id":"fishing_spot_001","components":{"position":{"x":1,"y":1},"metadata":{"objectId":"fishing_spot_001","name":"Fishing Spot","type":"fishingSpot","width":1,"height":1,"blocksMovement":true},"renderable":{"type":"fishingSpot"},"fishable":{"catchChancePercent":5,"yield":{"name":"Raw Fish","type":"fish","count":1}}}},
			{"id":"chest_001","components":{"position":{"x":0,"y":2},"metadata":{"objectId":"chest_001","name":"Chest","type":"chest","width":1,"height":1,"blocksMovement":true},"renderable":{"type":"chest"},"lootable":{"once":true,"items":[]}}}
		]
	}`)}
	loaded, err := world.LoadFromGameFS(gameFS)
	if err != nil {
		t.Fatalf("LoadFromGameFS: %v", err)
	}
	game := NewGameWithWorld(loaded)
	game.RegisterBroadcaster(func(message.Message) {})
	game.RegisterSender(func(string, message.Message) {})
	spotId, ok := firstEntityWithComponent(game, component.ComponentIdFishable)
	if !ok {
		t.Fatal("fishable spot was not loaded")
	}
	return game, spotId
}

func joinFisher(t *testing.T, game *Game, clientId string, name string) model.EntityId {
	t.Helper()
	playerId := joinPlayer(t, game, clientId, name)
	equipFishingRod(t, game, clientId, playerId)
	return playerId
}

func equipFishingRod(t *testing.T, game *Game, clientId string, playerId model.EntityId) {
	t.Helper()
	inventory := game.componentManager.GetEntityComponent(component.ComponentIdInventory, playerId).(*component.CInventory)
	for _, item := range inventory.GetAllItems() {
		if item.Type == "fishingRod" {
			game.HandleEquip(clientId, item.Id)
			return
		}
	}
	t.Fatal("starter inventory has no fishing rod")
}
