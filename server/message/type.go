package message

type MessageType string

const (
	MessageTypeGameUpdate         = "gameUpdate"
	MessageTypeRegistered         = "registered"
	MessageTypeRegistrationFailed = "registrationFailed"
	MessageTypeWorld              = "world"
	MessageTypeChunkUpdate        = "chunkUpdate"
	MessageTypeConversation       = "conversation"
	MessageTypeQuestStarted       = "questStarted"
	MessageTypeQuestCompleted     = "questCompleted"
	MessageTypeChatMessage        = "chatMessage"
	MessageTypeCombatResolved     = "combatResolved"
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
