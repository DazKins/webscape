package gametime

import "testing"

func TestPhaseBoundariesAndRepeatedCycles(t *testing.T) {
	for _, test := range []struct {
		tick, cycleTick, remaining uint64
		night                      bool
	}{
		{0, 0, 1200, false},
		{1199, 1199, 1, false},
		{1200, 1200, 1200, true},
		{2399, 2399, 1, true},
		{2400, 0, 1200, false},
		{3600, 1200, 1200, true},
		{2400001, 1, 1199, false},
	} {
		state := AtTick(test.tick)
		if state.Tick != test.tick || state.CycleTick != test.cycleTick ||
			state.IsNight != test.night || state.TicksUntilTransition != test.remaining {
			t.Errorf("tick %d: %+v", test.tick, state)
		}
		if state.CycleProgress() != float64(test.cycleTick)/2400 {
			t.Errorf("tick %d: cycle progress %f", test.tick, state.CycleProgress())
		}
	}
}
