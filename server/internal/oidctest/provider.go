// Package oidctest is a local OIDC protocol fixture for integration tests.
// It is never used by the runtime and is not an authentication bypass.
package oidctest

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"
	"webscape/server/config"

	jose "github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
	"golang.org/x/oauth2"
)

const ClientID = "webscape-test"
const Secret = "test-fixture-only"
const SecretEnv = "WEBSCAPE_TEST_OIDC_SECRET"

type code struct{ nonce, challenge, redirect, subject string }
type Provider struct {
	URL    string
	server *httptest.Server
	key    *rsa.PrivateKey
	signer jose.Signer
	mu     sync.Mutex
	codes  map[string]code
	// Set before making requests; tests can exercise invalid provider responses.
	MutateClaims func(map[string]any)
	BadSignature bool
}

func New(t *testing.T) *Provider {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.RS256, Key: jose.JSONWebKey{Key: key, KeyID: "test-key"}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	p := &Provider{key: key, signer: signer, codes: make(map[string]code)}
	p.server = httptest.NewServer(http.HandlerFunc(p.serve))
	p.URL = p.server.URL
	t.Cleanup(p.server.Close)
	t.Setenv(SecretEnv, Secret)
	return p
}
func (p *Provider) Config(publicURL string) config.AuthConfig {
	return config.AuthConfig{Issuer: p.URL, ClientID: ClientID, ClientSecretEnv: SecretEnv, PublicURL: publicURL, SessionLifetimeSeconds: 3600}
}
func (p *Provider) serve(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/.well-known/openid-configuration":
		json.NewEncoder(w).Encode(map[string]any{"issuer": p.URL, "authorization_endpoint": p.URL + "/oauth/start", "token_endpoint": p.URL + "/oauth/exchange", "jwks_uri": p.URL + "/keys", "response_types_supported": []string{"code"}, "subject_types_supported": []string{"public"}, "id_token_signing_alg_values_supported": []string{"RS256"}, "token_endpoint_auth_methods_supported": []string{"client_secret_basic"}, "code_challenge_methods_supported": []string{"S256"}})
	case "/keys":
		json.NewEncoder(w).Encode(jose.JSONWebKeySet{Keys: []jose.JSONWebKey{{Key: &p.key.PublicKey, KeyID: "test-key", Algorithm: "RS256", Use: "sig"}}})
	case "/oauth/start":
		q := r.URL.Query()
		if q.Get("client_id") != ClientID || q.Get("scope") != "openid" || q.Get("response_type") != "code" || q.Get("code_challenge_method") != "S256" || q.Get("nonce") == "" {
			http.Error(w, "invalid authorization", 400)
			return
		}
		subject := q.Get("subject")
		if subject == "" {
			http.Error(w, "fixture requires subject", 400)
			return
		}
		value := oauth2.GenerateVerifier()
		p.mu.Lock()
		p.codes[value] = code{q.Get("nonce"), q.Get("code_challenge"), q.Get("redirect_uri"), subject}
		p.mu.Unlock()
		redirect, _ := url.Parse(q.Get("redirect_uri"))
		params := redirect.Query()
		params.Set("state", q.Get("state"))
		params.Set("code", value)
		redirect.RawQuery = params.Encode()
		http.Redirect(w, r, redirect.String(), 302)
	case "/oauth/exchange":
		r.ParseForm()
		id, secret, ok := r.BasicAuth()
		p.mu.Lock()
		c, found := p.codes[r.Form.Get("code")]
		delete(p.codes, r.Form.Get("code"))
		p.mu.Unlock()
		hash := sha256.Sum256([]byte(r.Form.Get("code_verifier")))
		if !ok || id != ClientID || secret != Secret || !found || r.Form.Get("redirect_uri") != c.redirect || base64.RawURLEncoding.EncodeToString(hash[:]) != c.challenge {
			http.Error(w, "invalid exchange", 400)
			return
		}
		claims := map[string]any{"iss": p.URL, "sub": c.subject, "aud": ClientID, "iat": time.Now().Unix(), "exp": time.Now().Add(time.Hour).Unix(), "nonce": c.nonce}
		if p.MutateClaims != nil {
			p.MutateClaims(claims)
		}
		raw, err := jwt.Signed(p.signer).Claims(claims).Serialize()
		if err != nil {
			http.Error(w, "signing error", 500)
			return
		}
		if p.BadSignature {
			parts := strings.Split(raw, ".")
			parts[2] = strings.Repeat("A", len(parts[2]))
			raw = strings.Join(parts, ".")
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"access_token": "fixture-access", "token_type": "Bearer", "id_token": raw})
	default:
		http.NotFound(w, r)
	}
}
func Browser() *http.Client {
	jar, _ := cookiejar.New(nil)
	return &http.Client{Jar: jar, Timeout: 10 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
}
func Get(t *testing.T, c *http.Client, address string) *http.Response {
	t.Helper()
	r, err := c.Get(address)
	if err != nil {
		t.Fatal(err)
	}
	r.Body.Close()
	return r
}

// Callback returns a valid callback URL without consuming it at the application.
func Callback(t *testing.T, c *http.Client, app, subject string) string {
	t.Helper()
	r := Get(t, c, app+"/auth/login")
	if r.StatusCode != 302 {
		t.Fatalf("login status %d", r.StatusCode)
	}
	u, err := url.Parse(r.Header.Get("Location"))
	if err != nil {
		t.Fatal(err)
	}
	q := u.Query()
	q.Set("subject", subject)
	u.RawQuery = q.Encode()
	r = Get(t, c, u.String())
	if r.StatusCode != 302 {
		t.Fatalf("authorize status %d", r.StatusCode)
	}
	return r.Header.Get("Location")
}
func Login(t *testing.T, c *http.Client, app, subject string) {
	t.Helper()
	r := Get(t, c, Callback(t, c, app, subject))
	if r.StatusCode != 303 || r.Header.Get("Location") != "/" {
		t.Fatalf("callback status=%d location=%s", r.StatusCode, r.Header.Get("Location"))
	}
}
