package message

type chatMessageData struct {
	EntityId string `json:"entityId"`
	Message  string `json:"message"`
}

func NewChatMessage(entityId string, message string) Message {
	return newMessage(
		MessageTypeChat,
		chatMessageData{
			EntityId: entityId,
			Message:  message,
		},
	)
}
