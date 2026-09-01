package game

import (
	"encoding/json"
	"strings"
	"testing"
	"webscape/server/game/component"
	"webscape/server/game/model"
	"webscape/server/game/world"
	"webscape/server/message"
	"webscape/server/util"
)

func TestRegistrationNormalizesAndSerializesPlayerName(t *testing.T) {
	gameWorld, err := world.LoadFromGameFS(interestWorldFS())
	if err != nil {
		t.Fatal(err)
	}
	game := NewGameWithWorld(gameWorld)
	sent := []message.Message{}
	game.RegisterSender(func(_ string, msg message.Message) { sent = append(sent, msg) })
	playerID := model.NewEntityId()

	game.HandleRegister("client", playerID, "  夜 空  ")

	if !game.IsRegistered("client") {
		t.Fatal("client was not registered")
	}
	registered := registrationPayload(t, sent[0])
	if registered.Type != "registered" || registered.EntityID != playerID.String() || registered.Name != "夜 空" {
		t.Fatalf("registered payload = %#v", registered)
	}

	player := game.componentManager.GetEntityComponent(component.ComponentIdPlayer, playerID)
	if player == nil {
		t.Fatal("player component was not created")
	}
	playerObject, ok := player.(*component.CPlayer).Serialize().(util.JObject)
	if !ok || playerObject["name"] != util.JString("夜 空") {
		t.Fatal("serialized player component did not contain the normalized name")
	}
	metadata := game.componentManager.GetEntityComponent(component.ComponentIdMetadata, playerID)
	metadataObject, ok := metadata.(*component.CMetadata).Serialize().(util.JObject)
	if !ok || metadataObject["name"] != util.JString("夜 空") {
		t.Fatal("serialized metadata component did not contain the normalized name")
	}
	appearance := game.componentManager.GetEntityComponent(component.ComponentIdAppearance, playerID)
	if appearance == nil {
		t.Fatal("registered player has no appearance component")
	}
	if err := component.ValidateAppearance(appearance.(*component.CAppearance).GetAppearance()); err != nil {
		t.Fatalf("registered player appearance is invalid: %v", err)
	}
}

func TestAppearanceSnapshotIsDeltaOnly(t *testing.T) {
	gameWorld, err := world.LoadFromGameFS(interestWorldFS())
	if err != nil {
		t.Fatal(err)
	}
	game := NewGameWithWorld(gameWorld)
	sent := []message.Message{}
	game.RegisterSender(func(_ string, msg message.Message) { sent = append(sent, msg) })
	playerID := model.NewEntityId()
	game.HandleRegister("client", playerID, "Player")

	componentIDs, _ := streamedComponentIDs(t, sent)
	if !componentIDs[component.ComponentIdAppearance.String()] {
		t.Fatal("initial player snapshot omitted appearance")
	}
	sent = nil
	game.syncClient("client")
	componentIDs, _ = streamedComponentIDs(t, sent)
	if componentIDs[component.ComponentIdAppearance.String()] {
		t.Fatal("unchanged appearance produced a redundant delta")
	}
}

func TestRegistrationRejectsInvalidNames(t *testing.T) {
	tests := []struct {
		name  string
		value string
	}{
		{name: "blank", value: " \t\n "},
		{name: "overlong unicode", value: strings.Repeat("界", 25)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gameWorld, err := world.LoadFromGameFS(interestWorldFS())
			if err != nil {
				t.Fatal(err)
			}
			game := NewGameWithWorld(gameWorld)
			sent := []message.Message{}
			game.RegisterSender(func(_ string, msg message.Message) { sent = append(sent, msg) })
			playerID := model.NewEntityId()

			game.HandleRegister("client", playerID, test.value)

			if game.IsRegistered("client") {
				t.Fatal("invalid name registered the client")
			}
			if game.componentManager.GetEntityComponent(component.ComponentIdPlayer, playerID) != nil {
				t.Fatal("invalid name created a player entity")
			}
			payload := registrationPayload(t, sent[0])
			if payload.Type != "registrationFailed" || payload.Reason == "" {
				t.Fatalf("failure payload = %#v", payload)
			}
		})
	}
}

