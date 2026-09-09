package config

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
)

type PersistenceConfig struct {
	Driver         string         `json:"driver"`
	WorldKey       string         `json:"worldKey"`
	TimeoutSeconds int            `json:"timeoutSeconds"`
	Postgres       PostgresConfig `json:"postgres"`
}

type PostgresConfig struct {
	// ConnectionStringEnv takes precedence over individual fields. The value is
	// resolved at startup, so deployment secrets need not be stored in config.json.
	ConnectionStringEnv string `json:"connectionStringEnv"`
	Host                string `json:"host"`
	Port                int    `json:"port"`
	Database            string `json:"database"`
	User                string `json:"user"`
	Password            string `json:"password"`
	PasswordEnv         string `json:"passwordEnv"`
	SSLMode             string `json:"sslMode"`
	SSLRootCert         string `json:"sslRootCert"`
	SSLCert             string `json:"sslCert"`
	SSLKey              string `json:"sslKey"`
}

func defaultPersistence() PersistenceConfig {
	return PersistenceConfig{Driver: "none", WorldKey: "default", TimeoutSeconds: 10, Postgres: PostgresConfig{Port: 5432, SSLMode: "verify-full"}}
}

func (c PersistenceConfig) Validate() error {
	if c.Driver != "none" && c.Driver != "postgres" {
		return errors.New("config persistence.driver must be none or postgres")
	}
	if c.WorldKey == "" {
		return errors.New("config persistence.worldKey must not be empty")
	}
	if c.TimeoutSeconds < 1 || c.TimeoutSeconds > 300 {
		return errors.New("config persistence.timeoutSeconds must be between 1 and 300")
	}
	p := c.Postgres
	if p.Port < 1 || p.Port > 65535 {
		return errors.New("config persistence.postgres.port must be between 1 and 65535")
	}
	switch p.SSLMode {
	case "disable", "require", "verify-ca", "verify-full":
	default:
		return errors.New("config persistence.postgres.sslMode must be disable, require, verify-ca or verify-full")
	}
	if c.Driver == "postgres" && p.ConnectionStringEnv == "" && (p.Host == "" || p.Database == "" || p.User == "") {
		return errors.New("configure persistence.postgres.connectionStringEnv or host, database and user")
	}
	if p.Password != "" && p.PasswordEnv != "" {
		return errors.New("configure only one of persistence.postgres.password and passwordEnv")
	}
	return nil
}

func (p PostgresConfig) ConnectionString() (string, error) {
	if p.ConnectionStringEnv != "" {
		value, ok := os.LookupEnv(p.ConnectionStringEnv)
		if !ok || value == "" {
			return "", fmt.Errorf("PostgreSQL connection environment variable %s is empty", p.ConnectionStringEnv)
		}
		return value, nil
	}
	password := p.Password
	if p.PasswordEnv != "" {
		value, ok := os.LookupEnv(p.PasswordEnv)
		if !ok || value == "" {
			return "", fmt.Errorf("PostgreSQL password environment variable %s is empty", p.PasswordEnv)
		}
		password = value
	}
	u := url.URL{Scheme: "postgres", Host: net.JoinHostPort(p.Host, strconv.Itoa(p.Port)), Path: "/" + p.Database, User: url.UserPassword(p.User, password)}
	q := url.Values{"sslmode": {p.SSLMode}}
	if p.SSLRootCert != "" {
		q.Set("sslrootcert", p.SSLRootCert)
	}
	if p.SSLCert != "" {
		q.Set("sslcert", p.SSLCert)
	}
	if p.SSLKey != "" {
		q.Set("sslkey", p.SSLKey)
	}
	u.RawQuery = q.Encode()
	return u.String(), nil
}
