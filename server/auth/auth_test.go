package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"
	"webscape/server/internal/oidctest"
)

type fixture struct {
	manager  *Manager
	server   *httptest.Server
	provider *oidctest.Provider
	browser  *http.Client
}

func setup(t *testing.T) fixture {
	t.Helper()
	p := oidctest.New(t)
	server := httptest.NewUnstartedServer(nil)
	cfg := p.Config("http://" + server.Listener.Addr().String())
	m, err := New(context.Background(), cfg, true)
	if err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	m.RegisterRoutes(mux)
	mux.Handle("/ws", m.RequireSocket(func(w http.ResponseWriter, r *http.Request, s *Session) { w.WriteHeader(204) }))
	server.Config.Handler = mux
	server.Start()
	t.Cleanup(server.Close)
	t.Cleanup(m.Close)
	return fixture{m, server, p, oidctest.Browser()}
}
func sessionStatus(t *testing.T, c *http.Client, app string) map[string]any {
	t.Helper()
	r, err := c.Get(app + "/auth/session")
	if err != nil {
		t.Fatal(err)
	}
	defer r.Body.Close()
	if r.Header.Get("Cache-Control") != "no-store" {
		t.Fatal("session response is cacheable")
	}
	var value map[string]any
	if err := json.NewDecoder(r.Body).Decode(&value); err != nil {
		t.Fatal(err)
	}
	return value
}
func TestLoginIdentityCookiesAndLogout(t *testing.T) {
	f := setup(t)
	app := f.server.URL
	oidctest.Login(t, f.browser, app, "one")
	status := sessionStatus(t, f.browser, app)
	if status["authenticated"] != true || status["accountId"] != PlayerID(f.provider.URL, "one").String() {
		t.Fatalf("status: %v", status)
	}
	// The session cookie contains no account ID, provider token or browser-readable credentials.
	r := oidctest.Get(t, f.browser, app+"/auth/login")
	for _, c := range r.Cookies() {
		if !c.HttpOnly || c.SameSite != http.SameSiteLaxMode || c.Path != "/" {
			t.Fatalf("cookie flags: %+v", c)
		}
	}
	oidctest.Login(t, f.browser, app, "one")
	status = sessionStatus(t, f.browser, app)
	u, _ := url.Parse(app)
	cookie := f.browser.Jar.Cookies(u)[0]
	f.manager.mu.Lock()
	lease := f.manager.leases[cookie.Value]
	f.manager.mu.Unlock()
	if lease == nil {
		t.Fatal("no authorization lease")
	}
	for _, test := range []struct {
		origin, csrf string
		want         int
	}{
		{"https://evil.example", status["csrfToken"].(string), 403}, {app, "wrong", 403}, {app, status["csrfToken"].(string), 204},
	} {
		req, _ := http.NewRequest("POST", app+"/auth/logout", nil)
		req.Header.Set("Origin", test.origin)
		req.Header.Set("X-CSRF-Token", test.csrf)
		r, err := f.browser.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		r.Body.Close()
		if r.StatusCode != test.want {
			t.Fatalf("logout: %d", r.StatusCode)
		}
	}
	select {
	case <-lease.Done:
	default:
		t.Fatal("logout did not revoke socket lease")
	}
	if lease.WithActive(func() { t.Fatal("revoked operation ran") }) {
		t.Fatal("lease still active")
	}
	if sessionStatus(t, f.browser, app)["authenticated"] != false {
		t.Fatal("logout retained session")
	}
	replay := oidctest.Browser()
	replay.Jar.SetCookies(u, []*http.Cookie{cookie})
	if sessionStatus(t, replay, app)["authenticated"] != false {
		t.Fatal("old cookie replay authenticated")
	}
}
func TestCallbackRejectsInvalidClaims(t *testing.T) {
	for _, test := range []struct {
		name   string
		mutate func(map[string]any)
	}{
		{"issuer", func(c map[string]any) { c["iss"] = "https://other.example" }},
		{"audience", func(c map[string]any) { c["aud"] = "another-client" }},
		{"expiry", func(c map[string]any) { c["exp"] = time.Now().Add(-time.Hour).Unix() }},
		{"nonce", func(c map[string]any) { c["nonce"] = "wrong" }},
		{"subject", func(c map[string]any) { c["sub"] = "" }},
	} {
		t.Run(test.name, func(t *testing.T) {
			f := setup(t)
			f.provider.MutateClaims = test.mutate
			callback := oidctest.Callback(t, f.browser, f.server.URL, "one")
			r := oidctest.Get(t, f.browser, callback)
			if r.Header.Get("Location") != "/?login=failed" || sessionStatus(t, f.browser, f.server.URL)["authenticated"] != false {
				t.Fatal("invalid token authenticated")
			}
		})
	}
}
func TestStateBindingAndSingleUse(t *testing.T) {
	f := setup(t)
	callback := oidctest.Callback(t, f.browser, f.server.URL, "one")
	stranger := oidctest.Browser()
	if r := oidctest.Get(t, stranger, callback); r.Header.Get("Location") != "/?login=failed" {
		t.Fatal("unbound callback accepted")
	}
	u, _ := url.Parse(callback)
	q := u.Query()
	q.Set("state", "wrong")
	u.RawQuery = q.Encode()
	if r := oidctest.Get(t, f.browser, u.String()); r.Header.Get("Location") != "/?login=failed" {
		t.Fatal("wrong state accepted")
	}
	// Race two callbacks carrying the same transaction cookie.
	appURL, _ := url.Parse(f.server.URL)
	cookie := f.browser.Jar.Cookies(appURL)
	clients := []*http.Client{oidctest.Browser(), oidctest.Browser()}
	for _, c := range clients {
		c.Jar.SetCookies(appURL, cookie)
	}
	var wg sync.WaitGroup
	statuses := make(chan string, 2)
	for _, c := range clients {
		wg.Add(1)
		go func(c *http.Client) {
			defer wg.Done()
			r, err := c.Get(callback)
			if err != nil {
				statuses <- "error"
				return
			}
			r.Body.Close()
			statuses <- r.Header.Get("Location")
		}(c)
	}
	wg.Wait()
	close(statuses)
	successes := 0
	for location := range statuses {
		if location == "/" {
			successes++
		}
	}
	if successes != 1 {
		t.Fatalf("single-use callback successes=%d", successes)
	}
	if r := oidctest.Get(t, f.browser, callback); r.Header.Get("Location") != "/?login=failed" {
		t.Fatal("replayed callback accepted")
	}
}
func TestSocketRequiresSessionAndExactOrigin(t *testing.T) {
	f := setup(t)
	for _, test := range []struct {
		origin string
		auth   bool
		want   int
	}{
		{f.server.URL, false, 401}, {"", false, 403}, {"https://evil.example", true, 403}, {f.server.URL + "/", true, 403}, {f.server.URL, true, 204},
	} {
		browser := oidctest.Browser()
		if test.auth {
			oidctest.Login(t, browser, f.server.URL, "one")
		}
		req, _ := http.NewRequest("GET", f.server.URL+"/ws", nil)
		req.Header.Set("Origin", test.origin)
		r, err := browser.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		r.Body.Close()
		if r.StatusCode != test.want {
			t.Fatalf("origin=%q auth=%v status=%d", test.origin, test.auth, r.StatusCode)
		}
	}
}
func TestSessionExpiresAndIdentityIsNamespaced(t *testing.T) {
	f := setup(t)
	f.manager.lifetime = 150 * time.Millisecond
	oidctest.Login(t, f.browser, f.server.URL, "one")
	f.manager.mu.Lock()
	var lease *Session
	for _, s := range f.manager.leases {
		lease = s
	}
	f.manager.mu.Unlock()
	if lease == nil {
		t.Fatal("missing lease")
	}
	select {
	case <-lease.Done:
	case <-time.After(time.Second):
		t.Fatal("expiry did not revoke lease")
	}
	if sessionStatus(t, f.browser, f.server.URL)["authenticated"] != false {
		t.Fatal("expired session remained valid")
	}
	other := setup(t)
	oidctest.Login(t, other.browser, other.server.URL, "one")
	a := PlayerID(f.provider.URL, "one")
	b := PlayerID(other.provider.URL, "one")
	if a == b || a == PlayerID(f.provider.URL, "two") || a != PlayerID(f.provider.URL, "one") {
		t.Fatal("identity collision or instability")
	}
	if sessionStatus(t, other.browser, other.server.URL)["accountId"] != b.String() {
		t.Fatal("second issuer did not work")
	}
}
func TestProductionCookieAndMissingSecret(t *testing.T) {
	f := setup(t)
	cfg := f.provider.Config("https://game.example.com")
	m, err := New(context.Background(), cfg, true)
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	mux := http.NewServeMux()
	m.RegisterRoutes(mux)
	r := httptest.NewRecorder()
	mux.ServeHTTP(r, httptest.NewRequest("GET", "https://game.example.com/auth/login", nil))
	cookies := r.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Name != "__Host-webscape" || !cookies[0].Secure || !cookies[0].HttpOnly || cookies[0].Domain != "" {
		t.Fatal("insecure production cookie")
	}
	if strings.Contains(r.Header().Get("Location"), oidctest.Secret) {
		t.Fatal("secret leaked in redirect")
	}
	t.Setenv(oidctest.SecretEnv, "")
	if _, err := New(context.Background(), cfg, true); err == nil {
		t.Fatal("missing secret did not fail closed")
	}
}

