package message

type MessageType string

const (
	MessageTypeGameUpdate     = "gameUpdate"
	MessageTypeJoined         = "joined"
	MessageTypeJoinFailed     = "joinFailed"
	MessageTypeWorld          = "world"
	MessageTypeEntityRemove   = "entityRemove"
	MessageTypeConversation   = "conversation"
	MessageTypeQuestCompleted = "questCompleted"
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
