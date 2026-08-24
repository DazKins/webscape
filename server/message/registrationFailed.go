package message

type registrationFailedData struct {
	Reason string `json:"reason"`
}

func NewRegistrationFailedMessage(reason string) Message {
	return newMessage(MessageTypeRegistrationFailed, registrationFailedData{Reason: reason})
}
