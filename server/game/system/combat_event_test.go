package system

import (
	"testing"
	"webscape/server/game/component"
	"webscape/server/game/gameevent"
	"webscape/server/game/model"
	"webscape/server/util"
)

type recordingEventEmitter struct {
	events []gameevent.Event
}

func (r *recordingEventEmitter) EmitGameEvent(event gameevent.Event) {
	r.events = append(r.events, event)
}

func TestCombatSystemEmitsKillEventsForPlayerKill(t *testing.T) {
	manager := component.NewComponentManager()
	playerId := model.NewEntityId()
	targetId := model.NewEntityId()
	emitter := &recordingEventEmitter{}
	system := &CombatSystem{
		SystemBase:   SystemBase{ComponentManager: manager},
		EventEmitter: emitter,
	}

	manager.SetEntityComponent(playerId, component.NewCPlayer("player"))
	manager.SetEntityComponent(targetId, component.NewCMetadata(util.JObject(map[string]util.Json{
		"name":       util.JString("Cave Rat"),
		"entityType": util.JString("rat"),
	})))

	system.emitKillEvents(playerId, targetId)

	if len(emitter.events) != 2 {
		t.Fatalf("emitted %d events, want 2", len(emitter.events))
	}
	assertEventId(t, emitter.events, "kill:entity:rat")
	assertEventId(t, emitter.events, "kill:name:cave_rat")
}

func TestCombatSystemDoesNotEmitKillEventsForNonPlayerKill(t *testing.T) {
	manager := component.NewComponentManager()
	attackerId := model.NewEntityId()
	targetId := model.NewEntityId()
	emitter := &recordingEventEmitter{}
	system := &CombatSystem{
		SystemBase:   SystemBase{ComponentManager: manager},
		EventEmitter: emitter,
	}

	manager.SetEntityComponent(targetId, component.NewCMetadata(util.JObject(map[string]util.Json{
		"name":       util.JString("Cave Rat"),
		"entityType": util.JString("rat"),
	})))

	system.emitKillEvents(attackerId, targetId)

	if len(emitter.events) != 0 {
		t.Fatalf("emitted %d events, want 0", len(emitter.events))
	}
}

func TestCombatSystemEmitsTypedCombatResult(t *testing.T) {
	emitter := &recordingEventEmitter{}
	system := &CombatSystem{EventEmitter: emitter}
	attackerId := model.NewEntityId()
	targetId := model.NewEntityId()

	system.emitCombatResolved(attackerId, targetId, attackResult{
		DidHit: true,
		Damage: 12,
		IsCrit: true,
	})

	if len(emitter.events) != 1 {
		t.Fatalf("emitted %d events, want 1", len(emitter.events))
	}
	event := emitter.events[0]
	payload, ok := event.Payload.(gameevent.CombatResolvedPayload)
	if event.Id != gameevent.EventIdCombatResolved || event.ActorEntityId != attackerId ||
		event.TargetEntityId != targetId || !ok || !payload.DidHit || payload.Damage != 12 || !payload.IsCritical {
		t.Fatalf("event = %#v", event)
	}
}

func assertEventId(t *testing.T, events []gameevent.Event, id string) {
	t.Helper()
	for _, event := range events {
		if event.Id == id {
			return
		}
	}
	t.Fatalf("event %q was not emitted; got %#v", id, events)
}
