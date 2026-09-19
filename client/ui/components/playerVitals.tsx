import { useEffect, useId, useState } from "react";
import type Game from "../../game/game";
import { PlayerVitalsUpdateEventName } from "../../events/playerVitals";
import styles from "./playerVitals.module.css";

type Vital = { current: number; maximum: number };

function readVitals(game: Game) {
  const player = game.getMyEntity();
  const health = player?.getComponent("health");
  const mana = player?.getComponent("mana");
  return {
    health: health ? { current: health.currentHealth, maximum: health.maxHealth } as Vital : null,
    mana: mana ? { current: mana.currentMana, maximum: mana.maxMana } as Vital : null,
  };
}

function VitalBar({ label, value, kind }: { label: string; value: Vital; kind: "health" | "mana" }) {
  const tooltipId = useId();
  const maximum = Math.max(0, value.maximum);
  const current = Math.max(0, Math.min(maximum, value.current));
  const fraction = maximum > 0 ? current / maximum : 0;
  return (
    <div className={styles.vital} data-kind={kind} data-low={kind === "health" && fraction <= 0.25}>
      <div className={styles.track} role="meter" tabIndex={0} aria-label={label} aria-describedby={tooltipId} aria-valuemin={0} aria-valuemax={maximum} aria-valuenow={current}>
        <div className={styles.reservoir}>
          <div className={styles.fill} style={{ width: `${fraction * 100}%` }} />
        </div>
        <span className={styles.symbol}>
          <svg viewBox="0 0 32 32" aria-hidden="true" focusable="false">
            {kind === "health" ? (
              <>
                <path d="M16 27 5 16C-3 7 9-2 16 7 23-2 35 7 27 16Z" fill="currentColor" stroke="#591d31" strokeWidth="2" strokeLinejoin="round" />
                <path d="M7 7c3-2 6 0 7 3L7 16C3 12 4 9 7 7Z" fill="#fff" opacity=".3" />
                <path d="m16 25 9-10-9 4Z" fill="#8c2541" opacity=".6" />
              </>
            ) : (
              <>
                <path d="m16 2 11 11-5 12-6 5-6-5-5-12Z" fill="currentColor" stroke="#17376d" strokeWidth="2" strokeLinejoin="round" />
                <path d="m16 3-5 10 5 15 5-15Z" fill="#d2faff" />
                <path d="m6 13 5 0 5 15-6-4Z" fill="#477ddc" />
                <path d="m16 3 10 10h-5Z" fill="#fff" opacity=".65" />
              </>
            )}
          </svg>
        </span>
        <span className={styles.amount}>{current} / {maximum}</span>
      </div>
      <span className={styles.tooltip} id={tooltipId} role="tooltip">{label}</span>
    </div>
  );
}

export default function PlayerVitals({ game }: { game: Game }) {
  const [vitals, setVitals] = useState(() => readVitals(game));
  useEffect(() => {
    const update = () => setVitals(readVitals(game));
    update();
    game.addEventListener(PlayerVitalsUpdateEventName, update);
    return () => game.removeEventListener(PlayerVitalsUpdateEventName, update);
  }, [game]);

  if (!vitals.health && !vitals.mana) return null;
  return (
    <section
      className={styles.vitals}
      aria-label="Your health and mana"
      onPointerEnter={() => game.setPointerOverUi(true)}
      onPointerLeave={() => game.setPointerOverUi(false)}
      onPointerDown={event => event.stopPropagation()}
      onPointerUp={event => {
        event.stopPropagation();
        if (event.pointerType !== "mouse") game.setPointerOverUi(false);
      }}
      onClick={event => event.stopPropagation()}
      onKeyDown={event => event.stopPropagation()}
      onContextMenu={event => { event.preventDefault(); event.stopPropagation(); }}
    >
      {vitals.health && <VitalBar label="Health" value={vitals.health} kind="health" />}
      {vitals.mana && <VitalBar label="Mana" value={vitals.mana} kind="mana" />}
    </section>
  );
}
