package config

import "fmt"

// ConnectionConfig controls transport health and player inactivity separately.
type ConnectionConfig struct {
	PingIntervalSeconds int `json:"pingIntervalSeconds"`
	PongTimeoutSeconds  int `json:"pongTimeoutSeconds"`
	IdleTimeoutSeconds  int `json:"idleTimeoutSeconds"`
	IdleWarningSeconds  int `json:"idleWarningSeconds"`
}

func DefaultConnections() ConnectionConfig {
	return ConnectionConfig{PingIntervalSeconds: 20, PongTimeoutSeconds: 60, IdleTimeoutSeconds: 900, IdleWarningSeconds: 60}
}

func (c ConnectionConfig) Validate() error {
	if c.PingIntervalSeconds < 1 || c.PongTimeoutSeconds <= c.PingIntervalSeconds || c.PongTimeoutSeconds > 3600 {
		return fmt.Errorf("config server.connections requires 0 < pingIntervalSeconds < pongTimeoutSeconds <= 3600")
	}
	if c.IdleWarningSeconds < 1 || c.IdleTimeoutSeconds <= c.IdleWarningSeconds || c.IdleTimeoutSeconds > 86400 {
		return fmt.Errorf("config server.connections requires 0 < idleWarningSeconds < idleTimeoutSeconds <= 86400")
	}
	return nil
}
