// Package auth integrates OIDC login with short-lived, server-side sessions.
// Game state knows only the derived player ID, never provider tokens or claims.
package auth

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"webscape/server/config"
	"webscape/server/game/model"

	"github.com/alexedwards/scs/v2"
	"github.com/alexedwards/scs/v2/memstore"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

// Session is a revocable authorization lease shared by a browser's sockets.
// WithActive serializes accepted operations against logout and expiry.
type Session struct {
	PlayerID model.EntityId
	Expires  time.Time
	Done     chan struct{}
	mu       sync.Mutex
	revoked  atomic.Bool
}

func (s *Session) WithActive(action func()) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.revoked.Load() || !time.Now().Before(s.Expires) {
		return false
	}
	action()
	return true
}
func (s *Session) Active() bool { return !s.revoked.Load() && time.Now().Before(s.Expires) }
func (s *Session) revoke() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.revoked.Swap(true) {
		close(s.Done)
	}
}

// PlayerID deliberately includes both issuer and subject. The fixed namespace
// and length-safe JSON encoding are a persistent identity contract.
func PlayerID(issuer, subject string) model.EntityId {
	data, _ := json.Marshal([]string{"webscape:oidc:player:v1", issuer, subject})
	hash := sha256.Sum256(data)
	var id uuid.UUID
	copy(id[:], hash[:16])
	id[6] = (id[6] & 0x0f) | 0x80 // UUID v8, custom SHA-256 identity
	id[8] = (id[8] & 0x3f) | 0x80
	return model.EntityId(id)
}

// SCS's in-memory store and cleanup worker live for the process lifetime.
// Each server copies the session configuration and owns/revokes its own leases.
var sessionDefaults = scs.New()

type Manager struct {
	sessions *scs.SessionManager
	store    *memstore.MemStore
	oauth    oauth2.Config
	verifier *oidc.IDTokenVerifier
	client   *http.Client
	origin   string
	issuer   string
	lifetime time.Duration
	// Serialize session HTTP transactions, including one-time callback consumption.
	// Network token exchange is outside this lock.
	mu     sync.Mutex
	leases map[string]*Session
	timers map[string]*time.Timer
	closed bool
}

func New(ctx context.Context, cfg config.AuthConfig, devMode bool) (*Manager, error) {
	if err := cfg.Validate(devMode); err != nil {
		return nil, err
	}
	secret, err := cfg.ClientSecret()
	if err != nil {
		return nil, err
	}
	client := &http.Client{Timeout: 10 * time.Second}
	provider, err := oidc.NewProvider(oidc.ClientContext(ctx, client), cfg.Issuer)
	if err != nil {
		return nil, errors.New("OIDC discovery failed; check issuer and connectivity")
	}
	var metadata struct {
		AuthMethods []string `json:"token_endpoint_auth_methods_supported"`
		JWKS        string   `json:"jwks_uri"`
		Challenges  []string `json:"code_challenge_methods_supported"`
	}
	if err := provider.Claims(&metadata); err != nil {
		return nil, err
	}
	style := oauth2.AuthStyleInHeader
	if len(metadata.AuthMethods) > 0 {
		supported := false
		for _, method := range metadata.AuthMethods {
			if method == "client_secret_basic" {
				supported = true
			}
		}
		if !supported {
			for _, method := range metadata.AuthMethods {
				if method == "client_secret_post" {
					supported = true
					style = oauth2.AuthStyleInParams
				}
			}
		}
		if !supported {
			return nil, errors.New("OIDC provider must support client_secret_basic or client_secret_post")
		}
	}
	if len(metadata.Challenges) > 0 {
		supported := false
		for _, method := range metadata.Challenges {
			if method == "S256" {
				supported = true
			}
		}
		if !supported {
			return nil, errors.New("OIDC provider must support PKCE S256")
		}
	}
	endpoint := provider.Endpoint()
	for _, raw := range []string{endpoint.AuthURL, endpoint.TokenURL, metadata.JWKS} {
		// Apply the same transport restrictions to discovered endpoints.
		copy := cfg
		copy.Issuer = raw
		if err := copy.Validate(devMode); err != nil {
			return nil, errors.New("OIDC provider advertised an unsafe endpoint")
		}
	}
	endpoint.AuthStyle = style
	origin := strings.TrimSuffix(cfg.PublicURL, "/")
	sessionConfig := *sessionDefaults
	sessions := &sessionConfig
	store := sessions.Store.(*memstore.MemStore)
	sessions.Lifetime = time.Duration(cfg.SessionLifetimeSeconds) * time.Second
	sessions.Cookie.Name = "__Host-webscape"
	sessions.Cookie.Secure = true
	if strings.HasPrefix(origin, "http://") {
		sessions.Cookie.Name = "webscape_local"
		sessions.Cookie.Secure = false
	}
	sessions.Cookie.HttpOnly = true
	sessions.Cookie.SameSite = http.SameSiteLaxMode
	sessions.Cookie.Persist = false
	return &Manager{sessions: sessions, store: store, client: client, origin: origin, issuer: cfg.Issuer,
		lifetime: sessions.Lifetime, leases: make(map[string]*Session), timers: make(map[string]*time.Timer),
		oauth:    oauth2.Config{ClientID: cfg.ClientID, ClientSecret: secret, Endpoint: endpoint, RedirectURL: origin + "/auth/callback", Scopes: []string{oidc.ScopeOpenID}},
		verifier: provider.Verifier(&oidc.Config{ClientID: cfg.ClientID}),
	}, nil
}

