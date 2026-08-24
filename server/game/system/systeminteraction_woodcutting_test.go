package system

import (
	"testing"
	"webscape/server/game/component"
	"webscape/server/game/gameevent"
	"webscape/server/game/model"
	"webscape/server/math"
	"webscape/server/util"
)

type woodcuttingStarterStub struct {
	started bool
}

func (s *woodcuttingStarterStub) StartWoodcuttingFor(model.EntityId, model.EntityId) bool {
	return s.started
}

type gameEventRecorder struct {
	events []gameevent.Event
}

func (r *gameEventRecorder) EmitGameEvent(event gameevent.Event) {
	r.events = append(r.events, event)
}

func TestChopInteractionEmitsObjectEventOnlyWhenWoodcuttingStarts(t *testing.T) {
	for _, test := range []struct {
		name       string
		starts     bool
		wantEvents int
	}{
		{name: "started", starts: true, wantEvents: 1},
		{name: "rejected", starts: false, wantEvents: 0},
	} {
		t.Run(test.name, func(t *testing.T) {
			manager := component.NewComponentManager()
			playerId := model.NewEntityId()
			treeId := model.NewEntityId()
			manager.SetEntityComponents(
				playerId,
				component.NewCPosition(math.Vec2{X: 0, Y: 0}),
				component.NewCInteracting(treeId, component.InteractionOptionChop),
			)
			manager.SetEntityComponents(
				treeId,
				component.NewCPosition(math.Vec2{X: 1, Y: 0}),
				component.NewCMetadata(util.JObject{
					"objectId": util.JString("tree_001"),
				}),
			)
			recorder := &gameEventRecorder{}
			system := InteractionSystem{
				SystemBase:         SystemBase{ComponentManager: manager},
				EventEmitter:       recorder,
				WoodcuttingStarter: &woodcuttingStarterStub{started: test.starts},
			}

			system.Update()

			if len(recorder.events) != test.wantEvents {
				t.Fatalf("events = %#v, want %d", recorder.events, test.wantEvents)
			}
			if test.wantEvents == 1 {
				event := recorder.events[0]
				if event.Id != "interact:object:tree_001:chop" ||
					event.ActorEntityId != playerId || event.TargetEntityId != treeId {
					t.Fatalf("event = %#v", event)
				}
			}
		})
	}
}
