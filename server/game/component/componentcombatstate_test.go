package component

import (
	"testing"
	"webscape/server/game/model"
	"webscape/server/util"
)

func TestCombatStateSerializesSynchronizedMagicCast(t *testing.T) {
	targetId := model.NewEntityId()
	stats := NewCCombatStats(1, 2, 3, 4, 5, 0.1, 1.5, 4, 3)
	stats.SetAttackProfile(model.AttackMethodMagic, 2, 1, "magicBolt")
	state := NewCCombatState(targetId)
	state.Restart(stats, 40)
	state.BeginCasting(40, 43)

	serialized := state.Serialize().(util.JObject)
	if serialized["targetEntityId"] != util.JString(targetId.String()) ||
		serialized["phase"] != util.JString("casting") ||
		serialized["phaseStartedTick"] != util.JNumber(40) ||
		serialized["attackMethod"] != util.JString("magic") ||
		serialized["windUpTicks"] != util.JNumber(2) {
		t.Fatalf("serialized combat state = %#v", serialized)
	}
}
