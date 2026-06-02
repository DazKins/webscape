package game

import (
	"encoding/json"
	"testing"
	"testing/fstest"
	"webscape/server/game/component"
	"webscape/server/game/model"
	"webscape/server/game/world"
	"webscape/server/message"
)

func TestLootableObjectAddsAuthoredItemsToPlayerInventory(t *testing.T) {
	testWorld, err := world.LoadFromGameFS(fstest.MapFS{
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
									{ "name": "Mysterious Key", "type": "quest" }
								]
							}
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
	game.HandleJoin("client-1", playerEntityId, "player")
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
		if item.Name == name && item.Type == itemType {
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
