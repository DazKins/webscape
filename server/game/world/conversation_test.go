package world

import (
	"encoding/json"
	"testing"
	"testing/fstest"
)

func TestConversationStartBranchValidation(t *testing.T) {
	for _, test := range []struct {
		name   string
		branch ConversationStartBranch
		valid  bool
	}{
		{"active step", ConversationStartBranch{QuestId: "errand", Status: "active", StepId: "return", NodeId: "end"}, true},
		{"any active step", ConversationStartBranch{QuestId: "errand", Status: "active", NodeId: "end"}, true},
		{"completed", ConversationStartBranch{QuestId: "errand", Status: "completed", NodeId: "end"}, true},
		{"missing quest", ConversationStartBranch{Status: "active", NodeId: "end"}, false},
		{"unknown quest", ConversationStartBranch{QuestId: "missing", Status: "active", NodeId: "end"}, false},
		{"unknown step", ConversationStartBranch{QuestId: "errand", Status: "active", StepId: "missing", NodeId: "end"}, false},
		{"unknown node", ConversationStartBranch{QuestId: "errand", Status: "active", NodeId: "missing"}, false},
		{"unknown status", ConversationStartBranch{QuestId: "errand", Status: "ready", NodeId: "end"}, false},
		{"step with completed status", ConversationStartBranch{QuestId: "errand", Status: "completed", StepId: "return", NodeId: "end"}, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			document := conversationDocument{
				FormatVersion: 1, Id: "greetings",
				Conversations: []Conversation{{
					Id: "guide", StartNodeId: "end", StartBranches: []ConversationStartBranch{test.branch},
					Nodes: []ConversationNode{{Id: "end", Messages: []ConversationMessage{{Text: "Hello."}}, EndConversation: true}},
				}},
			}
			data, err := json.Marshal(document)
			if err != nil {
				t.Fatal(err)
			}
			registry, err := loadConversationRegistry(fstest.MapFS{"greetings.json": {Data: data}}, []string{"greetings.json"})
			if err == nil {
				quests := NewQuestRegistry()
				quests.quests["errand"] = &Quest{Id: "errand", Steps: []QuestStep{{Id: "return"}}}
				err = registry.validateQuestReferences(quests)
			}
			if (err == nil) != test.valid {
				t.Fatalf("validation error = %v, want valid %v", err, test.valid)
			}
		})
	}
}
