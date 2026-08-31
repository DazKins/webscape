import { useEffect, useState } from "react";
import styles from "./combatLog.module.css";
import panelStyles from "./uiPanel.module.css";
import Game from "../../game/game";
import { CombatLogUpdateEvent, CombatLogUpdateEventName } from "../../events/combatlog";

type Props = {
  game: Game;
};

type CombatLogEntry = {
  text: string;
  kind: string;
};

function entryClass(kind: string): string {
  if (kind === "crit" || kind === "woodcutting-good" || kind === "woodcutting-reward" || kind === "fishing-reward") {
    return styles.entryCrit;
  }
  if (kind === "miss" || kind === "woodcutting-miss") {
    return styles.entryMiss;
  }
  return styles.entryHit;
}

export function CombatLogContent(props: Props) {
  const [entries, setEntries] = useState<CombatLogEntry[]>([]);

  const updateCombatLog = () => {
    const myEntity = props.game.getMyEntity();
    if (!myEntity) {
      setEntries([]);
      return;
    }

    const combatLogComponent = myEntity.getComponent("combatlog");
    if (!combatLogComponent || !combatLogComponent.entries) {
      setEntries([]);
      return;
    }

    setEntries(combatLogComponent.entries || []);
  };

  useEffect(() => {
    updateCombatLog();

    const handler = (_event: CombatLogUpdateEvent) => {
      updateCombatLog();
    };

    props.game.addEventListener(CombatLogUpdateEventName, handler as EventListener);
    return () => {
      props.game.removeEventListener(CombatLogUpdateEventName, handler as EventListener);
    };
  }, [props.game]);

  return (
    <div className={panelStyles.panelContent}>
      {entries.length === 0 ? (
        <div className={styles.empty}>No activity yet</div>
      ) : (
        entries.map((entry, index) => (
          <div
            key={index}
            className={`${styles.entry} ${entryClass(entry.kind)}`}
          >
            {entry.text}
          </div>
        ))
      )}
    </div>
  );
}

export default function CombatLog(props: Props) {
  return (
    <div className={`${panelStyles.panel} ${styles.container}`}>
      <div className={panelStyles.panelHeader}>Activity Log</div>
      <CombatLogContent game={props.game} />
    </div>
  );
}
