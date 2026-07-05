import { useEffect, useState, type SyntheticEvent } from "react";
import Game from "../../game/game";
import {
  QuestCompletedEvent,
  QuestCompletedEventName,
  type QuestCompletedPayload,
  type QuestRewardDelivery,
} from "../../events/questCompleted";
import styles from "./questCompletedOverlay.module.css";

type Props = {
  game: Game;
};

function rewardText(reward: QuestRewardDelivery): string {
  const count = reward.count > 1 ? ` x${reward.count}` : "";
  return `${reward.name}${count}`;
}

function deliveryText(delivery: QuestRewardDelivery["delivery"]): string {
  return delivery === "inventory" ? "Added to backpack" : "Dropped nearby";
}

export default function QuestCompletedOverlay(props: Props) {
  const [completion, setCompletion] = useState<QuestCompletedPayload | null>(null);

  useEffect(() => {
    const handler = (event: QuestCompletedEvent) => {
      setCompletion(event.payload);
    };
    props.game.addEventListener(QuestCompletedEventName, handler as EventListener);
    return () => {
      props.game.removeEventListener(QuestCompletedEventName, handler as EventListener);
    };
  }, [props.game]);

  if (!completion) {
    return null;
  }

  const stopOverlayEvent = (event: SyntheticEvent) => {
    event.stopPropagation();
  };
  const close = () => {
    props.game.setPointerOverUi(false);
    setCompletion(null);
  };

  return (
    <div className={styles.backdrop}>
      <div
        className={styles.overlay}
        role="dialog"
        aria-modal="true"
        aria-labelledby="quest-completed-title"
        onClick={stopOverlayEvent}
        onContextMenu={stopOverlayEvent}
        onPointerDown={(event) => {
          stopOverlayEvent(event);
          props.game.setPointerOverUi(true);
        }}
        onPointerEnter={() => props.game.setPointerOverUi(true)}
        onPointerLeave={() => props.game.setPointerOverUi(false)}
      >
        <div className={styles.kicker}>Quest Completed</div>
        <h2 id="quest-completed-title">{completion.displayName || completion.questId}</h2>
        {completion.description ? <p className={styles.description}>{completion.description}</p> : null}
        {completion.completedStepSummary ? (
          <p className={styles.step}>{completion.completedStepSummary}</p>
        ) : null}
        <div className={styles.rewards}>
          {completion.rewards.map((reward, index) => (
            <div key={`${reward.name}:${reward.delivery}:${index}`} className={styles.rewardRow}>
              <span>{rewardText(reward)}</span>
              <span>{deliveryText(reward.delivery)}</span>
            </div>
          ))}
        </div>
        <button type="button" className={styles.closeButton} onClick={close}>
          Close
        </button>
      </div>
    </div>
  );
}
