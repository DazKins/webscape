export const ShopUpdateEventName = "shopUpdate";
export const TradeResultEventName = "tradeResult";
export type TradeResultPayload = { targetEntityId: string; success: boolean; text: string };
export class TradeResultEvent extends Event {
 constructor(public payload: TradeResultPayload) { super(TradeResultEventName); }
}
