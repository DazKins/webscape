package game

import (
	"encoding/json"
	"testing"
	"testing/fstest"
	"webscape/server/game/component"
	"webscape/server/game/model"
	"webscape/server/game/world"
	"webscape/server/math"
	"webscape/server/message"
)

func TestLootableObjectAddsAuthoredItemsToPlayerInventory(t *testing.T) {
	testWorld, err := world.LoadFromGameFS(chunkTestFS(t, fstest.MapFS{
		"game.json": {
			Data: []byte(`{
				"formatVersion": 1,
				"id": "test_game",
				"files": {
					"maps": ["maps/test.json"],
					"conversations": [],
					"quests": []
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
						"id": "chest_001",
						"components": {
							"position": { "x": 1, "y": 0 },
							"lootable": {
								"once": true,
								"items": [
									{ "definitionId": "mysteriousKey", "count": 1 }
								]
							}
						}
					}
				]
			}`),
		},
	}))
	if err != nil {
		t.Fatalf("LoadFromGameFS returned error: %v", err)
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

	chestEntityId, ok := firstEntityWithComponent(game, component.ComponentIdLootable)
	if !ok {
		t.Fatal("no object entity has CLootable")
	}

	playerEntityId := model.NewEntityId()
	game.HandleRegister("client-1", playerEntityId, "player")
	if !gameUpdateIncludesInteraction(sent, chestEntityId.String(), "loot") {
		t.Fatal("initial entity update did not include loot interaction for chest")
	}
	sent = nil
	game.HandleInteract("client-1", chestEntityId, component.InteractionOptionLoot)
	game.update()

	inventory := game.componentManager.GetEntityComponent(component.ComponentIdInventory, playerEntityId).(*component.CInventory)
	if !inventoryContains(inventory, "Mysterious Key", "quest") {
		t.Fatal("player inventory does not contain looted Mysterious Key")
	}

	lootable := game.componentManager.GetEntityComponent(component.ComponentIdLootable, chestEntityId).(*component.CLootable)
	if !lootable.IsLooted() {
		t.Fatal("lootable chest was not marked looted")
	}

	for _, option := range game.getInteractionOptionsForEntity(chestEntityId) {
		if option == component.InteractionOptionLoot {
			t.Fatal("looted once-only chest still offers loot interaction")
		}
	}
	if gameUpdateIncludesInteraction(broadcasts, chestEntityId.String(), "loot") {
		t.Fatal("post-loot entity update still included loot interaction for chest")
	}
}

func inventoryContains(inventory *component.CInventory, name string, itemType string) bool {
	for _, item := range inventory.GetAllItems() {
		if item.Name() == name && item.Type() == itemType {
			return true
		}
	}
	return false
}

func gameUpdateIncludesInteraction(messages []message.Message, entityId string, interaction string) bool {
	for _, msg := range messages {
		if msg.Metadata.Type != message.MessageTypeGameUpdate {
			continue
		}
		var payload struct {
			Data struct {
				Entities []struct {
					EntityId              string   `json:"entityId"`
					AvailableInteractions []string `json:"availableInteractions"`
				} `json:"entities"`
			} `json:"data"`
		}
		if err := json.Unmarshal([]byte(msg.Marshal()), &payload); err != nil {
			continue
		}
		for _, update := range payload.Data.Entities {
			if update.EntityId != entityId {
				continue
			}
			for _, availableInteraction := range update.AvailableInteractions {
				if availableInteraction == interaction {
					return true
				}
			}
		}
	}
	return false
}

func TestDefinitionReferencesPreserveEquipmentAndStackRules(t *testing.T) {
	g, player, inventory := newDropTestGame(t)
	rewards := g.deliverQuestRewards(player, []world.QuestRewardItem{{DefinitionID: "chainmailChestplate", Count: 1}})
	mail := inventory.FindByDefinition("chainmailChestplate")
	if mail == nil || !mail.IsEquipable() || mail.CombatStats().ArmorBonus != 4 || len(rewards) != 1 || rewards[0].DefinitionID != "chainmailChestplate" {
		t.Fatal("quest reference did not create the catalogue's equipment")
	}
	arrows := inventory.FindByDefinition("arrow")
	before := arrows.Quantity
	for !inventory.IsFull() {
		inventory.AddItem(model.NewItem("bread"))
	}
	loot := component.NewCLootable(true, []component.LootItem{{DefinitionID: "arrow", Count: 100}})
	target := g.componentManager.CreateNewEntity(component.NewCPosition(math.Vec2{X: 1, Y: 0}), loot)
	g.update()
	g.LootEntityFor(player, target)
	if !loot.IsLooted() || inventory.FindByDefinition("arrow").Quantity != before+100 || inventory.GetItemCount() != component.InventoryCapacity {
		t.Fatal("catalogue stack rule not applied when looting into a full inventory")
	}
}
