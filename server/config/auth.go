package config

import (
	"errors"
	"net"
	"net/url"
	"os"
	"strings"
)

// AuthConfig describes a confidential OIDC web client. Credentials are resolved
// only at startup, never exposed by the browser configuration endpoint.
type AuthConfig struct {
	Issuer                 string `json:"issuer"`
	ClientID               string `json:"clientId"`
	ClientSecretEnv        string `json:"clientSecretEnv"`
	PublicURL              string `json:"publicUrl"`
	SessionLifetimeSeconds int    `json:"sessionLifetimeSeconds"`
}

func (a AuthConfig) Validate(devMode bool) error {
	if a.ClientID == "" || a.ClientSecretEnv == "" {
		return errors.New("config auth.clientId and auth.clientSecretEnv are required")
	}
	if a.SessionLifetimeSeconds < 60 || a.SessionLifetimeSeconds > 86400 {
		return errors.New("config auth.sessionLifetimeSeconds must be between 60 and 86400")
	}
	for _, raw := range []string{a.Issuer, a.PublicURL} {
		u, err := url.Parse(raw)
		if err != nil || u.Hostname() == "" || u.User != nil || u.RawQuery != "" || u.ForceQuery || u.Fragment != "" {
			return errors.New("auth URLs must be absolute URLs without credentials, query or fragment")
		}
		ip := net.ParseIP(u.Hostname())
		loopback := u.Hostname() == "localhost" || (ip != nil && ip.IsLoopback())
		if u.Scheme != "https" && !(devMode && u.Scheme == "http" && loopback) {
			return errors.New("auth URLs require HTTPS (HTTP loopback is allowed only with server.devMode)")
		}
	}
	u, _ := url.Parse(a.PublicURL)
	if u.Path != "" && u.Path != "/" {
		return errors.New("auth.publicUrl must be an origin without a path")
	}
	return nil
}

func (a AuthConfig) ClientSecret() (string, error) {
	value := os.Getenv(a.ClientSecretEnv)
	if strings.TrimSpace(value) == "" {
		return "", errors.New("configured OIDC client secret environment variable is empty")
	}
	return value, nil
}
