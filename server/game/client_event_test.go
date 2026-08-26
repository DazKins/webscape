package game

import (
	"testing"
	"webscape/server/game/component"
	"webscape/server/game/gameevent"
	"webscape/server/game/model"
	"webscape/server/game/world"
	"webscape/server/message"
)

func TestClientEventsAreProjectedToInterestedClients(t *testing.T) {
	gameWorld := loadInterestWorld(t)
	game := NewGameWithWorldAndChunkRadius(gameWorld, 0)
	sent := map[string][]message.Message{}
	game.RegisterSender(func(clientID string, msg message.Message) {
		sent[clientID] = append(sent[clientID], msg)
	})

	nearPlayerId := model.NewEntityId()
	farPlayerId := model.NewEntityId()
	game.HandleRegister("near-client", nearPlayerId, "Near")
	game.HandleRegister("far-client", farPlayerId, "Far")
	movePlayerForTest(game, farPlayerId, 4, 0)
	game.syncClient("far-client")
	sent = map[string][]message.Message{}

	game.EmitGameEvent(gameevent.NewChatSpoken(nearPlayerId, "hello"))
	if len(sent) != 0 {
		t.Fatalf("client event was sent before the event queue flushed: %#v", sent)
	}
	game.flushPendingClientEvents()

	assertOnlyMessageType(t, sent["near-client"], message.MessageTypeChatMessage)
	if len(sent["far-client"]) != 0 {
		t.Fatalf("distant client received chat event: %v", messageTypes(sent["far-client"]))
	}

	sent = map[string][]message.Message{}
	game.EmitGameEvent(gameevent.NewCombatResolved(nearPlayerId, farPlayerId, true, 7, false))
	game.flushPendingClientEvents()
	assertOnlyMessageType(t, sent["near-client"], message.MessageTypeCombatResolved)
	assertOnlyMessageType(t, sent["far-client"], message.MessageTypeCombatResolved)
}

func TestClientEventsAreSentAfterStateUpdates(t *testing.T) {
	gameWorld := loadInterestWorld(t)
	game := NewGameWithWorldAndChunkRadius(gameWorld, 0)
	sent := []message.Message{}
	game.RegisterSender(func(clientID string, msg message.Message) {
		if clientID == "client" {
			sent = append(sent, msg)
		}
	})

	playerId := model.NewEntityId()
	game.HandleRegister("client", playerId, "Player")
	sent = nil
	health := game.componentManager.GetEntityComponent(component.ComponentIdHealth, playerId).(*component.CHealth)
	health.SetCurrentHealth(health.GetCurrentHealth() - 1)
	game.componentManager.SetEntityComponent(playerId, health)
	game.EmitGameEvent(gameevent.NewCombatResolved(playerId, playerId, true, 1, false))

	game.update()

	types := messageTypes(sent)
	gameUpdateIndex := indexOfMessageType(types, message.MessageTypeGameUpdate)
	combatEventIndex := indexOfMessageType(types, message.MessageTypeCombatResolved)
	if gameUpdateIndex < 0 || combatEventIndex < 0 || gameUpdateIndex >= combatEventIndex {
		t.Fatalf("message order = %v, want gameUpdate before combatResolved", types)
	}
}

func loadInterestWorld(t *testing.T) *world.World {
	t.Helper()
	gameWorld, err := world.LoadFromGameFS(interestWorldFS())
	if err != nil {
		t.Fatal(err)
	}
	return gameWorld
}

func assertOnlyMessageType(t *testing.T, messages []message.Message, want message.MessageType) {
	t.Helper()
	if len(messages) != 1 || messages[0].Metadata.Type != want {
		t.Fatalf("messages = %v, want [%s]", messageTypes(messages), want)
	}
}

func indexOfMessageType(types []message.MessageType, want message.MessageType) int {
	for index, messageType := range types {
		if messageType == want {
			return index
		}
	}
	return -1
}
