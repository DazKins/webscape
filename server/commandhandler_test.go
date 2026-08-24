package server

import (
	"testing"
	"testing/fstest"
	"webscape/server/command"
	"webscape/server/game"
	"webscape/server/game/world"
	"webscape/server/message"
)

func TestCommandHandlerIgnoresGameplayBeforeRegistration(t *testing.T) {
	gameWorld, err := world.LoadFromGameFS(fstest.MapFS{
		"game.json":  {Data: []byte(`{"formatVersion":2,"id":"test","world":{"chunkSize":{"x":1,"y":1}},"files":{"chunks":["chunk.json"],"conversations":[],"quests":[]}}`)},
		"chunk.json": {Data: []byte(`{"formatVersion":2,"id":"chunk","coordinate":{"x":0,"y":0},"terrain":["grass"],"heights":[0],"blockers":[false],"walls":[],"entities":[{"id":"spawn","components":{"position":{"x":0,"y":0},"playerSpawn":{}}}]}`)},
	})
	if err != nil {
		t.Fatal(err)
	}
	testGame := game.NewGameWithWorld(gameWorld)
	testGame.RegisterSender(func(string, message.Message) {})
	handler := NewClientCommandHandler(testGame)

	for _, commandType := range []command.CommandType{
		command.CommandTypeMove,
		command.CommandTypeChat,
		command.CommandTypeInteract,
		command.CommandTypeEquip,
		command.CommandTypeUnequip,
		command.CommandTypeConversationOption,
	} {
		handler.HandleCommand("client", command.Command{Type: commandType, Data: map[string]any{}})
	}

	if testGame.IsRegistered("client") {
		t.Fatal("pre-registration gameplay command registered the client")
	}
}
