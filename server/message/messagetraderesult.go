package message

import "webscape/server/game/model"

const MessageTypeTradeResult MessageType = "tradeResult"

type TradeResult struct {
	TargetEntityId string `json:"targetEntityId"`
	Success        bool   `json:"success"`
	Text           string `json:"text"`
}

func NewTradeResultMessage(target model.EntityId, success bool, text string) Message {
	return newMessage(MessageTypeTradeResult, TradeResult{TargetEntityId: target.String(), Success: success, Text: text})
}
