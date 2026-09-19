package message

import "webscape/server/game/model"

const MessageTypeBankResult MessageType = "bankResult"

type BankResult struct {
	TargetEntityId string `json:"targetEntityId"`
	Success        bool   `json:"success"`
	Text           string `json:"text"`
}

func NewBankResultMessage(target model.EntityId, success bool, text string) Message {
	return newMessage(MessageTypeBankResult, BankResult{TargetEntityId: target.String(), Success: success, Text: text})
}
