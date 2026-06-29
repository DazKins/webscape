package game

import (
	"encoding/json"
	"fmt"
	"testing"
	"testing/fstest"
	"webscape/server/game/component"
	"webscape/server/game/model"
	"webscape/server/game/world"
	"webscape/server/math"
	"webscape/server/message"
)

func TestSpawnEntityInitializesChildAndDoesNotSerializeSpawnComponent(t *testing.T) {
	testWorld := loadSpawnTestWorld(t, 20)
	game := NewGameWithWorld(testWorld)
	game.RegisterBroadcaster(func(message.Message) {})

	spawnEntityId, spawn := onlySpawn(t, game)
	childEntityId := assertSpawnChild(t, game, spawn, math.Vec2{X: 2, Y: 0})

	metadataEntities := game.componentManager.GetComponent(component.ComponentIdMetadata)
	if len(metadataEntities) != 2 {
		t.Fatalf("metadata entity count = %d, want static object plus spawned child", len(metadataEntities))
	}

	snapshot := game.serializedComponentsSnapshot()
	if _, ok := snapshot[component.ComponentIdSpawn]; ok {
		t.Fatal("serialized snapshot includes server-only spawn component")
	}
	if _, ok := snapshot[component.ComponentIdPosition][spawnEntityId]; !ok {
		t.Fatal("serialized snapshot does not include spawn entity position")
	}
	if _, ok := snapshot[component.ComponentIdPosition][childEntityId]; !ok {
		t.Fatal("serialized snapshot does not include spawned child position")
	}

	sent := []message.Message{}
	game.RegisterSender(func(clientID string, msg message.Message) {
		if clientID == "client-1" {
			sent = append(sent, msg)
		}
	})
	game.HandleJoin("client-1", model.NewEntityId(), "player")
	gameUpdateEntities := gameUpdateEntityComponents(sent)
	if gameUpdateEntities[spawnEntityId.String()][component.ComponentIdSpawn.String()] {
		t.Fatal("joining client received spawn component")
	}
	if !gameUpdateEntities[spawnEntityId.String()][component.ComponentIdPosition.String()] {
		t.Fatal("joining client did not receive spawn entity position")
	}
	if !gameUpdateEntities[childEntityId.String()][component.ComponentIdPosition.String()] {
		t.Fatal("joining client did not receive spawned child position")
	}

	for i := 0; i < 3; i++ {
		game.update()
	}
	if nextChildEntityId := spawn.GetChildEntityId(); nextChildEntityId != childEntityId {
		t.Fatalf("spawn created another child while first child was alive: got %s, want %s", nextChildEntityId, childEntityId)
	}
}

func TestSpawnRespawnsAfterDelayAtOriginalPosition(t *testing.T) {
	testWorld := loadSpawnTestWorld(t, 2)
	game := NewGameWithWorld(testWorld)
	game.RegisterBroadcaster(func(message.Message) {})

	_, spawn := onlySpawn(t, game)
	initialChildEntityId := assertSpawnChild(t, game, spawn, math.Vec2{X: 2, Y: 0})

	position := game.componentManager.GetEntityComponent(component.ComponentIdPosition, initialChildEntityId).(*component.CPosition)
	position.SetPosition(math.Vec2{X: 3, Y: 0})
	game.componentManager.SetEntityComponent(initialChildEntityId, position)
	health := game.componentManager.GetEntityComponent(component.ComponentIdHealth, initialChildEntityId).(*component.CHealth)
	health.SetCurrentHealth(0)
	game.componentManager.SetEntityComponent(initialChildEntityId, health)

	game.update()
	if game.componentManager.HasEntity(initialChildEntityId) {
		t.Fatal("dead child still exists after health system update")
	}
	if spawn.HasChildEntityId() {
		t.Fatal("spawn still tracks removed child")
	}
	if spawn.GetRemainingRespawnTicks() != 2 {
		t.Fatalf("remaining respawn ticks = %d, want 2", spawn.GetRemainingRespawnTicks())
	}

	game.update()
	if spawn.HasChildEntityId() {
		t.Fatal("spawn respawned child too early")
	}
	if spawn.GetRemainingRespawnTicks() != 1 {
		t.Fatalf("remaining respawn ticks = %d, want 1", spawn.GetRemainingRespawnTicks())
	}

	game.update()
	respawnedChildEntityId := assertSpawnChild(t, game, spawn, math.Vec2{X: 2, Y: 0})
	if respawnedChildEntityId == initialChildEntityId {
		t.Fatal("respawn reused the removed child runtime entity id")
	}
}

