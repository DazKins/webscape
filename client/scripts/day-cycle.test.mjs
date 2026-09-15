import assert from "node:assert/strict";
import test from "node:test";
import DayCycleClock from "../game/dayCycle.ts";

const settings = { dayTicks: 1200, cycleTicks: 2400 };
function clockAt(tick) {
  const clock = new DayCycleClock();
  clock.configure(settings, 500);
  clock.receive(tick, 1000);
  return clock;
}

test("confirmed ticks control phase and countdown at both boundaries", () => {
  for (const [tick, isNight, seconds] of [[0, false, 600], [1199, false, 1],
    [1200, true, 600], [2399, true, 1], [2400, false, 600]]) {
    const state = clockAt(tick).read(1000);
    assert.equal(state.isNight, isNight);
    assert.equal(state.secondsRemaining, seconds);
  }
});

test("visual interpolation stops within the last confirmed tick during a stall", () => {
  for (const tick of [100, 1199, 2399]) {
    const clock = clockAt(tick);
    const initial = clock.read(1000);
    const halfway = clock.read(1250);
    const stalled = clock.read(1e9);
    assert.ok(halfway.cycleProgress > initial.cycleProgress);
    assert.ok(stalled.cycleProgress < (tick + 1) / 2400);
    assert.ok(stalled.phaseProgress < 1);
    assert.equal(stalled.isNight, initial.isNight);
    assert.equal(stalled.secondsRemaining, initial.secondsRemaining);
    assert.deepEqual(stalled, clock.read(2e9));
  }
});

test("fresh ticks correct skipped updates, while duplicates do not rewind visuals", () => {
  const clock = clockAt(1199);
  const before = clock.read(1250);
  clock.receive(1199, 1250);
  clock.receive(1198, 1250);
  assert.deepEqual(clock.read(1250), before);
  clock.receive(1400, 2000);
  assert.equal(clock.read(2000).isNight, true);
  assert.equal(clock.read(2000).secondsRemaining, 500);
});

test("reconnect waits for metadata and accepts a restarted server's lower tick", () => {
  const clock = clockAt(2000);
  clock.reset();
  assert.equal(clock.read(1000), undefined);
  clock.configure(settings, 500);
  assert.equal(clock.read(1000), undefined);
  clock.receive(5, 1000);
  assert.equal(clock.read(1000).isNight, false);
  assert.equal(clock.read(1000).secondsRemaining, 598);
});

test("advertised tick interval controls countdown units, never phase boundaries", () => {
  const clock = new DayCycleClock();
  clock.configure(settings, 250);
  clock.receive(1200, 1000);
  assert.equal(clock.read(1000).isNight, true);
  assert.equal(clock.read(1000).secondsRemaining, 300);
});

test("invalid metadata and ticks cannot create a clock", () => {
  const clock = new DayCycleClock();
  for (const invalid of [undefined, { dayTicks: 0, cycleTicks: 2400 },
    { dayTicks: 2400, cycleTicks: 2400 }, { dayTicks: 1.5, cycleTicks: 2400 }]) {
    clock.configure(invalid, 500);
    clock.receive(0, 1000);
    assert.equal(clock.read(1000), undefined);
  }
  clock.configure(settings, 500);
  for (const tick of [-1, 0.5, NaN, Infinity]) clock.receive(tick, 1000);
  assert.equal(clock.read(1000), undefined);
});
