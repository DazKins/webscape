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

type CompletedQuest = {
  questId: string;
};

export function QuestPanelContent(props: Props) {
  const [active, setActive] = useState<QuestProgress[]>([]);
  const [completed, setCompleted] = useState<CompletedQuest[]>([]);
  const [quests, setQuests] = useState<QuestDefinition[]>([]);

  const updateQuestLog = () => {
    setQuests(props.game.getQuestDefinitions());

    const myEntity = props.game.getMyEntity();
    if (!myEntity) {
      setActive([]);
      setCompleted([]);
      return;
    }

    const questLog = myEntity.getComponent("questlog");
    setActive(questLog?.active ?? []);
    setCompleted(questLog?.completed ?? []);
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

  return (
    <div className={panelStyles.panelContent}>
      <section className={styles.section}>
        <h3 className={styles.sectionTitle}>Active</h3>
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
      </section>
      <section className={styles.section}>
        <h3 className={styles.sectionTitle}>Completed</h3>
        {completed.length === 0 ? (
          <div className={styles.empty}>No completed quests</div>
        ) : (
          completed.map((record) => {
            const quest = quests.find((candidate) => candidate.id === record.questId);
            return (
              <div key={record.questId} className={styles.quest}>
                <div className={styles.questName}>{quest?.displayName || quest?.id || record.questId}</div>
                {quest?.description ? <div className={styles.stepText}>{quest.description}</div> : null}
                <div className={styles.rewardList}>
                  {quest?.rewards.items.map((reward) => (
                    <div key={`${record.questId}:${reward.name}:${reward.type}`} className={styles.reward}>
                      {reward.name}
                      {reward.count > 1 ? ` x${reward.count}` : ""}
                    </div>
                  ))}
                </div>
              </div>
            );
          })
        )}
      </section>
    </div>
  );
}

export default function QuestPanel(props: Props) {
  return (
    <div className={`${panelStyles.panel} ${styles.container}`}>
      <div className={panelStyles.panelHeader}>Quests</div>
      <QuestPanelContent game={props.game} />
    </div>
  );
}
