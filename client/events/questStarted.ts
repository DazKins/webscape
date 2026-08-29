export const QuestStartedEventName = "questStarted";

export type QuestStartedPayload = {
  questId: string;
  displayName?: string;
  description?: string;
  currentStepId?: string;
  currentStepSummary?: string;
};

export class QuestStartedEvent extends Event {
  constructor(public payload: QuestStartedPayload) {
    super(QuestStartedEventName);
  }
}
