export const BankUpdateEventName = "bankUpdate";
export const BankResultEventName = "bankResult";
export type BankResultPayload = { targetEntityId: string; success: boolean; text: string };
export class BankResultEvent extends Event {
 constructor(public payload: BankResultPayload) { super(BankResultEventName); }
}
