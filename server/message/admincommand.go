package message

// Admin responses are private session notifications, never public player speech.
func NewAdminCommandResult(command string, success bool, text string) Message {
	return newMessage(MessageTypeAdminCommandResult, struct {
		Command string `json:"command"`
		Success bool   `json:"success"`
		Message string `json:"message"`
	}{command, success, text})
}
