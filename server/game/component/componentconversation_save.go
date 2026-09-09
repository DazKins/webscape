package component

import (
	"fmt"
)

type savedConversation struct {
	ConversationId string `json:"conversationId"`
}

func (c *CConversation) Save() (SavedComponent, error) {
	return marshalSaved(savedConversation{ConversationId: c.conversationId})
}
func init() { registerComponentRestore(ComponentIdConversation, 1, restoreConversation) }
func restoreConversation(saved SavedComponent) (Component, error) {
	var s savedConversation
	if err := decodeSaved(saved.Data, &s); err != nil {
		return nil, err
	}
	return &CConversation{conversationId: s.ConversationId}, nil
}

func (c *CConversation) ValidateSaved(ctx SaveContext) error {
	if !ctx.HasConversation(c.conversationId) {
		return fmt.Errorf("unknown saved conversation")
	}
	return nil
}
