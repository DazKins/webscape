export const CombatLogUpdateEventName = "combatLogUpdate";

export class CombatLogUpdateEvent extends Event {
  constructor() {
    super(CombatLogUpdateEventName);
  }
}
