package game

import (
	"encoding/json"
	"testing"
	"webscape/server/game/component"
	"webscape/server/game/model"
	"webscape/server/game/world"
	"webscape/server/message"
	"webscape/server/util"
)

func TestViewerConnectionReceivesSpawnAreaWithoutCreatingPlayer(t *testing.T) {
	gameWorld, err := world.LoadFromGameFS(interestWorldFS())
	if err != nil {
		t.Fatal(err)
	}
	game := NewGameWithWorldAndChunkRadius(gameWorld, 0)
	sent := []message.Message{}
	game.RegisterSender(func(_ string, msg message.Message) { sent = append(sent, msg) })

	game.HandleConnect("viewer")

	if game.IsRegistered("viewer") {
		t.Fatal("viewer connection was registered as a player")
	}
	if len(game.componentManager.GetComponent(component.ComponentIdPlayer)) != 0 {
		t.Fatal("viewer connection created a player entity")
	}
	state := game.clients["viewer"]
	if state == nil || !state.loadedChunks[world.ChunkCoord{X: 0, Y: 0}] {
		t.Fatalf("viewer stream state = %#v", state)
	}
	assertMessageTypesInclude(t, sent,
		message.MessageTypeWorld,
		message.MessageTypeChunkUpdate,
		message.MessageTypeGameUpdate,
	)
}

func TestViewerReceivesOnlyPublicComponents(t *testing.T) {
	gameWorld, err := world.LoadFromGameFS(interestWorldFS())
	if err != nil {
		t.Fatal(err)
	}
	game := NewGameWithWorldAndChunkRadius(gameWorld, 0)
	sentByClient := map[string][]message.Message{}
	game.RegisterSender(func(clientID string, msg message.Message) {
		sentByClient[clientID] = append(sentByClient[clientID], msg)
	})
	playerID := model.NewEntityId()
	game.HandleRegister("player-client", playerID, "Visible Player")

	game.HandleConnect("viewer")

	componentIDs, interactions := streamedComponentIDs(t, sentByClient["viewer"])
	for _, publicID := range []component.ComponentId{
		component.ComponentIdPosition,
		component.ComponentIdMetadata,
		component.ComponentIdRenderable,
		component.ComponentIdAppearance,
	} {
		if !componentIDs[publicID.String()] {
			t.Fatalf("viewer snapshot omitted public component %q", publicID)
		}
	}
	for _, privateID := range []component.ComponentId{
		component.ComponentIdPlayer,
		component.ComponentIdInventory,
		component.ComponentIdEquipped,
		component.ComponentIdQuestLog,
		component.ComponentIdCombatLog,
		component.ComponentIdBaseStats,
		component.ComponentIdCombatStats,
	} {
		if componentIDs[privateID.String()] {
			t.Fatalf("viewer snapshot exposed private component %q", privateID)
		}
	}
	if interactions != 0 {
		t.Fatalf("viewer snapshot exposed %d gameplay interactions", interactions)
	}
}

func TestRegistrationUpgradesExistingViewerStream(t *testing.T) {
	gameWorld, err := world.LoadFromGameFS(interestWorldFS())
	if err != nil {
		t.Fatal(err)
	}
	game := NewGameWithWorldAndChunkRadius(gameWorld, 0)
	sent := []message.Message{}
	game.RegisterSender(func(_ string, msg message.Message) { sent = append(sent, msg) })
	game.HandleConnect("client")
	sent = nil

	game.HandleRegister("client", model.NewEntityId(), "New Player")

	if !game.IsRegistered("client") {
		t.Fatal("viewer was not upgraded to a registered player")
	}
	if len(sent) == 0 || sent[0].Metadata.Type != message.MessageTypeRegistered {
		t.Fatalf("upgrade messages start with %v, want registered", messageTypes(sent))
	}
	for _, msg := range sent {
		if msg.Metadata.Type == message.MessageTypeWorld || msg.Metadata.Type == message.MessageTypeChunkUpdate {
			t.Fatalf("viewer upgrade unnecessarily resent %q", msg.Metadata.Type)
		}
	}
	componentIDs, _ := streamedComponentIDs(t, sent)
	if !componentIDs[component.ComponentIdInventory.String()] ||
		!componentIDs[component.ComponentIdPlayer.String()] {
		t.Fatal("registered snapshot omitted private player components")
	}
}

func TestViewerReceivesOngoingPublicUpdates(t *testing.T) {
	gameWorld, err := world.LoadFromGameFS(interestWorldFS())
	if err != nil {
		t.Fatal(err)
	}
	game := NewGameWithWorldAndChunkRadius(gameWorld, 0)
	sent := []message.Message{}
	game.RegisterSender(func(_ string, msg message.Message) { sent = append(sent, msg) })
	game.HandleConnect("viewer")
	nearID := authoredRuntimeID(t, game, "near")
	sent = nil

	game.componentManager.SetEntityComponent(nearID, component.NewCMetadata(util.JObject{
		"name":     util.JString("Near, but alive"),
		"entityId": util.JString(nearID.String()),
	}))
	game.update()

	componentIDs, _ := streamedComponentIDs(t, sent)
	if !componentIDs[component.ComponentIdMetadata.String()] {
		t.Fatal("viewer did not receive an ongoing public entity update")
	}
}

func TestViewerDisconnectRemovesStreamState(t *testing.T) {
	gameWorld, err := world.LoadFromGameFS(interestWorldFS())
	if err != nil {
		t.Fatal(err)
	}
	game := NewGameWithWorld(gameWorld)
	game.RegisterSender(func(string, message.Message) {})
	game.HandleConnect("viewer")

	game.HandleLeave("viewer")

	if game.clients["viewer"] != nil {
		t.Fatal("viewer stream state remained after disconnect")
	}
}

func assertMessageTypesInclude(t *testing.T, messages []message.Message, expected ...message.MessageType) {
	t.Helper()
	seen := map[message.MessageType]bool{}
	for _, msg := range messages {
		seen[msg.Metadata.Type] = true
	}
	for _, messageType := range expected {
		if !seen[messageType] {
			t.Fatalf("message types %v omitted %q", messageTypes(messages), messageType)
		}
	}
}

func streamedComponentIDs(t *testing.T, messages []message.Message) (map[string]bool, int) {
	t.Helper()
	componentIDs := map[string]bool{}
	interactionCount := 0
	for _, msg := range messages {
		if msg.Metadata.Type != message.MessageTypeGameUpdate {
			continue
		}
		var payload struct {
			Data struct {
				Entities []struct {
					ComponentID           string   `json:"componentId"`
					AvailableInteractions []string `json:"availableInteractions"`
				} `json:"entities"`
			} `json:"data"`
		}
		if err := json.Unmarshal([]byte(msg.Marshal()), &payload); err != nil {
			t.Fatal(err)
		}
		for _, entityUpdate := range payload.Data.Entities {
			componentIDs[entityUpdate.ComponentID] = true
			interactionCount += len(entityUpdate.AvailableInteractions)
		}
	}
	return componentIDs, interactionCount
}
