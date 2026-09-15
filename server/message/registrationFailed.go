package message

type registrationFailedData struct {
	Code   string `json:"code,omitempty"`
	Reason string `json:"reason"`
}

func NewRegistrationFailedMessage(reason string) Message {
	code := ""
	if reason == "this player is already active" {
		code = "playerAlreadyActive"
	}
	return newMessage(MessageTypeRegistrationFailed, registrationFailedData{Reason: reason, Code: code})
}
