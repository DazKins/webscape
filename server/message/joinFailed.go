package message

type joinFailedData struct {
	Reason string `json:"reason"`
}

func NewJoinFailedMessage(reason string) Message {
	return newMessage(MessageTypeJoinFailed, joinFailedData{Reason: reason})
}
