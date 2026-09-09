package config

import (
	"encoding/json"
	"net/url"
	"testing"
)

func TestPersistenceDefaultsIndependentOfDevMode(t *testing.T) {
	for _, dev := range []bool{false, true} {
		c, err := LoadFromFile("../../config.json")
		if err != nil {
			t.Fatal(err)
		}
		c.Server.DevMode = dev
		raw, _ := json.Marshal(c)
		loaded, err := load(raw)
		if err != nil || loaded.Persistence.Driver != "none" {
			t.Fatalf("default persistence: %v", err)
		}
		c.Persistence.Driver = "postgres"
		c.Persistence.Postgres.ConnectionStringEnv = "WEBSCAPE_TEST_URL"
		raw, _ = json.Marshal(c)
		loaded, err = load(raw)
		if err != nil || loaded.Persistence.Driver != "postgres" {
			t.Fatalf("explicit postgres: %v", err)
		}
	}
}

func TestPostgresConfigurationAndSecretResolution(t *testing.T) {
	c := defaultPersistence()
	c.Driver = "postgres"
	if err := c.Validate(); err == nil {
		t.Fatal("accepted missing connection")
	}
	c.Postgres = PostgresConfig{Host: "::1", Port: 5432, Database: "game db", User: "game@user", PasswordEnv: "WEBSCAPE_TEST_PASSWORD", SSLMode: "verify-full", SSLRootCert: "/certs/root.pem"}
	if err := c.Validate(); err != nil {
		t.Fatal(err)
	}
	t.Setenv("WEBSCAPE_TEST_PASSWORD", "x'@:/?&$ y")
	dsn, err := c.Postgres.ConnectionString()
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := url.Parse(dsn)
	if err != nil {
		t.Fatal(err)
	}
	password, _ := parsed.User.Password()
	if password != "x'@:/?&$ y" || parsed.User.Username() != "game@user" || parsed.Host != "[::1]:5432" || parsed.Query().Get("sslmode") != "verify-full" {
		t.Fatal("connection fields were not escaped correctly")
	}
	c.Postgres.ConnectionStringEnv = "WEBSCAPE_TEST_URL"
	t.Setenv("WEBSCAPE_TEST_URL", "")
	if _, err := c.Postgres.ConnectionString(); err == nil {
		t.Fatal("empty connection env silently fell back")
	}
	t.Setenv("WEBSCAPE_TEST_URL", "postgres://test/db?sslmode=require")
	if got, err := c.Postgres.ConnectionString(); err != nil || got != "postgres://test/db?sslmode=require" {
		t.Fatal("connection env did not take precedence")
	}
	for _, mutate := range []func(*PersistenceConfig){
		func(c *PersistenceConfig) { c.Driver = "sqlite" },
		func(c *PersistenceConfig) { c.WorldKey = "" },
		func(c *PersistenceConfig) { c.TimeoutSeconds = 0 },
		func(c *PersistenceConfig) { c.Postgres.Port = 0 },
		func(c *PersistenceConfig) { c.Postgres.SSLMode = "prefer" },
		func(c *PersistenceConfig) { c.Postgres.Password = "inline" },
	} {
		bad := c
		mutate(&bad)
		if err := bad.Validate(); err == nil {
			t.Fatal("accepted invalid persistence settings")
		}
	}
}
