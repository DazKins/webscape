import { useEffect, useState } from "react";
import Game, { type QuestDefinition } from "../../game/game";
import { QuestLogUpdateEvent, QuestLogUpdateEventName } from "../../events/questlog";
import panelStyles from "./uiPanel.module.css";
import styles from "./questPanel.module.css";

type Props = {
  game: Game;
};

type QuestProgress = {
  questId: string;
  stepId: string;
  currentStepIndex: number;
  currentCount: number;
};

export default function QuestPanel(props: Props) {
  const [active, setActive] = useState<QuestProgress[]>([]);
  const [quests, setQuests] = useState<QuestDefinition[]>([]);

  const updateQuestLog = () => {
    setQuests(props.game.getQuestDefinitions());

    const myEntity = props.game.getMyEntity();
    if (!myEntity) {
      setActive([]);
      return;
    }

    const questLog = myEntity.getComponent("questlog");
    setActive(questLog?.active ?? []);
  };

  useEffect(() => {
    updateQuestLog();

    const handler = (_event: QuestLogUpdateEvent) => {
      updateQuestLog();
    };

    props.game.addEventListener(QuestLogUpdateEventName, handler as EventListener);
    return () => {
      props.game.removeEventListener(QuestLogUpdateEventName, handler as EventListener);
    };
  }, [props.game]);

  if (quests.length === 0 && active.length === 0) {
    return null;
  }

  return (
    <div className={`${panelStyles.panel} ${styles.container}`}>
      <div className={panelStyles.panelHeader}>Quests</div>
      <div className={panelStyles.panelContent}>
        {active.length === 0 ? (
          <div className={styles.empty}>No active quests</div>
        ) : (
          active.map((progress) => {
            const quest = quests.find((candidate) => candidate.id === progress.questId);
            const step = quest?.steps[progress.currentStepIndex];
            const requiredCount = step?.requirement.count ?? 1;
            return (
              <div key={progress.questId} className={styles.quest}>
                <div className={styles.questName}>{quest?.displayName || quest?.id || progress.questId}</div>
                <div className={styles.stepText}>{step?.description || progress.stepId}</div>
                <div className={styles.progress}>
                  {Math.min(progress.currentCount, requiredCount)} / {requiredCount}
                </div>
              </div>
            );
          })
        )}
      </div>
    </div>
  );
}
