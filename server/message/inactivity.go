package message

// Inactivity is connection lifecycle information, not gameplay replication.
func NewInactivityMessage(remainingMs, timeoutMs int64, warning bool) Message {
	return newMessage(MessageTypeInactivity, struct {
		RemainingMs int64 `json:"remainingMs"`
		TimeoutMs   int64 `json:"timeoutMs"`
		Warning     bool  `json:"warning"`
	}{remainingMs, timeoutMs, warning})
}

const (
	CloseSessionEnded = 4001
	CloseInactive     = 4002
	CloseReplaced     = 4003
)