func TestRegistrationAcceptsTwentyFourUnicodeCharacters(t *testing.T) {
	gameWorld, err := world.LoadFromGameFS(interestWorldFS())
	if err != nil {
		t.Fatal(err)
	}
	game := NewGameWithWorld(gameWorld)
	game.RegisterSender(func(string, message.Message) {})

	game.HandleRegister("client", model.NewEntityId(), strings.Repeat("界", 24))

	if !game.IsRegistered("client") {
		t.Fatal("24-character Unicode name was rejected")
	}
}

func TestRegistrationRejectsDuplicateConnectionAndActiveEntityID(t *testing.T) {
	gameWorld, err := world.LoadFromGameFS(interestWorldFS())
	if err != nil {
		t.Fatal(err)
	}
	game := NewGameWithWorld(gameWorld)
	sentByClient := map[string][]message.Message{}
	game.RegisterSender(func(clientID string, msg message.Message) {
		sentByClient[clientID] = append(sentByClient[clientID], msg)
	})
	playerID := model.NewEntityId()
	game.HandleRegister("client-1", playerID, "First")

	sentByClient["client-1"] = nil
	game.HandleRegister("client-1", model.NewEntityId(), "Second")
	duplicate := registrationPayload(t, sentByClient["client-1"][0])
	if duplicate.Type != "registrationFailed" {
		t.Fatalf("duplicate connection response = %#v", duplicate)
	}

	game.HandleRegister("client-2", playerID, "Other")
	conflict := registrationPayload(t, sentByClient["client-2"][0])
	if conflict.Type != "registrationFailed" || game.IsRegistered("client-2") {
		t.Fatalf("active id conflict response = %#v", conflict)
	}
}

func TestRegistrationAllowsDuplicateNames(t *testing.T) {
	gameWorld, err := world.LoadFromGameFS(interestWorldFS())
	if err != nil {
		t.Fatal(err)
	}
	game := NewGameWithWorld(gameWorld)
	game.RegisterSender(func(string, message.Message) {})

	game.HandleRegister("client-1", model.NewEntityId(), "Same Name")
	game.HandleRegister("client-2", model.NewEntityId(), "Same Name")

	if !game.IsRegistered("client-1") || !game.IsRegistered("client-2") {
		t.Fatal("duplicate display names were rejected")
	}
}

func TestPreRegistrationGameplayCommandsAreIgnored(t *testing.T) {
	gameWorld, err := world.LoadFromGameFS(interestWorldFS())
	if err != nil {
		t.Fatal(err)
	}
	game := NewGameWithWorld(gameWorld)
	game.RegisterSender(func(string, message.Message) {})

	game.HandleMove("unregistered", 1, 1)
	game.HandleChat("unregistered", "hello")
	game.HandleInteract("unregistered", model.NewEntityId(), component.InteractionOptionTalk)
	game.HandleConversationOption("unregistered", "conversation", "node", "option")
	game.HandleEquip("unregistered", model.NewItemId())
	game.HandleUnequip("unregistered", model.SlotHead)

	if game.IsRegistered("unregistered") {
		t.Fatal("gameplay command registered a client")
	}
	if len(game.componentManager.GetComponent(component.ComponentIdPlayer)) != 0 {
		t.Fatal("gameplay command created a player")
	}
}

type registrationMessagePayload struct {
	Type     string
	EntityID string
	Name     string
	Reason   string
}

func registrationPayload(t *testing.T, msg message.Message) registrationMessagePayload {
	t.Helper()
	var payload struct {
		Metadata struct {
			Type string `json:"type"`
		} `json:"metadata"`
		Data struct {
			EntityID string `json:"entityId"`
			Name     string `json:"name"`
			Reason   string `json:"reason"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(msg.Marshal()), &payload); err != nil {
		t.Fatal(err)
	}
	return registrationMessagePayload{
		Type: payload.Metadata.Type, EntityID: payload.Data.EntityID,
		Name: payload.Data.Name, Reason: payload.Data.Reason,
	}
}
