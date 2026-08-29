import { useEffect, useState, type SyntheticEvent } from "react";
import Game from "../../game/game";
import {
  QuestStartedEvent,
  QuestStartedEventName,
  type QuestStartedPayload,
} from "../../events/questStarted";
import styles from "./questStartedOverlay.module.css";

type Props = {
  game: Game;
};

export default function QuestStartedOverlay(props: Props) {
  const [notifications, setNotifications] = useState<QuestStartedPayload[]>([]);

  useEffect(() => {
    const handler = (event: QuestStartedEvent) => {
      setNotifications((current) => [...current, event.payload]);
    };
    props.game.addEventListener(QuestStartedEventName, handler as EventListener);
    return () => {
      props.game.removeEventListener(QuestStartedEventName, handler as EventListener);
    };
  }, [props.game]);

  const quest = notifications[0];
  if (!quest) {
    return null;
  }

  const stopOverlayEvent = (event: SyntheticEvent) => {
    event.stopPropagation();
  };
  const close = () => {
    props.game.setPointerOverUi(false);
    setNotifications((current) => current.slice(1));
  };

  return (
    <div className={styles.backdrop}>
      <div
        className={styles.overlay}
        role="dialog"
        aria-modal="true"
        aria-labelledby="quest-started-title"
        onClick={stopOverlayEvent}
        onContextMenu={stopOverlayEvent}
        onPointerDown={(event) => {
          stopOverlayEvent(event);
          props.game.setPointerOverUi(true);
        }}
        onPointerEnter={() => props.game.setPointerOverUi(true)}
        onPointerLeave={() => props.game.setPointerOverUi(false)}
      >
        <div className={styles.emblem} aria-hidden="true">!</div>
        <div className={styles.kicker}>New Quest Started</div>
        <h2 id="quest-started-title">{quest.displayName || quest.questId}</h2>
        {quest.description ? <p className={styles.description}>{quest.description}</p> : null}
        {quest.currentStepSummary ? (
          <div className={styles.objective}>
            <span>Current objective</span>
            <strong>{quest.currentStepSummary}</strong>
          </div>
        ) : null}
        <button type="button" className={styles.closeButton} onClick={close}>
          Begin Quest
        </button>
      </div>
    </div>
  );
}
