package game

import (
	"encoding/json"
	"strconv"
	"testing"
	"testing/fstest"
	"webscape/server/game/component"
	"webscape/server/game/gameevent"
	"webscape/server/game/model"
	"webscape/server/game/world"
	"webscape/server/message"
)

func TestQuestProgressAdvancesFromGenericEvents(t *testing.T) {
	testWorld, err := world.LoadFromGameFS(fstest.MapFS{
		"game.json": {
			Data: []byte(`{
				"formatVersion": 1,
				"id": "test_game",
				"files": {
					"maps": ["maps/test.json"],
					"conversations": ["conversations/guide.json"],
					"quests": ["quests/tutorial.json"]
				}
			}`),
		},
		"maps/test.json": {
			Data: []byte(`{
				"formatVersion": 1,
				"id": "test",
				"size": { "x": 2, "y": 1 },
				"terrain": ["grass", "road"],
				"heights": [0, 0],
				"entities": [
					{
						"id": "player_spawn",
						"components": {
							"position": { "x": 0, "y": 0 },
							"playerSpawn": {}
						}
					},
					{
						"id": "guide",
						"components": {
							"position": { "x": 1, "y": 0 },
							"conversation": { "conversationId": "guide" }
						}
					}
				]
			}`),
		},
		"conversations/guide.json": {
			Data: []byte(`{
				"formatVersion": 1,
				"id": "guide_doc",
				"conversations": [
					{
						"id": "guide",
						"startNodeId": "start",
						"nodes": [
							{
								"id": "start",
								"messages": [{ "text": "Help me." }],
								"endConversation": true
							}
						]
					}
				]
			}`),
		},
		"quests/tutorial.json": {
			Data: []byte(`{
				"formatVersion": 2,
				"id": "tutorial_quests",
				"quests": [
					{
						"id": "first_errand",
						"startEventId": "conversation:node:guide:start",
						"steps": [
							{
								"id": "talk",
								"description": "Talk.",
								"requirement": { "eventId": "conversation:node:guide:start", "count": 1 }
							},
							{
								"id": "kill_rats",
								"description": "Kill rats.",
								"requirement": { "eventId": "kill:entity:rat", "count": 2 }
							},
							{
								"id": "collect_scroll",
								"description": "Collect a scroll.",
								"requirement": { "eventId": "collect:name:ancient_scroll", "count": 1 }
							}
						],
						"rewards": {
							"items": [
								{ "name": "Guide Token", "type": "quest", "count": 1 }
							]
						}
					},
					{
						"id": "celebrate_errand",
						"startEventId": "quest:completed:first_errand",
						"steps": [
							{
								"id": "finish",
								"description": "Celebrate finishing the errand.",
								"requirement": { "eventId": "quest:completed:first_errand", "count": 1 }
							}
						],
						"rewards": {
							"items": [
								{ "name": "Celebration Token", "type": "quest", "count": 1 }
							]
						}
					}
				]
			}`),
		},
	})
	if err != nil {
		t.Fatalf("LoadFromGameFS returned error: %v", err)
	}

	game := NewGameWithWorld(testWorld)
	game.RegisterBroadcaster(func(message.Message) {})
	game.RegisterSender(func(string, message.Message) {})

	playerEntityId := model.NewEntityId()
	game.HandleJoin("client-1", playerEntityId, "player")

	npcEntityId, ok := firstEntityWithComponent(game, component.ComponentIdConversation)
	if !ok {
		t.Fatal("no NPC entity has CConversation")
	}

	game.HandleInteract("client-1", npcEntityId, component.InteractionOptionTalk)
	game.update()
	assertQuestProgress(t, game, playerEntityId, "first_errand", "kill_rats", 1, 0, false)

	game.EmitGameEvent(gameevent.New("kill:entity:goblin", playerEntityId))
	assertQuestProgress(t, game, playerEntityId, "first_errand", "kill_rats", 1, 0, false)

	game.EmitGameEvent(gameevent.New("kill:entity:rat", playerEntityId))
	assertQuestProgress(t, game, playerEntityId, "first_errand", "kill_rats", 1, 1, false)

	game.EmitGameEvent(gameevent.New("kill:entity:rat", playerEntityId))
	assertQuestProgress(t, game, playerEntityId, "first_errand", "collect_scroll", 2, 0, false)

	game.AddItemToPlayerInventory(playerEntityId, model.CreateAncientScroll())
	assertQuestProgress(t, game, playerEntityId, "first_errand", "", 0, 0, true)
	assertQuestProgress(t, game, playerEntityId, "celebrate_errand", "", 0, 0, true)

	questLog := game.componentManager.GetEntityComponent(component.ComponentIdQuestLog, playerEntityId).(*component.CQuestLog)
	completed := questLog.GetCompletedQuests()
	if len(completed) != 2 || completed[0].QuestId != "first_errand" || completed[1].QuestId != "celebrate_errand" {
		t.Fatalf("completed quests = %#v, want first_errand then celebrate_errand", completed)
	}

	inventory := game.componentManager.GetEntityComponent(component.ComponentIdInventory, playerEntityId).(*component.CInventory)
	if !inventoryContains(inventory, "Guide Token", "quest") {
		t.Fatal("quest reward Guide Token was not added to inventory")
	}
}

