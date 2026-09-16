package config

import (
	"fmt"
	"testing"
)

func TestAdminCommandConfiguration(t *testing.T) {
	const id = "416884e7-057e-4aef-abdb-b8ea45c07ea1"
	base := `{"formatVersion":1,"server":{"address":":8080","devMode":%t},"client":{"folder":"client/dist"},"game":{"folder":"game-project"},"auth":{"issuer":"https://id.example.com","clientId":"game","clientSecretEnv":"TEST_SECRET","publicUrl":"https://game.example.com","sessionLifetimeSeconds":3600},"streaming":{"chunkRadius":1}%s}`
	for _, tt := range []struct {
		name                    string
		dev                     bool
		setting                 string
		enabled, allowed, valid bool
	}{
		{"prod default", false, "", false, false, true},
		{"dev default", true, "", true, true, true},
		{"disabled dev", true, `,"adminCommands":{"enabled":false}`, false, false, true},
		{"prod empty list", false, `,"adminCommands":{"enabled":true,"playerIds":[]}`, true, false, true},
		{"prod allowed", false, `,"adminCommands":{"enabled":true,"playerIds":["` + id + `"]}`, true, true, true},
		{"dev restricted", true, `,"adminCommands":{"playerIds":["00000000-0000-0000-0000-000000000001"]}`, true, false, true},
		{"invalid id", true, `,"adminCommands":{"playerIds":["name"]}`, false, false, false},
		{"nil id", true, `,"adminCommands":{"playerIds":["00000000-0000-0000-0000-000000000000"]}`, false, false, false},
		{"invalid flag", true, `,"adminCommands":{"enabled":"yes"}`, false, false, false},
		{"unknown setting", true, `,"adminCommands":{"enable":true}`, false, false, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := load([]byte(fmt.Sprintf(base, tt.dev, tt.setting)))
			if (err == nil) != tt.valid {
				t.Fatalf("error=%v", err)
			}
			if tt.valid && (cfg.AdminCommands.IsEnabled(tt.dev) != tt.enabled || cfg.AdminCommands.Allows(id, tt.dev) != tt.allowed) {
				t.Fatalf("unexpected policy: %+v", cfg.AdminCommands)
			}
		})
	}
}
