package config

import "testing"

func TestAuthFailsClosed(t *testing.T) {
	base := AuthConfig{Issuer: "https://id.example.com", ClientID: "game", ClientSecretEnv: "TEST_SECRET", PublicURL: "https://game.example.com", SessionLifetimeSeconds: 3600}
	if err := base.Validate(false); err != nil {
		t.Fatal(err)
	}
	for _, mutate := range []func(*AuthConfig){
		func(c *AuthConfig) { c.Issuer = "" }, func(c *AuthConfig) { c.ClientID = "" }, func(c *AuthConfig) { c.ClientSecretEnv = "" }, func(c *AuthConfig) { c.PublicURL = "http://game.example.com" },
		func(c *AuthConfig) { c.PublicURL = "https://game.example.com/path" }, func(c *AuthConfig) { c.Issuer = "https://user:pass@id.example.com" }, func(c *AuthConfig) { c.Issuer = "https://id.example.com#fragment" },
		func(c *AuthConfig) { c.SessionLifetimeSeconds = 0 }, func(c *AuthConfig) { c.SessionLifetimeSeconds = 86401 },
	} {
		c := base
		mutate(&c)
		if c.Validate(false) == nil {
			t.Fatalf("accepted %+v", c)
		}
	}
	c := base
	c.Issuer = "http://127.0.0.1:9000"
	c.PublicURL = "http://localhost:8080"
	if c.Validate(false) == nil || c.Validate(true) != nil {
		t.Fatal("HTTP development restriction failed")
	}
	c.Issuer = "http://192.168.1.2:9000"
	if c.Validate(true) == nil {
		t.Fatal("non-loopback HTTP accepted")
	}
	if (Config{Auth: AuthConfig{}}).Auth.Validate(true) == nil {
		t.Fatal("missing auth accepted")
	}
}
