package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"webscape/server/config"
	"webscape/server/internal/oidctest"
)

func TestGuestSessionsWithoutProvider(t *testing.T) {
	app := httptest.NewUnstartedServer(nil)
	m, err := New(context.Background(), config.AuthConfig{Mode: "none", PublicURL: "http://" + app.Listener.Addr().String(), SessionLifetimeSeconds: 3600}, true)
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	mux := http.NewServeMux()
	m.RegisterRoutes(mux)
	mux.Handle("/ws", m.RequireSocket(func(w http.ResponseWriter, r *http.Request, s *Session) { w.WriteHeader(204) }))
	app.Config.Handler = mux
	app.Start()
	defer app.Close()
	a, b := oidctest.Browser(), oidctest.Browser()
	if status := sessionStatus(t, a, app.URL); status["guest"] != true || status["authenticated"] != false {
		t.Fatal(status)
	}
	for _, browser := range []*http.Client{a, b} {
		res, err := browser.Get(app.URL + "/auth/login")
		if err != nil {
			t.Fatal(err)
		}
		res.Body.Close()
	}
	first, second := sessionStatus(t, a, app.URL), sessionStatus(t, b, app.URL)
	if first["authenticated"] != true || first["accountId"] == second["accountId"] {
		t.Fatal(first, second)
	}
	if again := sessionStatus(t, a, app.URL); again["accountId"] != first["accountId"] {
		t.Fatal("connection retry lost guest identity")
	}
	u, _ := url.Parse(app.URL)
	var token string
	for _, c := range a.Jar.Cookies(u) {
		if c.Name == "webscape_local_guest" {
			token = c.Value
		}
	}
	lease := m.leases[token]
	if lease == nil {
		t.Fatal("missing guest lease")
	}
	for _, test := range []struct {
		origin string
		want   int
	}{{"https://evil.example", 403}, {app.URL, 204}} {
		req, _ := http.NewRequest("GET", app.URL+"/ws", nil)
		req.Header.Set("Origin", test.origin)
		res, err := a.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		res.Body.Close()
		if res.StatusCode != test.want {
			t.Fatal(res.StatusCode)
		}
	}
	req, _ := http.NewRequest("POST", app.URL+"/auth/logout", nil)
	req.Header.Set("Origin", app.URL)
	req.Header.Set("X-CSRF-Token", first["csrfToken"].(string))
	res, err := a.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode != 204 || lease.Active() {
		t.Fatal("guest logout did not revoke lease")
	}
	if status := sessionStatus(t, a, app.URL); status["authenticated"] != false {
		t.Fatal(status)
	}
	res, err = a.Get(app.URL + "/auth/login")
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if status := sessionStatus(t, a, app.URL); status["accountId"] == first["accountId"] {
		t.Fatal("new guest claimed old identity")
	}
	res, err = a.Get(app.URL + "/auth/callback?code=ignored")
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode != 404 {
		t.Fatal("guest mode exposed OIDC callback")
	}
}

func TestGuestStartsWithoutNavigation(t *testing.T) {
	f := setup(t, true)
	for _, origin := range []string{"", "https://evil.example", f.server.URL} {
		req, _ := http.NewRequest(http.MethodPost, f.server.URL+"/auth/guest", nil)
		if origin != "" {
			req.Header.Set("Origin", origin)
		}
		res, err := f.browser.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		if origin != f.server.URL {
			res.Body.Close()
			if res.StatusCode != http.StatusForbidden {
				t.Fatalf("origin %q: got %d", origin, res.StatusCode)
			}
			if sessionStatus(t, f.browser, f.server.URL)["authenticated"] != false {
				t.Fatal("rejected request created a session")
			}
			continue
		}
		var session map[string]any
		err = json.NewDecoder(res.Body).Decode(&session)
		res.Body.Close()
		if err != nil || res.StatusCode != http.StatusOK || session["authenticated"] != true || session["guest"] != true || session["accountId"] == "" || session["csrfToken"] == "" {
			t.Fatalf("guest start: status %d, session %v, error %v", res.StatusCode, session, err)
		}
		if again := sessionStatus(t, f.browser, f.server.URL); again["accountId"] != session["accountId"] {
			t.Fatal("connection retry changed guest identity")
		}
		req, _ = http.NewRequest(http.MethodPost, f.server.URL+"/auth/logout", nil)
		req.Header.Set("Origin", f.server.URL)
		req.Header.Set("X-CSRF-Token", session["csrfToken"].(string))
		res, err = f.browser.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		res.Body.Close()
		if res.StatusCode != http.StatusNoContent || sessionStatus(t, f.browser, f.server.URL)["authenticated"] != false {
			t.Fatal("page restart could not clear guest session")
		}
	}
}

func TestGuestAndSignInTogether(t *testing.T) {
	f := setup(t, true)
	app := f.server.URL
	status := sessionStatus(t, f.browser, app)
	if status["allowGuests"] != true || status["signInEnabled"] != true || status["authenticated"] != false {
		t.Fatal(status)
	}
	res := oidctest.Get(t, f.browser, app+"/auth/guest")
	for _, cookie := range res.Cookies() {
		if !cookie.Expires.IsZero() || cookie.MaxAge > 0 {
			t.Fatal("guest cookie persisted")
		}
	}
	guest := sessionStatus(t, f.browser, app)
	if guest["guest"] != true || guest["authenticated"] != true {
		t.Fatal(guest)
	}
	u, _ := url.Parse(app)
	token := f.browser.Jar.Cookies(u)[0].Value
	lease := f.manager.leases[token]
	if lease == nil || !lease.Guest {
		t.Fatal("guest lease missing")
	}

	oidctest.Login(t, f.browser, app, "one")
	signedIn := sessionStatus(t, f.browser, app)
	if signedIn["guest"] != false || signedIn["accountId"] != PlayerID(f.provider.URL, "one").String() || lease.Active() {
		t.Fatal("guest to account transition failed", signedIn)
	}
	oidctest.Get(t, f.browser, app+"/auth/guest")
	next := sessionStatus(t, f.browser, app)
	if next["guest"] != true || next["accountId"] == guest["accountId"] || next["accountId"] == signedIn["accountId"] {
		t.Fatal("guest reused previous progress", next)
	}
	oidctest.Login(t, f.browser, app, "one")
	if sessionStatus(t, f.browser, app)["accountId"] != signedIn["accountId"] {
		t.Fatal("account identity changed")
	}
}

func TestGuestEndpointDisabled(t *testing.T) {
	f := setup(t)
	res := oidctest.Get(t, f.browser, f.server.URL+"/auth/guest")
	if res.StatusCode != http.StatusNotFound {
		t.Fatal(res.StatusCode)
	}
	if sessionStatus(t, f.browser, f.server.URL)["authenticated"] != false {
		t.Fatal("guest login accepted")
	}
}
