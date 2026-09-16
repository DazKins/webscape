package config

import (
	"fmt"

	"github.com/google/uuid"
)

// AdminCommandsConfig defaults to enabled only in development. An empty allowlist
// permits all development players, but never grants production access.
type AdminCommandsConfig struct {
	Enabled   *bool    `json:"enabled,omitempty"`
	PlayerIDs []string `json:"playerIds"`
}

func (c AdminCommandsConfig) IsEnabled(devMode bool) bool {
	if c.Enabled != nil {
		return *c.Enabled
	}
	return devMode
}

func (c AdminCommandsConfig) Allows(playerID string, devMode bool) bool {
	if !c.IsEnabled(devMode) {
		return false
	}
	if len(c.PlayerIDs) == 0 {
		return devMode
	}
	for _, allowed := range c.PlayerIDs {
		id, err := uuid.Parse(allowed)
		if err == nil && id.String() == playerID {
			return true
		}
	}
	return false
}

func (c AdminCommandsConfig) Validate() error {
	for _, value := range c.PlayerIDs {
		id, err := uuid.Parse(value)
		if err != nil || id == uuid.Nil {
			return fmt.Errorf("config adminCommands.playerIds contains invalid player UUID %q", value)
		}
	}
	return nil
}
