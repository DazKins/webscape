package game

import (
	"log"
	"time"
	"webscape/server/game/system"
)

const DefaultTickInterval = 500 * time.Millisecond
const recoveryLogInterval = 5 * time.Second

type tickTimings struct {
	tick                            uint64
	lockWait, systems, sync, events time.Duration
	slowestSystem                   system.System
	slowestSystemDuration           time.Duration
}

// tickSchedule uses monotonic deadlines as an elapsed-time accumulator. Each
// completed step consumes exactly one interval, never the time spent doing work.
// Backlog is retained: recovery runs full sequential steps, never skips tick IDs.
type tickSchedule struct {
	interval        time.Duration
	next            time.Time
	recoveryStarted time.Time
	lastRecoveryLog time.Time
	recoveryTicks   uint64
}

func newTickSchedule(now time.Time, interval time.Duration) tickSchedule {
	return tickSchedule{interval: interval, next: now.Add(interval)}
}

func (s *tickSchedule) delay(now time.Time) time.Duration { return max(0, s.next.Sub(now)) }

func (s *tickSchedule) complete(started, finished time.Time, timings tickTimings, logf func(string, ...any)) {
	duration := finished.Sub(started)
	lateness := max(0, started.Sub(s.next))
	s.next = s.next.Add(s.interval)
	if duration > s.interval {
		logf("tick_overrun tick=%d duration=%s budget=%s start_lateness=%s lock_wait=%s systems=%s sync=%s events=%s slowest_system=%T slowest_system_duration=%s",
			timings.tick, duration, s.interval, lateness, timings.lockWait, timings.systems, timings.sync, timings.events, timings.slowestSystem, timings.slowestSystemDuration)
	}
	if !s.recoveryStarted.IsZero() {
		s.recoveryTicks++
	}
	if !finished.Before(s.next) {
		backlog := finished.Sub(s.next)
		pending := backlog/s.interval + 1
		if s.recoveryStarted.IsZero() {
			s.recoveryStarted, s.lastRecoveryLog = finished, finished
			logf("tick_recovery_started tick=%d overdue=%s pending_ticks=%d start_lateness=%s", timings.tick, backlog, pending, lateness)
		} else if finished.Sub(s.lastRecoveryLog) >= recoveryLogInterval {
			s.lastRecoveryLog = finished
			logf("tick_recovery_progress tick=%d overdue=%s pending_ticks=%d recovery_ticks=%d elapsed=%s", timings.tick, backlog, pending, s.recoveryTicks, finished.Sub(s.recoveryStarted))
		}
	} else if !s.recoveryStarted.IsZero() {
		logf("tick_recovery_finished tick=%d recovery_ticks=%d elapsed=%s", timings.tick, s.recoveryTicks, finished.Sub(s.recoveryStarted))
		s.recoveryStarted, s.lastRecoveryLog = time.Time{}, time.Time{}
		s.recoveryTicks = 0
	}
}

func (g *Game) StartUpdateLoop(interval time.Duration) {
	if interval <= 0 {
		panic("tick interval must be positive")
	}
	g.tickInterval = interval
	log.Printf("tick_loop_started interval=%s ticks_per_second=%.2f", interval, float64(time.Second)/float64(interval))
	go func() {
		schedule := newTickSchedule(time.Now(), interval)
		timer := time.NewTimer(interval)
		defer timer.Stop()
		for {
			// Check shutdown between every recovery step, including sustained overload.
			select {
			case <-g.done:
				return
			default:
			}
			if delay := schedule.delay(time.Now()); delay > 0 {
				timer.Reset(delay)
				select {
				case <-g.done:
					return
				case <-timer.C:
				}
			}
			started := time.Now()
			timings := g.update()
			schedule.complete(started, time.Now(), timings, log.Printf)
		}
	}()
}
