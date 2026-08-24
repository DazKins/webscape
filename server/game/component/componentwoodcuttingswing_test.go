package component

import (
	"testing"
	"webscape/server/game/model"
	"webscape/server/util"
)

func TestWoodcuttingSwingSerializesPlayerAndTarget(t *testing.T) {
	playerId := model.NewEntityId()
	targetId := model.NewEntityId()
	serialized := NewCWoodcuttingSwing(playerId, targetId).Serialize().(util.JObject)
	if serialized["playerEntityId"] != util.JString(playerId.String()) ||
		serialized["targetEntityId"] != util.JString(targetId.String()) {
		t.Fatalf("serialized swing = %#v", serialized)
	}
}
