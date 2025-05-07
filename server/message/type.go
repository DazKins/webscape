package message

import "fmt"

type MessageType int

const (
	MessageTypeEntityUpdate MessageType = iota
	MessageTypeJoin
	MessageTypeJoined
	MessageTypeWorld
	MessageTypeMove
	MessageTypeEntityRemove
)

var messageTypeStrings = []string{"entityUpdate", "join", "joined", "world", "move", "entityRemove"}

func (m MessageType) String() string {
	return messageTypeStrings[m]
}

type MessageTypeDto string

func (m MessageType) ToDto() MessageTypeDto {
	return MessageTypeDto(m.String())
}

func (m MessageTypeDto) FromDto() (MessageType, error) {
	for i, s := range messageTypeStrings {
		if s == string(m) {
			return MessageType(i), nil
		}
	}
	return MessageType(0), fmt.Errorf("unknown message type: %s", m)
}
