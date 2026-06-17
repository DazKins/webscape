package message

import (
	"encoding/json"
	"testing"
	"testing/fstest"
	"webscape/server/game/world"
)

func TestWorldMessageIncludesQuestRewards(t *testing.T) {
	testWorld, err := world.LoadFromGameFS(fstest.MapFS{
		"game.json": {
			Data: []byte(`{
				"formatVersion": 1,
				"id": "test_game",
				"files": {
					"maps": ["maps/test.json"],
					"conversations": [],
					"quests": ["quests/tutorial.json"]
				}
			}`),
		},
		"maps/test.json": {
			Data: []byte(`{
				"formatVersion": 1,
				"id": "test",
				"size": { "x": 1, "y": 1 },
				"terrain": ["grass"]
			}`),
		},
		"quests/tutorial.json": {
			Data: []byte(`{
				"formatVersion": 2,
				"id": "tutorial_quests",
				"quests": [
					{
						"id": "first_errand",
						"steps": [
							{
								"id": "talk",
								"description": "Talk to the guide.",
								"requirement": { "eventId": "conversation:node:guide:start", "count": 1 }
							}
						],
						"rewards": {
							"items": [
								{ "name": "Ancient Scroll", "type": "quest", "count": 1 }
							]
						}
					}
				]
			}`),
		},
	})
	if err != nil {
		t.Fatalf("LoadFromGameFS returned error: %v", err)
	}

	msg := NewWorldMessage(testWorld)
	var payload struct {
		Data struct {
			Quests []struct {
				Rewards struct {
					Items []struct {
						Name  string `json:"name"`
						Type  string `json:"type"`
						Count int    `json:"count"`
					} `json:"items"`
				} `json:"rewards"`
			} `json:"quests"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(msg.Marshal()), &payload); err != nil {
		t.Fatalf("unmarshal world message: %v", err)
	}
	if len(payload.Data.Quests) != 1 || len(payload.Data.Quests[0].Rewards.Items) != 1 {
		t.Fatalf("quest rewards = %#v, want one reward item", payload.Data.Quests)
	}
	reward := payload.Data.Quests[0].Rewards.Items[0]
	if reward.Name != "Ancient Scroll" || reward.Type != "quest" || reward.Count != 1 {
		t.Fatalf("reward = %#v, want Ancient Scroll quest x1", reward)
	}
}
