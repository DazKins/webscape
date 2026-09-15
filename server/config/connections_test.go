package config

import "testing"

func TestConnectionTimeoutValidation(t *testing.T) {
	defaults := DefaultConnections()
	if err := defaults.Validate(); err != nil {
		t.Fatal(err)
	}
	for _, modify := range []func(*ConnectionConfig){
		func(c *ConnectionConfig) { c.PingIntervalSeconds = 0 },
		func(c *ConnectionConfig) { c.PongTimeoutSeconds = c.PingIntervalSeconds },
		func(c *ConnectionConfig) { c.PongTimeoutSeconds = 3601 },
		func(c *ConnectionConfig) { c.IdleWarningSeconds = 0 },
		func(c *ConnectionConfig) { c.IdleWarningSeconds = c.IdleTimeoutSeconds },
		func(c *ConnectionConfig) { c.IdleTimeoutSeconds = 86401 },
	} {
		c := defaults
		modify(&c)
		if err := c.Validate(); err == nil {
			t.Fatalf("accepted invalid timeouts: %+v", c)
		}
	}
}
