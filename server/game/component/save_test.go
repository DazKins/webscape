package component

import (
	"encoding/json"
	"reflect"
	"testing"
	"webscape/server/game/model"
	"webscape/server/math"
	"webscape/server/util"
)

func TestSavePreservesPrivateState(t *testing.T) {
	spawn := NewCSpawn(math.Vec2{X: 7, Y: 9}, 13, "template", map[string]any{"renderable": map[string]any{"type": "human"}})
	spawn.SetChildEntityId(model.NewEntityId())
	spawn.MarkSpawned()
	spawn.SetRemainingRespawnTicks(4)
	log := NewCCombatLog(2)
	log.AddEntry(NewCombatLogEntry("hit", "damage"))
	quest := NewCQuestLog()
	quest.SetProgress("active", 2, "third", 3)
	quest.CompleteQuest("done")
	for _, original := range []Component{
		spawn, log, quest,
		&CWoodcuttable{maxDurability: 9, currentDurability: 0, respawnTicks: 11, remainingRespawnTicks: 4, depleted: true, yield: LootItem{"Oak", "logs", 3}, lastFellerEntityId: model.NewEntityId()},
		&CLootable{once: true, looted: true, items: []LootItem{{"Gold", "gold", 5}}},
		&CRandomWalk{walkTimer: 3, maxDistance: 4, origin: math.Vec2{X: 6, Y: 7}, hasOrigin: true},
		&CMetadata{metadata: util.JObject{"nested": util.JArray{util.JString("name"), util.JNull{}, util.JNumber(3)}}},
	} {
		t.Run(string(original.GetId()), func(t *testing.T) {
			saved, err := SaveComponent(original)
			if err != nil {
				t.Fatal(err)
			}
			restored, err := RestoreComponent(original.GetId(), saved)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(original, restored) {
				t.Fatalf("private component state changed: %#v -> %#v", original, restored)
			}
		})
	}
}

type unsavable struct{}

func (*unsavable) GetId() ComponentId { return "future-component" }
func TestSaveRejectsUnknownAndMalformedComponents(t *testing.T) {
	if _, err := SaveComponent(&unsavable{}); err == nil {
		t.Fatal("unknown component silently omitted")
	}
	for _, raw := range []string{`{}`, `null`, `{"maxHealth":100}`, `{"maxHealth":100,"currentHealth":50,"extra":1}`, `{} {}`} {
		if _, err := RestoreComponent(ComponentIdHealth, SavedComponent{Version: 1, Data: json.RawMessage(raw)}); err == nil {
			t.Fatalf("accepted %s", raw)
		}
	}
	if _, err := RestoreComponent(ComponentIdHealth, SavedComponent{Version: 2, Data: json.RawMessage(`{}`)}); err == nil {
		t.Fatal("unknown version accepted")
	}
	saved, err := SaveComponent(NewCFacing(model.NewEntityId()))
	if err != nil || saved.Version != 0 {
		t.Fatal("transient facing persisted")
	}
}
