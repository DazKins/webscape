package message

import (
	"encoding/json"
	"testing"
	"time"
	"webscape/server/game/world"
)

func TestWorldMessageContainsOnlyMetadataAndRegistries(t *testing.T) {
	testWorld, err := world.LoadFromGameFolder("../../game-project")
	if err != nil {
		t.Fatalf("LoadFromGameFolder: %v", err)
	}
	var payload struct {
		Data map[string]json.RawMessage `json:"data"`
	}
	message := NewWorldMessage(testWorld)
	if err := json.Unmarshal([]byte(message.Marshal()), &payload); err != nil {
		t.Fatal(err)
	}
	if _, ok := payload.Data["chunkSize"]; !ok {
		t.Fatal("world metadata omitted chunkSize")
	}
	if _, ok := payload.Data["playerSpawn"]; !ok {
		t.Fatal("world metadata omitted playerSpawn")
	}
	if _, ok := payload.Data["quests"]; !ok {
		t.Fatal("world metadata omitted quests")
	}
	for _, spatial := range []string{"terrain", "heights", "walls", "chunks"} {
		if _, ok := payload.Data[spatial]; ok {
			t.Fatalf("world metadata exposed spatial field %q", spatial)
		}
	}
}

func TestWorldMessageIncludesQuestRewards(t *testing.T) {
	testWorld, err := world.LoadFromGameFolder("../../game-project")
	if err != nil {
		t.Fatal(err)
	}
	var payload struct {
		Data struct {
			Quests []struct {
				Rewards struct {
					Items []struct {
						Name string `json:"name"`
					} `json:"items"`
				} `json:"rewards"`
			} `json:"quests"`
		} `json:"data"`
	}
	message := NewWorldMessage(testWorld)
	if err := json.Unmarshal([]byte(message.Marshal()), &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.Data.Quests) == 0 || len(payload.Data.Quests[0].Rewards.Items) == 0 {
		t.Fatal("quest rewards missing")
	}
}

func TestWorldMessageAdvertisesTickInterval(t *testing.T) {
	message := NewWorldMessage(world.NewWorld(4, 4), 250*time.Millisecond)
	var payload struct {
		Data struct {
			TickIntervalMs int `json:"tickIntervalMs"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(message.Marshal()), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Data.TickIntervalMs != 250 {
		t.Fatalf("interval=%d", payload.Data.TickIntervalMs)
	}
}
