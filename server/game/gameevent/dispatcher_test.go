package gameevent

import (
	"testing"
	"webscape/server/game/model"
)

func TestDispatcherNotifiesRegisteredHandlersInOrder(t *testing.T) {
	dispatcher := NewDispatcher()
	calls := []string{}
	dispatcher.Register(HandlerFunc(func(event Event) {
		calls = append(calls, "first:"+event.Id)
	}))
	dispatcher.Register(HandlerFunc(func(event Event) {
		calls = append(calls, "second:"+event.Id)
	}))

	dispatcher.Emit(New("test:event", model.NewEntityId()))

	if len(calls) != 2 || calls[0] != "first:test:event" || calls[1] != "second:test:event" {
		t.Fatalf("handler calls = %#v", calls)
	}
}

func TestDispatcherUsesHandlerSnapshotDuringEmission(t *testing.T) {
	dispatcher := NewDispatcher()
	calls := 0
	dispatcher.Register(HandlerFunc(func(Event) {
		calls++
		dispatcher.Register(HandlerFunc(func(Event) { calls++ }))
	}))

	dispatcher.Emit(New("first", model.NewEntityId()))
	if calls != 1 {
		t.Fatalf("calls after first emission = %d, want 1", calls)
	}

	dispatcher.Emit(New("second", model.NewEntityId()))
	if calls != 3 {
		t.Fatalf("calls after second emission = %d, want 3", calls)
	}
}

func TestDispatcherOnlyNotifiesMatchingSubscribers(t *testing.T) {
	dispatcher := NewDispatcher()
	calls := 0
	dispatcher.Subscribe("wanted", HandlerFunc(func(Event) { calls++ }))

	dispatcher.Emit(New("other", model.NewEntityId()))
	dispatcher.Emit(New("wanted", model.NewEntityId()))

	if calls != 1 {
		t.Fatalf("subscriber calls = %d, want 1", calls)
	}
}
