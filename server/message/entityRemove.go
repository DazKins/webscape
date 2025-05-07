package message

type entityRemoveData struct {
	EntityId string `json:"entityId"`
}

func NewEntityRemoveMessage(entityId string) Message {
	return newMessage(
		MessageTypeEntityRemove,
		entityRemoveData{
			EntityId: entityId,
		},
	)
}
