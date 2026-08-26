package game

import (
	"strings"
	"testing"
	"testing/fstest"
	"webscape/server/game/component"
	"webscape/server/game/gameevent"
	"webscape/server/game/model"
	"webscape/server/game/world"
	"webscape/server/message"
)

func TestWoodcuttingRollOutcomesAndTwoTickCooldown(t *testing.T) {
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
			if got := tree.GetCurrentDurability(); got != 5-test.wantDamage {
				t.Fatalf("durability after immediate attempt = %d, want %d", got, 5-test.wantDamage)
			}
			if got := countGameEvents(*events, gameevent.EventIdWoodcuttingSwing); got != 1 {
				t.Fatalf("swing signals after immediate attempt = %d, want 1", got)
			}
			assertActivityLogContains(t, game, playerId, test.wantLog)

			game.update()
			if got := tree.GetCurrentDurability(); got != 5-test.wantDamage {
				t.Fatalf("durability one tick later = %d, want unchanged %d", got, 5-test.wantDamage)
			}
			if got := countGameEvents(*events, gameevent.EventIdWoodcuttingSwing); got != 1 {
				t.Fatalf("swing signals during cooldown = %d, want 1 total", got)
			}
			game.update()
			if got := tree.GetCurrentDurability(); got != 5-2*test.wantDamage {
				t.Fatalf("durability two ticks later = %d, want %d", got, 5-2*test.wantDamage)
			}
			if got := countGameEvents(*events, gameevent.EventIdWoodcuttingSwing); got != 2 {
				t.Fatalf("swing signals on second attempt = %d, want 2 total", got)
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
	positionOne := game.componentManager.GetEntityComponent(component.ComponentIdPosition, playerOne).(*component.CPosition)
	positionTwo := game.componentManager.GetEntityComponent(component.ComponentIdPosition, playerTwo).(*component.CPosition)
	stateOne := component.NewCWoodcutting(treeId, positionOne.GetPosition())
	stateOne.SetCooldownRemaining(2)
	game.componentManager.SetEntityComponent(playerOne, stateOne)
	game.componentManager.SetEntityComponent(playerTwo, component.NewCWoodcutting(treeId, positionTwo.GetPosition()))

	game.woodcuttingSystem.Update()

	if got := countGameEvents(*events, gameevent.EventIdWoodcuttingSwing); got != 1 {
		t.Fatalf("felling swing signals = %d, want 1", got)
	}
	if !tree.IsDepleted() || tree.GetCurrentDurability() != 0 || tree.GetRemainingRespawnTicks() != 60 {
		t.Fatalf("felled tree state = depleted %v durability %d respawn %d", tree.IsDepleted(), tree.GetCurrentDurability(), tree.GetRemainingRespawnTicks())
	}
	if countInventoryItemsByName(game, playerOne, "Logs") != 0 {
		t.Fatal("non-felling player received Logs")
	}
	if countInventoryItemsByName(game, playerTwo, "Logs") != 1 {
		t.Fatal("felling player did not receive exactly one Logs item")
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
	assertActivityLogContains(t, game, playerTwo, "tree has regrown")
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
