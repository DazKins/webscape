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
	"webscape/server/util"
)

func TestClientInterestStreamsChunksSnapshotsAndTombstones(t *testing.T) {
	gameWorld, err := world.LoadFromGameFS(interestWorldFS())
	if err != nil {
		t.Fatal(err)
	}
	game := NewGameWithWorldAndChunkRadius(gameWorld, 0)
	sent := make([]message.Message, 0)
	game.RegisterSender(func(_ string, msg message.Message) { sent = append(sent, msg) })
	playerID := model.NewEntityId()
	game.HandleRegister("client", playerID, "player")
	nearID := authoredRuntimeID(t, game, "near")
	farID := authoredRuntimeID(t, game, "far")
	state := game.clients["client"]
	if !state.loadedChunks[world.ChunkCoord{X: 0, Y: 0}] || len(state.loadedChunks) != 1 {
		t.Fatalf("initial chunks = %#v", state.loadedChunks)
	}
	if baselineKnows(state, farID) {
		t.Fatal("registration exposed distant entity")
	}
	nearKnown := knownComponentCount(state, nearID)
	if nearKnown == 0 {
		t.Fatal("registration omitted nearby entity snapshot")
	}

	sent = sent[:0]
	movePlayerForTest(game, playerID, 4, 0)
	game.syncClient("client")
	if len(sent) < 2 || sent[0].Metadata.Type != message.MessageTypeChunkUpdate || sent[1].Metadata.Type != message.MessageTypeGameUpdate {
		t.Fatalf("transition order = %v", messageTypes(sent))
	}
	if baselineKnows(state, nearID) {
		t.Fatal("visibility loss retained near entity baseline")
	}
	if !baselineKnows(state, farID) {
		t.Fatal("newly visible far entity did not receive full snapshot")
	}
	if tombstoneCount(t, sent, nearID.String()) != nearKnown {
		t.Fatalf("near tombstones did not match known components")
	}

	sent = sent[:0]
	movePlayerForTest(game, playerID, 0, 0)
	game.syncClient("client")
	if knownComponentCount(state, nearID) != nearKnown {
		t.Fatal("re-entry did not restore complete current snapshot")
	}
	if tombstoneCount(t, sent, farID.String()) == 0 {
		t.Fatal("far visibility loss emitted no tombstones")
	}
}

func TestActualDeletionUsesSameTombstonePath(t *testing.T) {
	gameWorld, err := world.LoadFromGameFS(interestWorldFS())
	if err != nil {
		t.Fatal(err)
	}
	game := NewGameWithWorldAndChunkRadius(gameWorld, 0)
	sent := make([]message.Message, 0)
	game.RegisterSender(func(_ string, msg message.Message) { sent = append(sent, msg) })
	playerID := model.NewEntityId()
	game.HandleRegister("client", playerID, "player")
	nearID := authoredRuntimeID(t, game, "near")
	want := knownComponentCount(game.clients["client"], nearID)
	sent = sent[:0]
	game.componentManager.RemoveEntity(nearID)
	game.syncClient("client")
	if tombstoneCount(t, sent, nearID.String()) != want {
		t.Fatal("deletion did not tombstone every known component")
	}
}

func TestMovementCommandIsLimitedToLoadedAuthoredChunks(t *testing.T) {
	gameWorld, err := world.LoadFromGameFS(interestWorldFS())
	if err != nil {
		t.Fatal(err)
	}
	game := NewGameWithWorldAndChunkRadius(gameWorld, 0)
	game.RegisterSender(func(string, message.Message) {})
	playerID := model.NewEntityId()
	game.HandleRegister("client", playerID, "player")
	game.HandleMove("client", 4, 0)
	if game.componentManager.GetEntityComponent(component.ComponentIdPathing, playerID) != nil {
		t.Fatal("movement accepted target outside interest")
	}
	game.HandleMove("client", -1, 0)
	if game.componentManager.GetEntityComponent(component.ComponentIdPathing, playerID) != nil {
		t.Fatal("movement accepted void target")
	}
	game.HandleMove("client", 1, 1)
	if game.componentManager.GetEntityComponent(component.ComponentIdPathing, playerID) == nil {
		t.Fatal("movement rejected loaded authored target")
	}
}

func interestWorldFS() fstest.MapFS {
	manifest := `{"formatVersion":2,"id":"interest","world":{"chunkSize":{"x":2,"y":2}},"files":{"chunks":["chunks/0.json","chunks/1.json","chunks/2.json"],"conversations":[],"quests":[]}}`
	return fstest.MapFS{
		"game.json":     {Data: []byte(manifest)},
		"chunks/0.json": {Data: []byte(interestChunk("zero", 0, `[{"id":"player_spawn","components":{"position":{"x":0,"y":0},"playerSpawn":{}}},{"id":"near","components":{"position":{"x":1,"y":0},"metadata":{"name":"Near"},"renderable":{"type":"rock"}}}]`))},
		"chunks/1.json": {Data: []byte(interestChunk("one", 1, `[]`))},
		"chunks/2.json": {Data: []byte(interestChunk("two", 2, `[{"id":"far","components":{"position":{"x":0,"y":0},"metadata":{"name":"Far"},"renderable":{"type":"rock"}}}]`))},
	}
}

func interestChunk(id string, x int, entities string) string {
	return fmt.Sprintf(`{"formatVersion":2,"id":%q,"coordinate":{"x":%d,"y":0},"terrain":["grass","grass","grass","grass"],"heights":[0,0,0,0],"blockers":[false,false,false,false],"walls":[],"entities":%s}`, id, x, entities)
}

func authoredRuntimeID(t *testing.T, game *Game, authoredID string) model.EntityId {
	t.Helper()
	for entityID, value := range game.componentManager.GetComponent(component.ComponentIdMetadata) {
		metadata, ok := value.(*component.CMetadata).GetMetadata().(util.JObject)
		if !ok {
			continue
		}
		if id, ok := metadata["entityId"].(util.JString); ok && string(id) == authoredID {
			return entityID
		}
	}
	t.Fatalf("authored entity %q not found", authoredID)
	return model.EntityId{}
}

func movePlayerForTest(game *Game, id model.EntityId, x, y int) {
	position := game.componentManager.GetEntityComponent(component.ComponentIdPosition, id).(*component.CPosition)
	position.SetPosition(math.Vec2{X: x, Y: y})
	game.componentManager.SetEntityComponent(id, position)
}

func baselineKnows(state *clientStreamState, id model.EntityId) bool {
	return knownComponentCount(state, id) > 0
}
func knownComponentCount(state *clientStreamState, id model.EntityId) int {
	count := 0
	for _, entities := range state.baseline {
		if _, ok := entities[id]; ok {
			count++
		}
	}
	return count
}
func messageTypes(messages []message.Message) []message.MessageType {
	result := make([]message.MessageType, len(messages))
	for i, msg := range messages {
		result[i] = msg.Metadata.Type
	}
	return result
}
func tombstoneCount(t *testing.T, messages []message.Message, id string) int {
	t.Helper()
	count := 0
	for _, msg := range messages {
		if msg.Metadata.Type != message.MessageTypeGameUpdate {
			continue
		}
		var payload struct {
			Data struct {
				Entities []struct {
					EntityID string `json:"entityId"`
					Data     any    `json:"data"`
				} `json:"entities"`
			} `json:"data"`
		}
		if err := json.Unmarshal([]byte(msg.Marshal()), &payload); err != nil {
			t.Fatal(err)
		}
		for _, update := range payload.Data.Entities {
			if update.EntityID == id && update.Data == nil {
				count++
			}
		}
	}
	return count
}
