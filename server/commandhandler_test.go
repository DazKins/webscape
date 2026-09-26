package server

import (
	"bytes"
	"encoding/json"
	"math"
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
		command.CommandTypeInventoryMove,
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

func TestMalformedBankCommandsDoNotChangeItems(t *testing.T) {
	g := newCommandHandlerTestGame(t)
	g.HandleRegister("player", model.NewEntityId(), "Player")
	handler := NewClientCommandHandler(g, func(string) (model.EntityId, string, bool) { return model.EntityId{}, "", false })
	before, err := g.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	for _, typ := range []command.CommandType{command.CommandTypeBankDeposit, command.CommandTypeBankWithdraw, command.CommandTypeBankClose} {
		for _, data := range []map[string]any{nil, {}, {"targetEntityId": 42}, {"targetEntityId": "bad"}, {"targetEntityId": model.NewEntityId().String(), "itemId": false}} {
			handler.HandleCommand("player", command.Command{Type: typ, Data: data})
		}
		for _, quantity := range []any{nil, "1", -1.0, 0.0, 0.5, 1.0, float64(model.MaxStackQuantity) + 1, math.NaN(), math.Inf(1)} {
			handler.HandleCommand("player", command.Command{Type: typ, Data: map[string]any{"targetEntityId": model.NewEntityId().String(), "itemId": model.NewItemId().String(), "quantity": quantity}})
		}
	}
	after, err := g.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Fatal("invalid bank commands changed stored state")
	}
}

func TestMalformedInventoryMovesDoNotChangeState(t *testing.T) {
	g := newCommandHandlerTestGame(t)
	g.HandleRegister("player", model.NewEntityId(), "Player")
	handler := NewClientCommandHandler(g, nil)
	before, err := g.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	for _, data := range []map[string]any{nil, {}, {"itemId": 42}, {"itemId": "bad", "slot": 0.0}} {
		handler.HandleCommand("player", command.Command{Type: command.CommandTypeInventoryMove, Data: data})
	}
	for _, slot := range []any{nil, "1", -1.0, 20.0, 0.5, math.NaN(), math.Inf(1)} {
		handler.HandleCommand("player", command.Command{Type: command.CommandTypeInventoryMove, Data: map[string]any{"itemId": model.NewItemId().String(), "slot": slot}})
	}
	after, err := g.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Fatal("invalid moves changed state")
	}
}

func TestInventoryMoveCommandValidatesSequence(t *testing.T) {
	g := newCommandHandlerTestGame(t)
	var itemID string
	g.RegisterSender(func(_ string, msg message.Message) {
		if msg.Metadata.Type != message.MessageTypeGameUpdate {
			return
		}
		raw, _ := json.Marshal(msg.Data)
		var update struct {
			Entities []struct {
				ComponentId string `json:"componentId"`
				Data        struct {
					Items []struct {
						ID string `json:"id"`
					} `json:"items"`
				} `json:"data"`
			} `json:"entities"`
		}
		if err := json.Unmarshal(raw, &update); err != nil {
			t.Fatal(err)
		}
		for _, entity := range update.Entities {
			if entity.ComponentId == "inventory" && len(entity.Data.Items) > 0 {
				itemID = entity.Data.Items[0].ID
			}
		}
	})
	g.HandleRegister("player", model.NewEntityId(), "Player")
	if itemID == "" {
		t.Fatal("missing inventory")
	}
	handler := NewClientCommandHandler(g, nil)
	before, err := g.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	for _, sequence := range []any{nil, "1", 0.0, -1.0, 0.5, math.NaN(), math.Inf(1), 9007199254740992.0} {
		handler.HandleCommand("player", command.Command{Type: command.CommandTypeInventoryMove, Data: map[string]any{"itemId": itemID, "slot": 19.0, "sequence": sequence}})
	}
	after, err := g.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Fatal("invalid sequence changed inventory")
	}
	handler.HandleCommand("player", command.Command{Type: command.CommandTypeInventoryMove, Data: map[string]any{"itemId": itemID, "slot": 19.0, "sequence": 1.0}})
	after, err = g.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(before, after) {
		t.Fatal("valid sequenced move was ignored")
	}
}
