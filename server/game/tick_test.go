package game

import (
	"fmt"
	"strings"
	"testing"
	"testing/synctest"
	"time"
	"webscape/server/game/system"
	"webscape/server/game/world"
)

func TestTickScheduleRecoversWithoutDroppingElapsedTime(t *testing.T) {
	base := time.Now()
	s := newTickSchedule(base, 500*time.Millisecond)
	var logs []string
	logf := func(format string, args ...any) { logs = append(logs, fmt.Sprintf(format, args...)) }
	started := base.Add(500 * time.Millisecond)
	finished := started.Add(1200 * time.Millisecond)
	s.complete(started, finished, tickTimings{tick: 1}, logf)
	if s.delay(finished) != 0 {
		t.Fatal("overdue tick was delayed")
	}
	// At t=1.7s ticks 2 and 3 remain due. Both run and consume only 500ms
	// of schedule debt each, despite taking 50ms of CPU time.
	for tick := uint64(2); tick <= 3; tick++ {
		started = finished
		finished = started.Add(50 * time.Millisecond)
		s.complete(started, finished, tickTimings{tick: tick}, logf)
	}
	if got := s.delay(finished); got != 200*time.Millisecond {
		t.Fatalf("delay=%s want=200ms", got)
	}
	joined := strings.Join(logs, "\n")
	for _, want := range []string{"tick_overrun tick=1", "budget=500ms", "tick_recovery_started", "pending_ticks=2", "tick_recovery_finished tick=3 recovery_ticks=2"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("missing %q in %s", want, joined)
		}
	}
}

func TestTickScheduleIncludesWorkAndDoesNotAccumulateTimerJitter(t *testing.T) {
	base := time.Now()
	s := newTickSchedule(base, 500*time.Millisecond)
	logs := 0
	logf := func(string, ...any) { logs++ }
	for tick := uint64(1); tick <= 100; tick++ {
		started := base.Add(time.Duration(tick)*500*time.Millisecond + time.Millisecond)
		finished := started.Add(100 * time.Millisecond)
		s.complete(started, finished, tickTimings{tick: tick}, logf)
		if got := s.delay(finished); got != 399*time.Millisecond {
			t.Fatalf("drift at tick %d: %s", tick, got)
		}
	}
	if logs != 0 {
		t.Fatalf("normal ticks produced %d warning logs", logs)
	}
}

func TestTickScheduleReportsSchedulerStallsAndSustainedBacklog(t *testing.T) {
	base := time.Now()
	s := newTickSchedule(base, 500*time.Millisecond)
	var logs []string
	logf := func(f string, a ...any) { logs = append(logs, fmt.Sprintf(f, a...)) }
	// A delayed start can miss deadlines even with a fast update.
	started := base.Add(3 * time.Second)
	s.complete(started, started.Add(time.Millisecond), tickTimings{tick: 1}, logf)
	if len(logs) != 1 || !strings.Contains(logs[0], "tick_recovery_started") {
		t.Fatalf("logs=%v", logs)
	}
	// Sustained overload retains all debt and emits periodic recovery progress.
	for tick := uint64(2); tick <= 8; tick++ {
		started = started.Add(time.Second)
		s.complete(started, started.Add(time.Second), tickTimings{tick: tick, lockWait: time.Second}, logf)
	}
	if s.next != base.Add(4500*time.Millisecond) {
		t.Fatal("backlog was dropped")
	}
	if !strings.Contains(strings.Join(logs, "\n"), "tick_recovery_progress") {
		t.Fatalf("no progress log: %v", logs)
	}
	if !strings.Contains(strings.Join(logs, "\n"), "lock_wait=1s") {
		t.Fatal("missing lock timing")
	}
}

// Fake time exercises the actual goroutine/timer loop without wall-clock sleeps.
type delayedTickSystem struct {
	calls int
	delay time.Duration
	every bool
}

func (s *delayedTickSystem) Update() {
	s.calls++
	if s.calls == 1 || s.every {
		time.Sleep(s.delay)
	}
}

func TestUpdateLoopRunsRecoveryStepsAndResumesCadence(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		g := NewGameWithWorld(world.NewWorld(4, 4))
		slow := &delayedTickSystem{delay: 1200 * time.Millisecond}
		g.systems = []system.System{slow}
		g.StartUpdateLoop(500 * time.Millisecond)
		synctest.Wait()
		time.Sleep(1800 * time.Millisecond)
		synctest.Wait()
		g.stateMutex.Lock()
		tick := g.currentTick
		g.stateMutex.Unlock()
		if tick != 3 {
			t.Fatalf("recovered tick=%d want=3", tick)
		}
		time.Sleep(200 * time.Millisecond)
		synctest.Wait()
		g.stateMutex.Lock()
		tick = g.currentTick
		g.stateMutex.Unlock()
		if tick != 4 {
			t.Fatalf("resumed tick=%d want=4", tick)
		}
		g.Stop()
	})
}

func TestUpdateLoopCanStopDuringContinuousRecovery(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		g := NewGameWithWorld(world.NewWorld(4, 4))
		slow := &delayedTickSystem{delay: 600 * time.Millisecond, every: true}
		g.systems = []system.System{slow}
		g.StartUpdateLoop(500 * time.Millisecond)
		synctest.Wait()
		time.Sleep(1200 * time.Millisecond)
		// Stop is serviced after the current full tick, even though the next is due.
		g.Stop()
		synctest.Wait()
		if slow.calls != 2 {
			t.Fatalf("calls after stop=%d want=2", slow.calls)
		}
	})
}
