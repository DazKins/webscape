import { useEffect, useId, useRef, useState } from "react";
import type Game from "../../game/game";
import styles from "./dayCycleIndicator.module.css";

export default function DayCycleIndicator({ game }: { game: Game }) {
  const [display, setDisplay] = useState<{ isNight: boolean; seconds: number }>();
  const ringRef = useRef<SVGCircleElement>(null);
  const tooltipId = useId();

  useEffect(() => {
    let previous = "";
    const update = () => {
      const state = game.dayCycle.read();
      ringRef.current?.setAttribute("stroke-dasharray", `${(1 - (state?.phaseProgress ?? 0)) * 100} 100`);
      const key = state ? `${state.isNight}:${state.secondsRemaining}` : "waiting";
      if (key === previous) return;
      previous = key;
      setDisplay(state ? { isNight: state.isNight, seconds: state.secondsRemaining } : undefined);
    };
    update();
    game.addEventListener("frameRendered", update);
    return () => game.removeEventListener("frameRendered", update);
  }, [game]);

  const nextPhase = display?.isNight ? "day" : "night";
  const minutes = Math.floor((display?.seconds ?? 0) / 60);
  const seconds = (display?.seconds ?? 0) % 60;
  const tooltip = display
    ? `${minutes} ${minutes === 1 ? "minute" : "minutes"} ${seconds} ${seconds === 1 ? "second" : "seconds"} until ${nextPhase}`
    : "Waiting for world time";

  return (
    <div className={styles.indicator} data-night={display?.isNight ?? false}>
      <button
        type="button"
        className={styles.dial}
        aria-label={display ? `${display.isNight ? "Night" : "Day"}. ${tooltip}` : tooltip}
        aria-describedby={tooltipId}
        onPointerDown={event => event.stopPropagation()}
        onPointerUp={event => event.stopPropagation()}
        onKeyDown={event => event.stopPropagation()}
        onContextMenu={event => { event.preventDefault(); event.stopPropagation(); }}
      >
        <svg viewBox="0 0 72 72" className={styles.face} aria-hidden="true">
          <circle cx="36" cy="36" r="33" className={styles.rim} />
          <circle cx="36" cy="36" r="28" className={styles.track} />
          <circle ref={ringRef} cx="36" cy="36" r="28" pathLength="100" className={styles.progress} transform="rotate(-90 36 36)" />
          {display?.isNight ? (
            <g className={styles.symbol}>
              <path d="M40 21a15 15 0 1 0 12 22A16 16 0 0 1 40 21Z" fill="currentColor" stroke="none" />
              <path d="M48 21v6m-3-3h6m2 7v4m-2-2h4" />
            </g>
          ) : (
            <g className={styles.symbol}>
              <circle cx="36" cy="36" r="8" fill="currentColor" stroke="none" />
              <path d="M36 19v4m0 26v4M19 36h4m26 0h4M24 24l3 3m18 18 3 3M24 48l3-3m18-18 3-3" />
            </g>
          )}
        </svg>
      </button>
      <span role="tooltip" id={tooltipId} className={styles.tooltip}>{tooltip}</span>
    </div>
  );
}
