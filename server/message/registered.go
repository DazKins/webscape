package message

type registeredData struct {
	EntityId string `json:"entityId"`
	Name     string `json:"name"`
}

func NewRegisteredMessage(entityId string, name string) Message {
	return newMessage(
		MessageTypeRegistered,
		registeredData{
			EntityId: entityId,
			Name:     name,
		},
	)
}
