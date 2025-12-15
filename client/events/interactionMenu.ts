export const InteractionMenuOpenEventName = "interactionMenuOpen";

export class InteractionMenuOpenEvent extends Event {
  constructor(
    public entityId: string,
    public name: string,
    public interactionOptions: string[],
    public positionX: number,
    public positionY: number
  ) {
    super(InteractionMenuOpenEventName);
  }
}
