package component

import (
	"testing"
	"webscape/server/game/model"
	"webscape/server/math"
	"webscape/server/util"
)

func TestFishingSerializationExcludesOriginPosition(t *testing.T) {
	targetEntityId := model.NewEntityId()
	fishing := NewCFishing(targetEntityId, 17, math.Vec2{X: 3, Y: 4})
	fishing.StartPhase(FishingPhaseWaiting, 18)

	want := util.JObject{
		"targetEntityId":   util.JString(targetEntityId.String()),
		"phase":            util.JString(FishingPhaseWaiting),
		"phaseStartedTick": util.JNumber(18),
	}
	serialized := fishing.Serialize()
	if !util.JsonEqual(serialized, want) {
		t.Fatalf("serialized fishing = %#v, want %#v", serialized, want)
	}
	if _, exists := serialized.(util.JObject)["originPosition"]; exists {
		t.Fatal("serialized fishing exposes its server-only origin position")
	}
}