func TestQuestCompletionSendsCompletionMessage(t *testing.T) {
	testWorld := loadRewardQuestWorld(t, 1)
	game := NewGameWithWorld(testWorld)
	game.RegisterBroadcaster(func(message.Message) {})
	sent := []message.Message{}
	game.RegisterSender(func(clientID string, msg message.Message) {
		if clientID == "client-1" {
			sent = append(sent, msg)
		}
	})

	playerEntityId := model.NewEntityId()
	game.HandleJoin("client-1", playerEntityId, "player")
	game.EmitGameEvent(gameevent.New("finish:quest", playerEntityId))

	completion := firstQuestCompletedPayload(t, sent)
	if completion.QuestId != "reward_test" {
		t.Fatalf("questCompleted questId = %q, want reward_test", completion.QuestId)
	}
	if completion.CompletedStepSummary != "Finish the quest." {
		t.Fatalf("questCompleted step summary = %q, want Finish the quest.", completion.CompletedStepSummary)
	}
	if len(completion.Rewards) != 1 {
		t.Fatalf("questCompleted rewards = %#v, want one reward", completion.Rewards)
	}
	if completion.Rewards[0].Name != "Reward Gem" ||
		completion.Rewards[0].Count != 1 ||
		completion.Rewards[0].Delivery != message.QuestRewardDeliveryInventory {
		t.Fatalf("questCompleted reward = %#v, want inventory Reward Gem", completion.Rewards[0])
	}
}

func TestGameProjectFirstErrandCompletionSendsMessageAndQuestLogUpdate(t *testing.T) {
	testWorld, err := world.LoadFromGameFolder("../../game-project")
	if err != nil {
		t.Fatalf("LoadFromGameFolder returned error: %v", err)
	}

	game := NewGameWithWorld(testWorld)
	broadcasts := []message.Message{}
	game.RegisterBroadcaster(func(msg message.Message) {
		broadcasts = append(broadcasts, msg)
	})
	sent := []message.Message{}
	game.RegisterSender(func(clientID string, msg message.Message) {
		if clientID == "client-1" {
			sent = append(sent, msg)
		}
	})

	playerEntityId := model.NewEntityId()
	game.HandleJoin("client-1", playerEntityId, "player")
	game.EmitGameEvent(gameevent.New("conversation:node:new_conversation:start", playerEntityId))
	game.EmitGameEvent(gameevent.New("collect:name:mysterious_key", playerEntityId))
	game.EmitGameEvent(gameevent.New("kill:entity:rat", playerEntityId))
	game.update()

	completion := firstQuestCompletedPayload(t, sent)
	if completion.QuestId != "first_errand" {
		t.Fatalf("questCompleted questId = %q, want first_errand", completion.QuestId)
	}
	if completion.Rewards[0].Name != "Ancient Scroll" {
		t.Fatalf("quest reward = %#v, want Ancient Scroll", completion.Rewards[0])
	}
	if !gameUpdateIncludesCompletedQuest(broadcasts, playerEntityId.String(), "first_errand") {
		t.Fatal("game update did not include completed first_errand questlog record")
	}
}

