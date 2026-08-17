package component

import (
	"testing"
	"webscape/server/game/model"
	"webscape/server/util"
)

func TestFacingSerialize(t *testing.T) {
	targetEntityId := model.NewEntityId()
	facing := NewCFacing(targetEntityId)

	if facing.GetId() != ComponentIdFacing {
		t.Fatalf("component id = %q, want %q", facing.GetId(), ComponentIdFacing)
	}
	if facing.GetTargetEntityId() != targetEntityId {
		t.Fatalf("target entity id = %q, want %q", facing.GetTargetEntityId(), targetEntityId)
	}

	serialized := facing.Serialize()
	object, ok := serialized.(util.JObject)
	if !ok {
		t.Fatalf("serialized facing = %#v, want object", serialized)
	}
	if object["targetEntityId"] != util.JString(targetEntityId.String()) {
		t.Fatalf("serialized targetEntityId = %#v, want %q", object["targetEntityId"], targetEntityId)
	}
}
