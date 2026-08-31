package component

import (
	"testing"
	"webscape/server/util"
)

func TestLocomotionSerializationExcludesLastMovementTick(t *testing.T) {
	locomotion := NewCLocomotion(LocomotionPhaseIdle, 4)
	locomotion.MarkMoving(7)
	locomotion.MarkMoving(8)

	want := util.JObject{
		"phase":            util.JString(LocomotionPhaseMoving),
		"phaseStartedTick": util.JNumber(7),
	}
	serialized := locomotion.Serialize()
	if !util.JsonEqual(serialized, want) {
		t.Fatalf("serialized locomotion = %#v, want %#v", serialized, want)
	}
	if _, exists := serialized.(util.JObject)["lastMovementTick"]; exists {
		t.Fatal("serialized locomotion exposes its server-only last movement tick")
	}
}