func TestQuestRewardOverflowSpawnsLootableRewardDrop(t *testing.T) {
	testWorld := loadRewardQuestWorld(t, component.InventoryCapacity)
	game := NewGameWithWorld(testWorld)
	game.RegisterBroadcaster(func(message.Message) {})
	sent := []message.Message{}
	game.RegisterSender(func(clientID string, msg message.Message) {
		if clientID == "client-1" {
			sent = append(sent, msg)
		}
	})

	playerEntityId := model.NewEntityId()
	game.HandleJoin("client-1", playerEntityId, "player")
	game.EmitGameEvent(gameevent.New("finish:quest", playerEntityId))

	inventory := game.componentManager.GetEntityComponent(component.ComponentIdInventory, playerEntityId).(*component.CInventory)
	if inventory.GetItemCount() != component.InventoryCapacity {
		t.Fatalf("inventory count = %d, want full capacity %d", inventory.GetItemCount(), component.InventoryCapacity)
	}

	rewardDropEntityId, ok := firstEntityWithComponent(game, component.ComponentIdRewardDrop)
	if !ok {
		t.Fatal("overflow rewards did not spawn a reward drop")
	}
	lootable := game.componentManager.GetEntityComponent(component.ComponentIdLootable, rewardDropEntityId).(*component.CLootable)
	if lootable.ItemCount() != 5 {
		t.Fatalf("reward drop item count = %d, want 5", lootable.ItemCount())
	}
	metadata := game.componentManager.GetEntityComponent(component.ComponentIdMetadata, rewardDropEntityId)
	if metadata == nil {
		t.Fatal("reward drop has no metadata")
	}

	completion := firstQuestCompletedPayload(t, sent)
	if len(completion.Rewards) != 2 {
		t.Fatalf("questCompleted rewards = %#v, want inventory and dropped entries", completion.Rewards)
	}
	if completion.Rewards[0].Delivery != message.QuestRewardDeliveryInventory || completion.Rewards[0].Count != 15 {
		t.Fatalf("inventory reward delivery = %#v, want 15 inventory", completion.Rewards[0])
	}
	if completion.Rewards[1].Delivery != message.QuestRewardDeliveryDropped || completion.Rewards[1].Count != 5 {
		t.Fatalf("dropped reward delivery = %#v, want 5 dropped", completion.Rewards[1])
	}
}

func TestLootingRewardDropRemovesParcel(t *testing.T) {
	testWorld := loadRewardQuestWorld(t, component.InventoryCapacity)
	game := NewGameWithWorld(testWorld)
	game.RegisterBroadcaster(func(message.Message) {})
	game.RegisterSender(func(string, message.Message) {})

	playerEntityId := model.NewEntityId()
	game.HandleJoin("client-1", playerEntityId, "player")
	game.EmitGameEvent(gameevent.New("finish:quest", playerEntityId))

	rewardDropEntityId, ok := firstEntityWithComponent(game, component.ComponentIdRewardDrop)
	if !ok {
		t.Fatal("overflow rewards did not spawn a reward drop")
	}
	inventory := game.componentManager.GetEntityComponent(component.ComponentIdInventory, playerEntityId).(*component.CInventory)
	for _, item := range inventory.GetAllItems()[:5] {
		inventory.RemoveItem(item.Id)
	}
	game.componentManager.SetEntityComponent(playerEntityId, inventory)

	game.LootEntityFor(playerEntityId, rewardDropEntityId)
	if game.componentManager.HasEntity(rewardDropEntityId) {
		t.Fatal("reward drop entity still exists after successful looting")
	}
	if countInventoryItems(inventory, "Reward Gem", "quest") != component.InventoryCapacity {
		t.Fatalf("Reward Gem count = %d, want %d", countInventoryItems(inventory, "Reward Gem", "quest"), component.InventoryCapacity)
	}
}

func assertQuestProgress(
	t *testing.T,
	game *Game,
	playerEntityId model.EntityId,
	questId string,
	stepId string,
	stepIndex int,
	count int,
	completed bool,
) {
	t.Helper()

	questLogComponent := game.componentManager.GetEntityComponent(component.ComponentIdQuestLog, playerEntityId)
	if questLogComponent == nil {
		t.Fatal("player does not have questlog")
	}
	questLog := questLogComponent.(*component.CQuestLog)
	if questLog.IsCompleted(questId) != completed {
		t.Fatalf("quest completed = %v, want %v", questLog.IsCompleted(questId), completed)
	}
	if completed {
		return
	}

	var progress *component.QuestProgress
	for _, active := range questLog.GetActiveProgress() {
		if active.QuestId == questId {
			progress = active
			break
		}
	}
	if progress == nil {
		t.Fatalf("quest %q is not active", questId)
	}
	if progress.StepId != stepId || progress.CurrentStepIndex != stepIndex || progress.CurrentCount != count {
		t.Fatalf(
			"quest progress = (%q, %d, %d), want (%q, %d, %d)",
			progress.StepId,
			progress.CurrentStepIndex,
			progress.CurrentCount,
			stepId,
			stepIndex,
			count,
		)
	}
}