func TestSpawnWithZeroDelayRespawnsOnNextSpawnSystemUpdate(t *testing.T) {
	testWorld := loadSpawnTestWorld(t, 0)
	game := NewGameWithWorld(testWorld)
	game.RegisterBroadcaster(func(message.Message) {})

	_, spawn := onlySpawn(t, game)
	initialChildEntityId := assertSpawnChild(t, game, spawn, math.Vec2{X: 2, Y: 0})
	health := game.componentManager.GetEntityComponent(component.ComponentIdHealth, initialChildEntityId).(*component.CHealth)
	health.SetCurrentHealth(0)
	game.componentManager.SetEntityComponent(initialChildEntityId, health)

	game.update()
	respawnedChildEntityId := assertSpawnChild(t, game, spawn, math.Vec2{X: 2, Y: 0})
	if respawnedChildEntityId == initialChildEntityId {
		t.Fatal("zero-delay respawn reused the removed child runtime entity id")
	}
}

func loadSpawnTestWorld(t *testing.T, respawnTicks int) *world.World {
	t.Helper()

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
			Data: []byte(fmt.Sprintf(`{
				"formatVersion": 1,
				"id": "test",
				"size": { "x": 4, "y": 1 },
				"terrain": ["grass", "grass", "grass", "grass"],
				"heights": [0, 0, 0, 0],
				"entities": [
					{
						"id": "player_spawn",
						"components": {
							"position": { "x": 0, "y": 0 },
							"playerSpawn": {}
						}
					},
					{
						"id": "sign_001",
						"components": {
							"position": { "x": 1, "y": 0 },
							"metadata": { "name": "Sign", "entityType": "sign" }
						}
					},
					{
						"id": "rat_spawn_001",
						"components": {
							"position": { "x": 2, "y": 0 },
							"spawn": {
								"respawnTicks": %d,
								"entity": {
									"components": {
										"metadata": { "name": "Rat", "entityType": "rat" },
										"renderable": { "type": "human" },
										"randomwalk": { "walkTimer": 10 },
										"health": { "maxHealth": 100 },
										"basestats": { "strength": 6, "dexterity": 5, "vitality": 6 },
										"equipped": {},
										"combatstats": {}
									}
								}
							}
						}
					}
				]
			}`, respawnTicks)),
		},
	})
	if err != nil {
		t.Fatalf("LoadFromGameFS returned error: %v", err)
	}
	return testWorld
}

func onlySpawn(t *testing.T, game *Game) (model.EntityId, *component.CSpawn) {
	t.Helper()

	spawns := game.componentManager.GetComponent(component.ComponentIdSpawn)
	if len(spawns) != 1 {
		t.Fatalf("spawn count = %d, want 1", len(spawns))
	}
	for entityId, spawn := range spawns {
		return entityId, spawn.(*component.CSpawn)
	}
	t.Fatal("spawn count was 1 but no spawn component was found")
	return model.EntityId{}, nil
}

func assertSpawnChild(t *testing.T, game *Game, spawn *component.CSpawn, wantPosition math.Vec2) model.EntityId {
	t.Helper()

	if !spawn.HasChildEntityId() {
		t.Fatal("spawn does not track a child entity")
	}
	childEntityId := spawn.GetChildEntityId()
	if !game.componentManager.HasEntity(childEntityId) {
		t.Fatalf("spawned child %s does not exist", childEntityId)
	}
	position := game.componentManager.GetEntityComponent(component.ComponentIdPosition, childEntityId)
	if position == nil {
		t.Fatalf("spawned child %s has no position", childEntityId)
	}
	if gotPosition := position.(*component.CPosition).GetPosition(); gotPosition != wantPosition {
		t.Fatalf("spawned child position = (%d, %d), want (%d, %d)", gotPosition.X, gotPosition.Y, wantPosition.X, wantPosition.Y)
	}
	return childEntityId
}

func gameUpdateEntityComponents(messages []message.Message) map[string]map[string]bool {
	result := make(map[string]map[string]bool)
	for _, msg := range messages {
		if msg.Metadata.Type != message.MessageTypeGameUpdate {
			continue
		}
		var payload struct {
			Data struct {
				Entities []struct {
					EntityId    string `json:"entityId"`
					ComponentId string `json:"componentId"`
				} `json:"entities"`
			} `json:"data"`
		}
		if err := json.Unmarshal([]byte(msg.Marshal()), &payload); err != nil {
			continue
		}
		for _, update := range payload.Data.Entities {
			if result[update.EntityId] == nil {
				result[update.EntityId] = make(map[string]bool)
			}
			result[update.EntityId][update.ComponentId] = true
		}
	}
	return result
}
