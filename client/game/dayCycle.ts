export type DayCycleSettings = { dayTicks: number; cycleTicks: number };

export type DayCycleState = {
  isNight: boolean;
  secondsRemaining: number;
  phaseProgress: number;
  cycleProgress: number;
};

// Confirmed ticks control phase and countdown. Only visual progress interpolates,
// and it stops before the next tick if updates stop arriving.
export default class DayCycleClock {
  private settings?: DayCycleSettings;
  private tick?: number;
  private receivedAtMs = 0;
  private tickIntervalMs = 500;

  configure(settings: DayCycleSettings | undefined, tickIntervalMs: number) {
    this.reset();
    if (settings && Number.isSafeInteger(settings.dayTicks) &&
        Number.isSafeInteger(settings.cycleTicks) && settings.dayTicks > 0 &&
        settings.cycleTicks > settings.dayTicks &&
        Number.isFinite(tickIntervalMs) && tickIntervalMs > 0) {
      this.settings = { ...settings };
      this.tickIntervalMs = tickIntervalMs;
    }
  }

  receive(tick: number, nowMs = performance.now()) {
    if (!Number.isSafeInteger(tick) || tick < 0 ||
        (this.tick !== undefined && tick <= this.tick)) return;
    this.tick = tick;
    this.receivedAtMs = nowMs;
  }

  reset() {
    this.tick = undefined;
    this.settings = undefined;
  }

  read(nowMs = performance.now()): DayCycleState | undefined {
    if (!this.settings || this.tick === undefined) return;
    const { dayTicks, cycleTicks } = this.settings;
    const cycleTick = this.tick % cycleTicks;
    const isNight = cycleTick >= dayTicks;
    const phaseStart = isNight ? dayTicks : 0;
    const phaseEnd = isNight ? cycleTicks : dayTicks;
    const fraction = Math.min(0.999, Math.max(0, (nowMs - this.receivedAtMs) / this.tickIntervalMs));
    return {
      isNight,
      secondsRemaining: Math.ceil((phaseEnd - cycleTick) * this.tickIntervalMs / 1000),
      phaseProgress: (cycleTick + fraction - phaseStart) / (phaseEnd - phaseStart),
      cycleProgress: (cycleTick + fraction) / cycleTicks,
    };
  }
}