func (m *Manager) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	for token := range m.leases {
		m.revoke(token)
	}
}
func (m *Manager) revoke(token string) {
	if s := m.leases[token]; s != nil {
		s.revoke()
		delete(m.leases, token)
	}
	if timer := m.timers[token]; timer != nil {
		timer.Stop()
		delete(m.timers, token)
	}
	_ = m.store.Delete(token)
}
func (m *Manager) lease(ctx context.Context) *Session {
	s := m.leases[m.sessions.Token(ctx)]
	if s == nil || !s.WithActive(func() {}) {
		return nil
	}
	return s
}
func (m *Manager) sessionHandler(fn http.HandlerFunc) http.Handler {
	wrapped := m.sessions.LoadAndSave(fn)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.mu.Lock()
		defer m.mu.Unlock()
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Referrer-Policy", "no-referrer")
		if m.closed {
			http.Error(w, "Authentication unavailable", http.StatusServiceUnavailable)
			return
		}
		wrapped.ServeHTTP(w, r)
	})
}
func (m *Manager) RegisterRoutes(mux *http.ServeMux) {
	mux.Handle("GET /auth/login", m.sessionHandler(m.login))
	mux.HandleFunc("GET /auth/callback", m.callback)
	mux.Handle("GET /auth/session", m.sessionHandler(m.status))
	mux.Handle("POST /auth/logout", m.sessionHandler(m.logout))
}
func (m *Manager) login(w http.ResponseWriter, r *http.Request) {
	// A new login invalidates any previous account in this browser first.
	m.revoke(m.sessions.Token(r.Context()))
	if err := m.sessions.RenewToken(r.Context()); err != nil {
		http.Error(w, "Login unavailable", 500)
		return
	}
	state, nonce, verifier := oauth2.GenerateVerifier(), oauth2.GenerateVerifier(), oauth2.GenerateVerifier()
	m.sessions.Put(r.Context(), "state", state)
	m.sessions.Put(r.Context(), "nonce", nonce)
	m.sessions.Put(r.Context(), "verifier", verifier)
	m.sessions.SetDeadline(r.Context(), time.Now().Add(5*time.Minute))
	http.Redirect(w, r, m.oauth.AuthCodeURL(state, oauth2.S256ChallengeOption(verifier), oidc.Nonce(nonce)), http.StatusFound)
}
func (m *Manager) callback(w http.ResponseWriter, r *http.Request) {
	var nonce, verifier string
	valid := false
	// Consume and commit the transaction before network I/O. Concurrent or replayed
	// callbacks cannot reuse it, even when the provider is slow or returns an error.
	m.sessionHandler(func(w http.ResponseWriter, r *http.Request) {
		expected := m.sessions.GetString(r.Context(), "state")
		if expected == "" || subtle.ConstantTimeCompare([]byte(expected), []byte(r.URL.Query().Get("state"))) != 1 {
			return
		}
		nonce = m.sessions.PopString(r.Context(), "nonce")
		verifier = m.sessions.PopString(r.Context(), "verifier")
		m.sessions.Remove(r.Context(), "state")
		m.sessions.Put(r.Context(), "completing", expected)
		valid = nonce != "" && verifier != "" && r.URL.Query().Get("code") != "" && r.URL.Query().Get("error") == ""
	}).ServeHTTP(w, r)
	if !valid {
		m.loginFailed(w, r)
		return
	}
	ctx, cancel := context.WithTimeout(oidc.ClientContext(r.Context(), m.client), 10*time.Second)
	defer cancel()
	token, err := m.oauth.Exchange(ctx, r.URL.Query().Get("code"), oauth2.VerifierOption(verifier))
	if err != nil {
		m.loginFailed(w, r)
		return
	}
	raw, _ := token.Extra("id_token").(string)
	id, err := m.verifier.Verify(ctx, raw)
	if err != nil || id.Subject == "" || subtle.ConstantTimeCompare([]byte(id.Nonce), []byte(nonce)) != 1 {
		m.loginFailed(w, r)
		return
	}
	m.sessionHandler(func(w http.ResponseWriter, r *http.Request) {
		if m.sessions.PopString(r.Context(), "completing") != r.URL.Query().Get("state") {
			m.loginFailed(w, r)
			return
		}
		m.revoke(m.sessions.Token(r.Context()))
		if err := m.sessions.RenewToken(r.Context()); err != nil {
			m.loginFailed(w, r)
			return
		}
		expires := time.Now().Add(m.lifetime)
		if id.Expiry.Before(expires) {
			expires = id.Expiry
		}
		m.sessions.SetDeadline(r.Context(), expires)
		m.sessions.Put(r.Context(), "csrf", oauth2.GenerateVerifier())
		key := m.sessions.Token(r.Context())
		s := &Session{PlayerID: PlayerID(m.issuer, id.Subject), Expires: expires, Done: make(chan struct{})}
		m.leases[key] = s
		m.timers[key] = time.AfterFunc(time.Until(expires), func() { m.mu.Lock(); defer m.mu.Unlock(); m.revoke(key) })
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}).ServeHTTP(w, r)
}
func (m *Manager) loginFailed(w http.ResponseWriter, r *http.Request) {
	// Never echo provider errors, codes or tokens to the browser or logs.
	http.Redirect(w, r, "/?login=failed", http.StatusSeeOther)
}
func (m *Manager) status(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	s := m.lease(r.Context())
	if s == nil {
		json.NewEncoder(w).Encode(map[string]any{"authenticated": false})
		return
	}
	json.NewEncoder(w).Encode(map[string]any{"authenticated": true, "accountId": s.PlayerID.String(), "csrfToken": m.sessions.GetString(r.Context(), "csrf"), "expiresAt": s.Expires})
}
func (m *Manager) logout(w http.ResponseWriter, r *http.Request) {
	csrf := m.sessions.GetString(r.Context(), "csrf")
	if !m.CheckOrigin(r) || csrf == "" || subtle.ConstantTimeCompare([]byte(csrf), []byte(r.Header.Get("X-CSRF-Token"))) != 1 {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	m.revoke(m.sessions.Token(r.Context()))
	if err := m.sessions.Destroy(r.Context()); err != nil {
		http.Error(w, "Logout failed", 500)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
func (m *Manager) CheckOrigin(r *http.Request) bool {
	origins := r.Header.Values("Origin")
	if len(origins) != 1 {
		return false
	}
	u, err := url.Parse(origins[0])
	return err == nil && u.String() == m.origin
}

// RequireSocket checks cookies without writing them during WebSocket hijacking.
func (m *Manager) RequireSocket(next func(http.ResponseWriter, *http.Request, *Session)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.mu.Lock()
		cookie, err := r.Cookie(m.sessions.Cookie.Name)
		var s *Session
		if err == nil && !m.closed {
			ctx, loadErr := m.sessions.Load(r.Context(), cookie.Value)
			if loadErr == nil {
				s = m.lease(ctx)
			}
		}
		m.mu.Unlock()
		if r.Method != http.MethodGet || !m.CheckOrigin(r) {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		if s == nil {
			http.Error(w, "Sign in required", http.StatusUnauthorized)
			return
		}
		next(w, r, s)
	})
}
