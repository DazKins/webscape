package command

import (
	"encoding/json"
)

type CommandType string

const (
	CommandTypeRegister           = "register"
	CommandTypeActivity           = "activity"
	CommandTypeMove               = "move"
	CommandTypeChat               = "chat"
	CommandTypeInteract           = "interact"
	CommandTypeEquip              = "equip"
	CommandTypeUnequip            = "unequip"
	CommandTypeDrop               = "drop"
	CommandTypeBankDeposit        = "bankDeposit"
	CommandTypeBankWithdraw       = "bankWithdraw"
	CommandTypeBankClose          = "bankClose"
	CommandTypeTrade              = "trade"
	CommandTypeTradeClose         = "tradeClose"
	CommandTypeConversationOption = "conversationOption"
)

type Command struct {
	Type CommandType    `json:"type"`
	Data map[string]any `json:"data"`
}

func Unmarshal(data string) (Command, error) {
	var command Command
	err := json.Unmarshal([]byte(data), &command)
	if err != nil {
		return Command{}, err
	}
	return command, nil
}
