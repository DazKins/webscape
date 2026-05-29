package world

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
)

type ConversationRegistry struct {
	conversations map[string]*Conversation
}

type conversationDocument struct {
	FormatVersion int            `json:"formatVersion"`
	Id            string         `json:"id"`
	DisplayName   string         `json:"displayName"`
	Conversations []Conversation `json:"conversations"`
}

type Conversation struct {
	Id          string             `json:"id"`
	StartNodeId string             `json:"startNodeId"`
	Nodes       []ConversationNode `json:"nodes"`
}

type ConversationNode struct {
	Id              string                `json:"id"`
	Messages        []ConversationMessage `json:"messages"`
	Options         []ConversationOption  `json:"options"`
	EndConversation bool                  `json:"endConversation"`
	Tags            []string              `json:"tags"`
}

type ConversationMessage struct {
	Speaker  string `json:"speaker"`
	Text     string `json:"text"`
	Portrait string `json:"portrait"`
}

type ConversationOption struct {
	Id         string `json:"id"`
	Text       string `json:"text"`
	NextNodeId string `json:"nextNodeId"`
}

func NewConversationRegistry() *ConversationRegistry {
	return &ConversationRegistry{
		conversations: map[string]*Conversation{},
	}
}

func loadConversationRegistry(gameFS fs.FS, paths []string) (*ConversationRegistry, error) {
	registry := NewConversationRegistry()
	for _, path := range paths {
		data, err := fs.ReadFile(gameFS, path)
		if err != nil {
			return nil, fmt.Errorf("read conversation %q: %w", path, err)
		}

		var document conversationDocument
		if err := json.Unmarshal(data, &document); err != nil {
			return nil, fmt.Errorf("parse conversation %q: %w", path, err)
		}
		if err := validateConversationDocument(document); err != nil {
			return nil, fmt.Errorf("validate conversation %q: %w", path, err)
		}

		for i := range document.Conversations {
			conversation := &document.Conversations[i]
			if _, exists := registry.conversations[conversation.Id]; exists {
				return nil, fmt.Errorf("duplicate conversation id %q", conversation.Id)
			}
			registry.conversations[conversation.Id] = conversation
		}
	}

	return registry, nil
}

func (r *ConversationRegistry) Get(id string) (*Conversation, bool) {
	if r == nil {
		return nil, false
	}

	conversation, ok := r.conversations[id]
	return conversation, ok
}

func (r *ConversationRegistry) Len() int {
	if r == nil {
		return 0
	}
	return len(r.conversations)
}

func (c *Conversation) GetNode(id string) (*ConversationNode, bool) {
	if c == nil {
		return nil, false
	}

	for i := range c.Nodes {
		if c.Nodes[i].Id == id {
			return &c.Nodes[i], true
		}
	}
	return nil, false
}

func validateConversationDocument(document conversationDocument) error {
	if document.FormatVersion != 1 {
		return fmt.Errorf("unsupported conversation format version %d", document.FormatVersion)
	}
	if document.Id == "" {
		return errors.New("conversation document id is required")
	}
	if len(document.Conversations) == 0 {
		return errors.New("conversation document must include at least one conversation")
	}

	seenConversationIds := map[string]bool{}
	for _, conversation := range document.Conversations {
		if conversation.Id == "" {
			return errors.New("conversation id is required")
		}
		if seenConversationIds[conversation.Id] {
			return fmt.Errorf("duplicate conversation id %q", conversation.Id)
		}
		seenConversationIds[conversation.Id] = true
		if len(conversation.Nodes) == 0 {
			return fmt.Errorf("conversation %q must include at least one node", conversation.Id)
		}

		nodeIds := map[string]bool{}
		for _, node := range conversation.Nodes {
			if node.Id == "" {
				return fmt.Errorf("conversation %q has a node without an id", conversation.Id)
			}
			if nodeIds[node.Id] {
				return fmt.Errorf("conversation %q has duplicate node id %q", conversation.Id, node.Id)
			}
			nodeIds[node.Id] = true
			if len(node.Messages) == 0 {
				return fmt.Errorf("conversation %q node %q must include at least one message", conversation.Id, node.Id)
			}
			for _, message := range node.Messages {
				if message.Text == "" {
					return fmt.Errorf("conversation %q node %q has an empty message", conversation.Id, node.Id)
				}
			}
			if node.EndConversation && len(node.Options) > 0 {
				return fmt.Errorf("conversation %q node %q cannot end and have options", conversation.Id, node.Id)
			}
			if !node.EndConversation && len(node.Options) == 0 {
				return fmt.Errorf("conversation %q node %q must end or have options", conversation.Id, node.Id)
			}
		}

		if !nodeIds[conversation.StartNodeId] {
			return fmt.Errorf("conversation %q start node %q does not exist", conversation.Id, conversation.StartNodeId)
		}
		for _, node := range conversation.Nodes {
			for _, option := range node.Options {
				if option.Id == "" {
					return fmt.Errorf("conversation %q node %q has an option without an id", conversation.Id, node.Id)
				}
				if option.Text == "" {
					return fmt.Errorf("conversation %q node %q option %q has empty text", conversation.Id, node.Id, option.Id)
				}
				if !nodeIds[option.NextNodeId] {
					return fmt.Errorf("conversation %q option %q targets missing node %q", conversation.Id, option.Id, option.NextNodeId)
				}
			}
		}
	}

	return nil
}
