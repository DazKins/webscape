package message

type joinedData struct {
	EntityId string `json:"entityId"`
}

func NewJoinedMessage(entityId string) Message {
	return newMessage(
		MessageTypeJoined,
		joinedData{
			EntityId: entityId,
		},
	)
}