type questCompletedPayload struct {
	QuestId              string                        `json:"questId"`
	CompletedStepSummary string                        `json:"completedStepSummary"`
	Rewards              []message.QuestRewardDelivery `json:"rewards"`
}

func firstQuestCompletedPayload(t *testing.T, messages []message.Message) questCompletedPayload {
	t.Helper()

	for _, msg := range messages {
		if msg.Metadata.Type != message.MessageTypeQuestCompleted {
			continue
		}
		var payload struct {
			Data questCompletedPayload `json:"data"`
		}
		if err := json.Unmarshal([]byte(msg.Marshal()), &payload); err != nil {
			t.Fatalf("unmarshal questCompleted: %v", err)
		}
		return payload.Data
	}
	t.Fatal("no questCompleted message was sent")
	return questCompletedPayload{}
}

func gameUpdateIncludesCompletedQuest(messages []message.Message, entityId string, questId string) bool {
	for _, msg := range messages {
		if msg.Metadata.Type != message.MessageTypeGameUpdate {
			continue
		}
		var payload struct {
			Data struct {
				Entities []struct {
					EntityId    string `json:"entityId"`
					ComponentId string `json:"componentId"`
					Data        struct {
						Completed []struct {
							QuestId string `json:"questId"`
						} `json:"completed"`
					} `json:"data"`
				} `json:"entities"`
			} `json:"data"`
		}
		if err := json.Unmarshal([]byte(msg.Marshal()), &payload); err != nil {
			continue
		}
		for _, update := range payload.Data.Entities {
			if update.EntityId != entityId || update.ComponentId != component.ComponentIdQuestLog.String() {
				continue
			}
			for _, completed := range update.Data.Completed {
				if completed.QuestId == questId {
					return true
				}
			}
		}
	}
	return false
}

func loadRewardQuestWorld(t *testing.T, rewardCount int) *world.World {
	t.Helper()

	testWorld, err := world.LoadFromGameFS(fstest.MapFS{
		"game.json": {
			Data: []byte(`{
				"formatVersion": 1,
				"id": "test_game",
				"files": {
					"maps": ["maps/test.json"],
					"conversations": [],
					"quests": ["quests/reward.json"]
				}
			}`),
		},
		"maps/test.json": {
			Data: []byte(`{
				"formatVersion": 1,
				"id": "test",
				"size": { "x": 3, "y": 3 },
				"terrain": ["grass", "grass", "grass", "grass", "grass", "grass", "grass", "grass", "grass"],
				"heights": [0, 0, 0, 0, 0, 0, 0, 0, 0],
				"entities": [
					{
						"id": "player_spawn",
						"components": {
							"position": { "x": 1, "y": 1 },
							"playerSpawn": {}
						}
					}
				]
			}`),
		},
		"quests/reward.json": {
			Data: []byte(`{
				"formatVersion": 2,
				"id": "reward_quests",
				"quests": [
					{
						"id": "reward_test",
						"displayName": "Reward Test",
						"description": "A quest with item rewards.",
						"startEventId": "finish:quest",
						"steps": [
							{
								"id": "finish",
								"description": "Finish the quest.",
								"requirement": { "eventId": "finish:quest", "count": 1 }
							}
						],
						"rewards": {
							"items": [
								{ "name": "Reward Gem", "type": "quest", "count": ` + intString(rewardCount) + ` }
							]
						}
					}
				]
			}`),
		},
	})
	if err != nil {
		t.Fatalf("LoadFromGameFS returned error: %v", err)
	}
	return testWorld
}

func countInventoryItems(inventory *component.CInventory, name string, itemType string) int {
	count := 0
	for _, item := range inventory.GetAllItems() {
		if item.Name == name && item.Type == itemType {
			count++
		}
	}
	return count
}

func intString(value int) string {
	return strconv.Itoa(value)
}
