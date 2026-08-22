package component

import (
	"reflect"
	"testing"
	"webscape/server/game/model"
	"webscape/server/util"
)

func TestCombatTextSerializeIncludesAttacker(t *testing.T) {
	targetId := model.NewEntityId()
	attackerId := model.NewEntityId()

	serialized := NewCCombatText(targetId, attackerId, "MISS", "miss").Serialize()
	want := util.JObject{
		"fromEntityId":     util.JString(targetId.String()),
		"attackerEntityId": util.JString(attackerId.String()),
		"text":             util.JString("MISS"),
		"kind":             util.JString("miss"),
	}

	if !reflect.DeepEqual(serialized, want) {
		t.Fatalf("unexpected combat text serialization: got %#v, want %#v", serialized, want)
	}
}
