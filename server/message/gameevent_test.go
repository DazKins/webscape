package message

import (
	"encoding/json"
	"testing"
	"webscape/server/game/model"
	"webscape/server/math"
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
			message: NewCombatResolvedMessage(attackerId, targetId, true, 9, true, model.AttackMethodMagic),
			typeId:  "combatResolved",
			assert: func(t *testing.T, data map[string]any) {
				if data["attackerEntityId"] != attackerId.String() || data["targetEntityId"] != targetId.String() ||
					data["didHit"] != true || data["damage"] != float64(9) || data["isCritical"] != true || data["attackMethod"] != "magic" {
					t.Fatalf("combat data = %#v", data)
				}
			},
		},
		{
			name: "combat projectile",
			message: NewCombatProjectileLaunchedMessage(
				attackerId, targetId, "magicBolt",
				math.Vec2{X: 1, Y: 2}, math.Vec2{X: 4, Y: 5}, 20, 21,
			),
			typeId: "combatProjectileLaunched",
			assert: func(t *testing.T, data map[string]any) {
				origin := data["origin"].(map[string]any)
				target := data["targetPosition"].(map[string]any)
				if data["attackerEntityId"] != attackerId.String() || data["targetEntityId"] != targetId.String() ||
					data["projectileType"] != "magicBolt" || origin["x"] != float64(1) || origin["y"] != float64(2) ||
					target["x"] != float64(4) || target["y"] != float64(5) || data["launchTick"] != float64(20) || data["impactTick"] != float64(21) {
					t.Fatalf("combat projectile data = %#v", data)
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
