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

func TestConversationInteractionRoutesOptionsAndEnds(t *testing.T) {
	game, sent, targetEntityId, playerEntityId := setupConversationTestGame(t)

	game.HandleInteract("client-1", targetEntityId, component.InteractionOptionTalk)
	game.update()

	active := game.componentManager.GetEntityComponent(component.ComponentIdActiveConversation, playerEntityId)
	if active == nil {
		t.Fatal("player does not have active conversation after talk")
	}
	assertConversationMessage(t, *sent, "greeting", targetEntityId.String(), "start", false)

	game.HandleConversationOption("client-1", "greeting", "start", "missing")
	if len(*sent) != 1 {
		t.Fatalf("stale option sent %d conversation messages, want 1", len(*sent))
	}
	activeConversation := game.componentManager.GetEntityComponent(
		component.ComponentIdActiveConversation,
		playerEntityId,
	).(*component.CActiveConversation)
	if activeConversation.GetCurrentNodeId() != "start" {
		t.Fatalf("current node = %q, want start", activeConversation.GetCurrentNodeId())
	}

	game.HandleConversationOption("client-1", "greeting", "start", "bye")
	assertConversationMessage(t, *sent, "greeting", targetEntityId.String(), "end", true)
	if active := game.componentManager.GetEntityComponent(component.ComponentIdActiveConversation, playerEntityId); active != nil {
		t.Fatal("active conversation was not removed on end node")
	}
}

func TestMoveCancelsActiveConversation(t *testing.T) {
	game, sent, targetEntityId, playerEntityId := setupConversationTestGame(t)

	game.HandleInteract("client-1", targetEntityId, component.InteractionOptionTalk)
	game.update()

	if active := game.componentManager.GetEntityComponent(component.ComponentIdActiveConversation, playerEntityId); active == nil {
		t.Fatal("player does not have active conversation after talk")
	}

	game.HandleMove("client-1", 0, 0)
	if active := game.componentManager.GetEntityComponent(component.ComponentIdActiveConversation, playerEntityId); active != nil {
		t.Fatal("active conversation was not removed after move")
	}

	sentCount := len(*sent)
	game.HandleConversationOption("client-1", "greeting", "start", "bye")
	if len(*sent) != sentCount {
		t.Fatalf("cancelled conversation sent %d new messages, want 0", len(*sent)-sentCount)
	}
}

func setupConversationTestGame(t *testing.T) (*Game, *[]message.Message, model.EntityId, model.EntityId) {
	t.Helper()

	testWorld, err := world.LoadFromGameFS(fstest.MapFS{
		"game.json": {
			Data: []byte(`{
				"formatVersion": 1,
				"id": "test_game",
				"files": {
					"maps": ["maps/test.json"],
					"conversations": ["conversations/greeting.json"],
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
						"id": "greeter",
						"components": {
							"position": { "x": 1, "y": 0 },
							"conversation": { "conversationId": "greeting" }
						}
					}
				]
			}`),
		},
		"conversations/greeting.json": {
			Data: []byte(`{
				"formatVersion": 1,
				"id": "greeting_doc",
				"conversations": [
					{
						"id": "greeting",
						"startNodeId": "start",
						"nodes": [
							{
								"id": "start",
								"messages": [{ "text": "Hello." }],
								"options": [{ "id": "bye", "text": "Bye.", "nextNodeId": "end" }]
							},
							{
								"id": "end",
								"messages": [{ "text": "Goodbye." }],
								"endConversation": true
							}
						]
					}
				]
			}`),
		},
	})
	if err != nil {
		t.Fatalf("LoadFromGameFS returned error: %v", err)
	}

	game := NewGameWithWorld(testWorld)
	game.RegisterBroadcaster(func(message.Message) {})
	sent := []message.Message{}
	game.RegisterSender(func(clientID string, msg message.Message) {
		if clientID == "client-1" {
			sent = append(sent, msg)
		}
	})

	targetEntityId, ok := firstEntityWithComponent(game, component.ComponentIdConversation)
	if !ok {
		t.Fatal("no NPC entity has CConversation")
	}

	playerEntityId := model.NewEntityId()
	game.HandleJoin("client-1", playerEntityId, "player")
	sent = nil

	return game, &sent, targetEntityId, playerEntityId
}

func firstEntityWithComponent(game *Game, componentId component.ComponentId) (model.EntityId, bool) {
	for entityId := range game.componentManager.GetComponent(componentId) {
		return entityId, true
	}
	return model.EntityId{}, false
}

func assertConversationMessage(
	t *testing.T,
	messages []message.Message,
	conversationId string,
	targetEntityId string,
	nodeId string,
	endConversation bool,
) {
	t.Helper()

	if len(messages) == 0 {
		t.Fatal("no messages were sent")
	}

	var payload struct {
		Metadata struct {
			Type string `json:"type"`
		} `json:"metadata"`
		Data struct {
			ConversationId  string `json:"conversationId"`
			TargetEntityId  string `json:"targetEntityId"`
			NodeId          string `json:"nodeId"`
			EndConversation bool   `json:"endConversation"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(messages[len(messages)-1].Marshal()), &payload); err != nil {
		t.Fatalf("could not unmarshal message: %v", err)
	}

	if payload.Metadata.Type != "conversation" {
		t.Fatalf("message type = %q, want conversation", payload.Metadata.Type)
	}
	if payload.Data.ConversationId != conversationId ||
		payload.Data.TargetEntityId != targetEntityId ||
		payload.Data.NodeId != nodeId ||
		payload.Data.EndConversation != endConversation {
		t.Fatalf(
			"conversation message = (%q, %q, %q, %v), want (%q, %q, %q, %v)",
			payload.Data.ConversationId,
			payload.Data.TargetEntityId,
			payload.Data.NodeId,
			payload.Data.EndConversation,
			conversationId,
			targetEntityId,
			nodeId,
			endConversation,
		)
	}
}
