// Package gametime derives the global day/night cycle exclusively from executed
// server ticks. At the standard 500 ms tick interval, each phase lasts 10 minutes.
package gametime

const DayTicks uint64 = 1200
const NightTicks uint64 = 1200
const CycleTicks = DayTicks + NightTicks

// State is a consistent view of game time for a single authoritative tick.
type State struct {
	Tick                 uint64
	CycleTick            uint64
	IsNight              bool
	TicksUntilTransition uint64
}

func AtTick(tick uint64) State {
	cycleTick := tick % CycleTicks
	nextTransition := DayTicks
	if cycleTick >= DayTicks {
		nextTransition = CycleTicks
	}
	return State{
		Tick: tick, CycleTick: cycleTick, IsNight: cycleTick >= DayTicks,
		TicksUntilTransition: nextTransition - cycleTick,
	}
}

func (s State) CycleProgress() float64 {
	return float64(s.CycleTick) / float64(CycleTicks)
}
