package game

import (
	"encoding/json"
	"testing"
	"testing/synctest"
	"time"
	"webscape/server/game/gametime"
	"webscape/server/game/system"
	"webscape/server/game/world"
	"webscape/server/message"
)

type timeObservingSystem struct {
	source system.GameTimeSource
	seen   gametime.State
}

func (s *timeObservingSystem) Update() { s.seen = s.source.CurrentGameTime() }

func TestGameTimeAdvancesOnlyWithExecutedTicksBeforeSystems(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		g := NewGameWithWorld(world.NewWorld(4, 4))
		observer := &timeObservingSystem{source: g}
		g.systems = []system.System{observer}
		g.currentTick = 1199
		time.Sleep(20 * time.Minute)
		if g.CurrentGameTime().IsNight || g.CurrentGameTime().Tick != 1199 {
			t.Fatal("wall time advanced the cycle without an executed tick")
		}
		g.update() // No connected players.
		if observer.seen.Tick != 1200 || !observer.seen.IsNight {
			t.Fatalf("systems saw stale game time: %+v", observer.seen)
		}
		for range 1200 {
			g.update()
		}
		if g.CurrentGameTime().Tick != 2400 || g.CurrentGameTime().IsNight {
			t.Fatalf("cycle did not advance without players: %+v", g.CurrentGameTime())
		}
	})
}

func TestEveryClientReceivesEmptyTickHeartbeatsAndReconnectTime(t *testing.T) {
	g := NewGameWithWorld(loadInterestWorld(t))
	g.systems = nil
	sent := map[string][]message.Message{}
	g.RegisterSender(func(client string, msg message.Message) { sent[client] = append(sent[client], msg) })
	g.currentTick = 1198
	g.HandleConnect("one")
	g.HandleConnect("two")
	for tick := uint64(1199); tick <= 1201; tick++ {
		clear(sent)
		g.update()
		for _, client := range []string{"one", "two"} {
			assertEmptyTickUpdate(t, sent[client], tick)
		}
	}
	g.HandleLeave("one")
	g.update()
	sent["one"] = nil
	g.HandleConnect("one")
	updates := 0
	for _, msg := range sent["one"] {
		if msg.Metadata.Type != message.MessageTypeGameUpdate {
			continue
		}
		updates++
		var payload struct {
			Data struct {
				ServerTick uint64 `json:"serverTick"`
			} `json:"data"`
		}
		if err := json.Unmarshal([]byte(msg.Marshal()), &payload); err != nil {
			t.Fatal(err)
		}
		if payload.Data.ServerTick != 1202 {
			t.Fatalf("reconnect tick = %d", payload.Data.ServerTick)
		}
	}
	if updates != 1 {
		t.Fatalf("reconnect received %d game updates", updates)
	}
}

func assertEmptyTickUpdate(t *testing.T, messages []message.Message, tick uint64) {
	t.Helper()
	assertOnlyMessageType(t, messages, message.MessageTypeGameUpdate)
	var payload struct {
		Data struct {
			ServerTick uint64          `json:"serverTick"`
			Entities   json.RawMessage `json:"entities"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(messages[0].Marshal()), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Data.ServerTick != tick || string(payload.Data.Entities) != "[]" {
		t.Fatalf("expected tick %d with entities [], got %s", tick, messages[0].Marshal())
	}
}

func TestGameTimeResetsAfterRestart(t *testing.T) {
	g := snapshotGame(t)
	g.currentTick = 1387
	restored := snapshotGame(t)
	if err := restored.RestoreSnapshot(savedBytes(t, g)); err != nil {
		t.Fatal(err)
	}
	if got := restored.CurrentGameTime(); got.Tick != 0 || got.IsNight {
		t.Fatalf("game time did not reset: %+v", got)
	}
}
