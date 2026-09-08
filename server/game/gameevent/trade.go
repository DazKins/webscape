package gameevent

const EventIdTradeResolved = "trade:resolved"

type TradeResolvedPayload struct {
	Success bool
	Text    string
}
