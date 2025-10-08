package message

type MessageType string

const (
	MessageTypeEntityUpdate = "entityUpdate"
	MessageTypeJoin         = "join"
	MessageTypeJoined       = "joined"
	MessageTypeJoinFailed   = "joinFailed"
	MessageTypeWorld        = "world"
	MessageTypeMove         = "move"
	MessageTypeEntityRemove = "entityRemove"
)

func (m MessageType) String() string {
	return string(m)
}

type MessageTypeDto string

func (m MessageType) ToDto() MessageTypeDto {
	return MessageTypeDto(m.String())
}

func (m MessageTypeDto) FromDto() (MessageType, error) {
	return MessageType(m), nil
}
