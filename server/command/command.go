package command

import (
	"encoding/json"
)

type CommandType string

const (
	CommandTypeJoin     = "join"
	CommandTypeMove     = "move"
	CommandTypeChat     = "chat"
	CommandTypeInteract = "interact"
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