func TestRejectsBadSignatureAndExpiredTransaction(t *testing.T) {
	f := setup(t)
	f.provider.BadSignature = true
	callback := oidctest.Callback(t, f.browser, f.server.URL, "one")
	if r := oidctest.Get(t, f.browser, callback); r.Header.Get("Location") != "/?login=failed" {
		t.Fatal("bad signature accepted")
	}
	f.provider.BadSignature = false
	callback = oidctest.Callback(t, f.browser, f.server.URL, "one")
	u, _ := url.Parse(f.server.URL)
	cookie := f.browser.Jar.Cookies(u)[0]
	f.manager.mu.Lock()
	ctx, err := f.manager.sessions.Load(context.Background(), cookie.Value)
	if err == nil {
		f.manager.sessions.SetDeadline(ctx, time.Now().Add(-time.Second))
		_, _, err = f.manager.sessions.Commit(ctx)
	}
	f.manager.mu.Unlock()
	if err != nil {
		t.Fatal(err)
	}
	if r := oidctest.Get(t, f.browser, callback); r.Header.Get("Location") != "/?login=failed" {
		t.Fatal("expired callback accepted")
	}
}

func TestPlayerIdentityContract(t *testing.T) {
	// Saved characters must retain this ID across application upgrades.
	if got := PlayerID("https://issuer.example.com", "person-1").String(); got != "83adc583-b9fc-874b-b79a-08d40f1af95e" {
		t.Fatalf("persistent identity changed: %s", got)
	}
}
