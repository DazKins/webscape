package auth

import (
	"context"
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
		t.Fatal("reload lost guest identity")
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
