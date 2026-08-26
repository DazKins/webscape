package message

import (
	"encoding/json"
	"testing"
	"webscape/server/game/model"
)

func TestClientGameEventMessagesSerializeDomainData(t *testing.T) {
	attackerId := model.NewEntityId()
	targetId := model.NewEntityId()
	tests := []struct {
		name    string
		message Message
		typeId  string
		assert  func(t *testing.T, data map[string]any)
	}{
		{
			name:    "chat",
			message: NewChatMessage(attackerId, "hello"),
			typeId:  "chatMessage",
			assert: func(t *testing.T, data map[string]any) {
				if data["fromEntityId"] != attackerId.String() || data["message"] != "hello" {
					t.Fatalf("chat data = %#v", data)
				}
			},
		},
		{
			name:    "combat",
			message: NewCombatResolvedMessage(attackerId, targetId, true, 9, true),
			typeId:  "combatResolved",
			assert: func(t *testing.T, data map[string]any) {
				if data["attackerEntityId"] != attackerId.String() || data["targetEntityId"] != targetId.String() ||
					data["didHit"] != true || data["damage"] != float64(9) || data["isCritical"] != true {
					t.Fatalf("combat data = %#v", data)
				}
			},
		},
		{
			name:    "woodcutting",
			message: NewWoodcuttingSwingMessage(attackerId, targetId),
			typeId:  "woodcuttingSwing",
			assert: func(t *testing.T, data map[string]any) {
				if data["playerEntityId"] != attackerId.String() || data["targetEntityId"] != targetId.String() {
					t.Fatalf("woodcutting data = %#v", data)
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var payload struct {
				Metadata struct {
					Type string `json:"type"`
				} `json:"metadata"`
				Data map[string]any `json:"data"`
			}
			if err := json.Unmarshal([]byte(test.message.Marshal()), &payload); err != nil {
				t.Fatal(err)
			}
			if payload.Metadata.Type != test.typeId {
				t.Fatalf("message type = %q, want %q", payload.Metadata.Type, test.typeId)
			}
			test.assert(t, payload.Data)
		})
	}
}
