package game

import (
	"testing"
	"testing/fstest"
	"webscape/server/game/component"
	"webscape/server/game/gameevent"
	"webscape/server/game/model"
	"webscape/server/game/world"
	"webscape/server/message"
)

func TestQuestProgressAdvancesFromGenericEvents(t *testing.T) {
	testWorld, err := world.LoadFromGameFS(fstest.MapFS{
		"game.json": {
			Data: []byte(`{
				"formatVersion": 1,
				"id": "test_game",
				"files": {
					"maps": ["maps/test.json"],
					"conversations": ["conversations/guide.json"],
					"quests": ["quests/tutorial.json"]
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
						"id": "guide",
						"components": {
							"position": { "x": 1, "y": 0 },
							"conversation": { "conversationId": "guide" }
						}
					}
				]
			}`),
		},
		"conversations/guide.json": {
			Data: []byte(`{
				"formatVersion": 1,
				"id": "guide_doc",
				"conversations": [
					{
						"id": "guide",
						"startNodeId": "start",
						"nodes": [
							{
								"id": "start",
								"messages": [{ "text": "Help me." }],
								"endConversation": true
							}
						]
					}
				]
			}`),
		},
		"quests/tutorial.json": {
			Data: []byte(`{
				"formatVersion": 1,
				"id": "tutorial_quests",
				"quests": [
					{
						"id": "first_errand",
						"startEventId": "conversation:node:guide:start",
						"steps": [
							{
								"id": "talk",
								"description": "Talk.",
								"requirement": { "eventId": "conversation:node:guide:start", "count": 1 }
							},
							{
								"id": "kill_rats",
								"description": "Kill rats.",
								"requirement": { "eventId": "kill:entity:rat", "count": 2 }
							},
							{
								"id": "collect_scroll",
								"description": "Collect a scroll.",
								"requirement": { "eventId": "collect:name:ancient_scroll", "count": 1 }
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
	game.RegisterSender(func(string, message.Message) {})

	playerEntityId := model.NewEntityId()
	game.HandleJoin("client-1", playerEntityId, "player")

	npcEntityId, ok := firstEntityWithComponent(game, component.ComponentIdConversation)
	if !ok {
		t.Fatal("no NPC entity has CConversation")
	}

	game.HandleInteract("client-1", npcEntityId, component.InteractionOptionTalk)
	game.update()
	assertQuestProgress(t, game, playerEntityId, "first_errand", "kill_rats", 1, 0, false)

	game.EmitGameEvent(gameevent.New("kill:entity:goblin", playerEntityId))
	assertQuestProgress(t, game, playerEntityId, "first_errand", "kill_rats", 1, 0, false)

	game.EmitGameEvent(gameevent.New("kill:entity:rat", playerEntityId))
	assertQuestProgress(t, game, playerEntityId, "first_errand", "kill_rats", 1, 1, false)

	game.EmitGameEvent(gameevent.New("kill:entity:rat", playerEntityId))
	assertQuestProgress(t, game, playerEntityId, "first_errand", "collect_scroll", 2, 0, false)

	game.AddItemToPlayerInventory(playerEntityId, model.CreateAncientScroll())
	assertQuestProgress(t, game, playerEntityId, "first_errand", "", 0, 0, true)
}

func assertQuestProgress(
	t *testing.T,
	game *Game,
	playerEntityId model.EntityId,
	questId string,
	stepId string,
	stepIndex int,
	count int,
	completed bool,
) {
	t.Helper()

	questLogComponent := game.componentManager.GetEntityComponent(component.ComponentIdQuestLog, playerEntityId)
	if questLogComponent == nil {
		t.Fatal("player does not have questlog")
	}
	questLog := questLogComponent.(*component.CQuestLog)
	if questLog.IsCompleted(questId) != completed {
		t.Fatalf("quest completed = %v, want %v", questLog.IsCompleted(questId), completed)
	}
	if completed {
		return
	}

	var progress *component.QuestProgress
	for _, active := range questLog.GetActiveProgress() {
		if active.QuestId == questId {
			progress = active
			break
		}
	}
	if progress == nil {
		t.Fatalf("quest %q is not active", questId)
	}
	if progress.StepId != stepId || progress.CurrentStepIndex != stepIndex || progress.CurrentCount != count {
		t.Fatalf(
			"quest progress = (%q, %d, %d), want (%q, %d, %d)",
			progress.StepId,
			progress.CurrentStepIndex,
			progress.CurrentCount,
			stepId,
			stepIndex,
			count,
		)
	}
}
