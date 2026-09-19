package component

import (
	"encoding/json"
	"reflect"
	"testing"
	"webscape/server/game/model"
	"webscape/server/util"
)

func TestSavePreservesPrivateState(t *testing.T) {
	log := NewCCombatLog(2)
	log.AddEntry(NewCombatLogEntry("hit", "damage"))
	quest := NewCQuestLog()
	quest.SetProgress("active", 2, "third", 3)
	quest.CompleteQuest("done")
	for _, original := range []Component{
		log, quest,
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

func TestWorldOnlyComponentsHaveNoPersistenceCodecs(t *testing.T) {
	for _, c := range []Component{&CBanker{}, &CSpawn{}, &COpenable{}, &CShop{}, &CFishable{}, &CWoodcuttable{}, &CLootable{}, &CRandomWalk{}, &CConversation{}, &CDroppedItem{}, &CRewardDrop{}} {
		if _, err := SaveComponent(c); err == nil {
			t.Fatalf("world component %s is still saved", c.GetId())
		}
		if _, err := RestoreComponent(c.GetId(), SavedComponent{Version: 1, Data: json.RawMessage(`{}`)}); err == nil {
			t.Fatalf("world component %s is still restored", c.GetId())
		}
	}
}
