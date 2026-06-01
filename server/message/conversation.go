package message

import (
	"webscape/server/game/model"
	"webscape/server/game/world"
)

type conversationMessageData struct {
	Text     string `json:"text"`
	Portrait string `json:"portrait,omitempty"`
}

type conversationOptionData struct {
	Id   string `json:"id"`
	Text string `json:"text"`
}

type conversationData struct {
	ConversationId  string                    `json:"conversationId"`
	TargetEntityId  string                    `json:"targetEntityId"`
	NodeId          string                    `json:"nodeId"`
	Messages        []conversationMessageData `json:"messages"`
	Options         []conversationOptionData  `json:"options,omitempty"`
	EndConversation bool                      `json:"endConversation"`
}

func NewConversationMessage(conversationId string, targetEntityId model.EntityId, node world.ConversationNode) Message {
	messages := make([]conversationMessageData, len(node.Messages))
	for i, message := range node.Messages {
		messages[i] = conversationMessageData{
			Text:     message.Text,
			Portrait: message.Portrait,
		}
	}

	options := make([]conversationOptionData, len(node.Options))
	for i, option := range node.Options {
		options[i] = conversationOptionData{
			Id:   option.Id,
			Text: option.Text,
		}
	}

	return newMessage(
		MessageTypeConversation,
		conversationData{
			ConversationId:  conversationId,
			TargetEntityId:  targetEntityId.String(),
			NodeId:          node.Id,
			Messages:        messages,
			Options:         options,
			EndConversation: node.EndConversation,
		},
	)
}
