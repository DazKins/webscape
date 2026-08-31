package component

import (
	"testing"
	"webscape/server/game/model"
	"webscape/server/math"
	"webscape/server/util"
)

func TestWoodcuttingSerializationExcludesOriginPosition(t *testing.T) {
	targetEntityId := model.NewEntityId()
	woodcutting := NewCWoodcutting(targetEntityId, 17, math.Vec2{X: 3, Y: 4})
	woodcutting.StartPhase(WoodcuttingPhaseRecovering, 18)

	want := util.JObject{
		"targetEntityId":   util.JString(targetEntityId.String()),
		"phase":            util.JString(WoodcuttingPhaseRecovering),
		"phaseStartedTick": util.JNumber(18),
	}
	serialized := woodcutting.Serialize()
	if !util.JsonEqual(serialized, want) {
		t.Fatalf("serialized woodcutting = %#v, want %#v", serialized, want)
	}
	if _, exists := serialized.(util.JObject)["originPosition"]; exists {
		t.Fatal("serialized woodcutting exposes its server-only origin position")
	}
}
