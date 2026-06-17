export const QuestCompletedEventName = "questCompleted";

export type QuestRewardDelivery = {
  name: string;
  type: string;
  count: number;
  delivery: "inventory" | "dropped";
};

export type QuestCompletedPayload = {
  questId: string;
  displayName?: string;
  description?: string;
  completedStepId?: string;
  completedStepSummary?: string;
  rewards: QuestRewardDelivery[];
};

export class QuestCompletedEvent extends Event {
  constructor(public payload: QuestCompletedPayload) {
    super(QuestCompletedEventName);
  }
}
