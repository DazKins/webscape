export const QuestLogUpdateEventName = "questlogupdate";

export class QuestLogUpdateEvent extends Event {
  constructor() {
    super(QuestLogUpdateEventName);
  }
}
