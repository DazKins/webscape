package server

import (
	"bytes"
	"testing"
	"testing/fstest"
	"webscape/server/command"
	"webscape/server/config"
	"webscape/server/game"
	"webscape/server/game/model"
	"webscape/server/game/world"
	"webscape/server/message"
)

func newCommandHandlerTestGame(t *testing.T) *game.Game {
	t.Helper()
	gameWorld, err := world.LoadFromGameFS(fstest.MapFS{
		"game.json":  {Data: []byte(`{"formatVersion":2,"id":"test","world":{"chunkSize":{"x":1,"y":1}},"files":{"chunks":["chunk.json"],"conversations":[],"quests":[]}}`)},
		"chunk.json": {Data: []byte(`{"formatVersion":2,"id":"chunk","coordinate":{"x":0,"y":0},"terrain":["grass"],"heights":[0],"blockers":[false],"walls":[],"entities":[{"id":"spawn","components":{"position":{"x":0,"y":0},"playerSpawn":{}}}]}`)},
	})
	if err != nil {
		t.Fatal(err)
	}
	testGame := game.NewGameWithWorld(gameWorld)
	testGame.RegisterSender(func(string, message.Message) {})
	return testGame
}

func TestCommandHandlerIgnoresGameplayBeforeRegistration(t *testing.T) {
	testGame := newCommandHandlerTestGame(t)
	handler := NewClientCommandHandler(testGame, func(string) (model.EntityId, string, bool) { return model.EntityId{}, "", false })

	for _, commandType := range []command.CommandType{
		command.CommandTypeMove,
		command.CommandTypeChat,
		command.CommandTypeInteract,
		command.CommandTypeEquip,
		command.CommandTypeUnequip,
		command.CommandTypeDrop,
		command.CommandTypeConversationOption,
	} {
		handler.HandleCommand("client", command.Command{Type: commandType, Data: map[string]any{}})
	}

	if testGame.IsRegistered("client") {
		t.Fatal("pre-registration gameplay command registered the client")
	}
}

func TestCommandHandlerRejectsInvalidDropPayloadsWithoutPanicking(t *testing.T) {
	testGame := newCommandHandlerTestGame(t)
	handler := NewClientCommandHandler(testGame, func(string) (model.EntityId, string, bool) { return model.EntityId{}, "", false })
	testGame.HandleRegister("client", model.NewEntityId(), "Player")
	for _, payload := range []map[string]any{
		nil, {}, {"itemId": nil}, {"itemId": 42}, {"itemId": "bad-id"}, {"itemId": model.NewItemId().String()},
	} {
		handler.HandleCommand("client", command.Command{Type: command.CommandTypeDrop, Data: payload})
	}
}

func TestMalformedTradeCommandsDoNotPanic(t *testing.T) {
	testGame := newCommandHandlerTestGame(t)
	testGame.HandleRegister("client", model.NewEntityId(), "Player")
	handler := NewClientCommandHandler(testGame, func(string) (model.EntityId, string, bool) { return model.EntityId{}, "", false })
	for _, typ := range []command.CommandType{command.CommandTypeTrade, command.CommandTypeTradeClose} {
		for _, data := range []map[string]any{
			nil, {}, {"targetEntityId": 42}, {"targetEntityId": "bad"},
			{"targetEntityId": model.NewEntityId().String()},
			{"targetEntityId": model.NewEntityId().String(), "action": "buy", "itemId": 42},
			{"targetEntityId": model.NewEntityId().String(), "action": "hack", "itemId": "ironSword"},
		} {
			handler.HandleCommand("client", command.Command{Type: typ, Data: data})
		}
	}
}

func TestAdminChatUsesRegisteredIdentityNotPayload(t *testing.T) {
	g := newCommandHandlerTestGame(t)
	owner, other := model.NewEntityId(), model.NewEntityId()
	g.HandleRegister("owner", owner, "Owner")
	g.HandleRegister("other", other, "Other")
	enabled := true
	g.ConfigureAdminCommands(config.AdminCommandsConfig{Enabled: &enabled, PlayerIDs: []string{owner.String()}}, false)
	handler := NewClientCommandHandler(g, func(string) (model.EntityId, string, bool) { return owner, "Owner", true })
	before, err := g.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	// Neither an allowlisted ID in the payload nor the identity resolver overrides
	// the player already registered to this connection.
	handler.HandleCommand("other", command.Command{Type: command.CommandTypeChat, Data: map[string]any{"message": "/reset", "id": owner.String(), "playerId": owner.String()}})
	handler.HandleCommand("unregistered", command.Command{Type: command.CommandTypeChat, Data: map[string]any{"message": "/reset"}})
	after, err := g.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Fatal("unauthorized reset changed character state")
	}
	handler.HandleCommand("owner", command.Command{Type: command.CommandTypeChat, Data: map[string]any{"message": "/reset"}})
	after, err = g.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(before, after) {
		t.Fatal("authorized chat reset did not replace character items")
	}
}
