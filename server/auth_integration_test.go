package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
	"webscape/server/auth"
	"webscape/server/command"
	"webscape/server/game/model"
	"webscape/server/internal/oidctest"
	"webscape/server/message"

	"github.com/gorilla/websocket"
)

func TestAuthenticatedWebSocketOwnershipAndRevocation(t *testing.T) {
	provider := oidctest.New(t)
	app := httptest.NewUnstartedServer(nil)
	manager, err := auth.New(context.Background(), provider.Config("http://"+app.Listener.Addr().String()), true)
	if err != nil {
		t.Fatal(err)
	}
	defer manager.Close()
	game := newCommandHandlerTestGame(t)
	game.RetainOfflinePlayers()
	ws := NewWsServer()
	defer ws.Close()
	handler := NewClientCommandHandler(game, ws.PlayerID)
	game.RegisterSender(ws.SendToClient)
	ws.SetConnectHandler(game.HandleConnect)
	ws.SetDisconnectHandler(func(id string) { game.HandleLeave(id) })
	ws.SetIncomingMessageHandler(func(id, raw string) {
		cmd, err := command.Unmarshal(raw)
		if err == nil {
			handler.HandleCommand(id, cmd)
		}
	})
	mux := http.NewServeMux()
	manager.RegisterRoutes(mux)
	mux.Handle("/ws", manager.RequireSocket(ws.HandleWebSocket))
	app.Config.Handler = mux
	app.Start()
	defer app.Close()
	socketURL := strings.Replace(app.URL, "http://", "ws://", 1) + "/ws"
	headers := http.Header{"Origin": {app.URL}}
	if conn, r, err := websocket.DefaultDialer.Dial(socketURL, headers); err == nil {
		conn.Close()
		t.Fatal("unauthenticated socket upgraded")
	} else if r.StatusCode != 401 {
		t.Fatalf("unauthenticated status %d", r.StatusCode)
	}
	browser := oidctest.Browser()
	oidctest.Login(t, browser, app.URL, "alice")
	req, _ := http.NewRequest("GET", app.URL, nil)
	for _, c := range browser.Jar.Cookies(req.URL) {
		req.AddCookie(c)
	}
	req.Header.Set("Origin", app.URL)
	conn, _, err := websocket.DefaultDialer.Dial(socketURL, req.Header)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	readType := func(conn *websocket.Conn, kind string) map[string]any {
		t.Helper()
		conn.SetReadDeadline(time.Now().Add(3 * time.Second))
		for {
			var msg struct {
				Metadata struct {
					Type string `json:"type"`
				} `json:"metadata"`
				Data map[string]any `json:"data"`
			}
			if err := conn.ReadJSON(&msg); err != nil {
				t.Fatal(err)
			}
			if msg.Metadata.Type == kind {
				return msg.Data
			}
		}
	}
	aliceID := auth.PlayerID(provider.URL, "alice").String()
	sendTestCommand(t, conn, "register", map[string]any{"id": auth.PlayerID(provider.URL, "bob").String(), "name": "Alice"})
	readType(conn, "registrationFailed")
	sendTestCommand(t, conn, "register", map[string]any{"name": "Alice"})
	if data := readType(conn, "registered"); data["entityId"] != aliceID {
		t.Fatalf("wrong authenticated identity: %v", data)
	}
	// A second socket cannot register the same character while it is active.
	second, _, err := websocket.DefaultDialer.Dial(socketURL, req.Header)
	if err != nil {
		t.Fatal(err)
	}
	defer second.Close()
	sendTestCommand(t, second, "register", map[string]any{"name": "Alice"})
	readType(second, "registrationFailed")
	// Malformed authenticated commands cannot panic a connection goroutine.
	for _, kind := range []string{"move", "chat", "interact", "equip", "unequip", "drop", "trade", "conversationOption"} {
		sendTestCommand(t, conn, kind, map[string]any{})
	}
	response, err := browser.Get(app.URL + "/auth/session")
	if err != nil {
		t.Fatal(err)
	}
	var status struct {
		CSRF string `json:"csrfToken"`
	}
	json.NewDecoder(response.Body).Decode(&status)
	response.Body.Close()
	logout, _ := http.NewRequest("POST", app.URL+"/auth/logout", nil)
	logout.Header.Set("Origin", app.URL)
	logout.Header.Set("X-CSRF-Token", status.CSRF)
	response, err = browser.Do(logout)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != 204 {
		t.Fatal("logout failed")
	}
	for _, socket := range []*websocket.Conn{conn, second} {
		socket.SetReadDeadline(time.Now().Add(2 * time.Second))
		for {
			_, _, err := socket.ReadMessage()
			if err != nil {
				if !websocket.IsCloseError(err, 4001) {
					t.Fatalf("expected session close: %v", err)
				}
				break
			}
		}
	}
	// Old cookies cannot reconnect or receive a new world snapshot.
	if old, _, err := websocket.DefaultDialer.Dial(socketURL, req.Header); err == nil {
		old.Close()
		t.Fatal("logged-out cookie reconnected")
	}
	// A short-lived provider token also closes an otherwise idle socket.
	provider.MutateClaims = func(claims map[string]any) { claims["exp"] = time.Now().Add(2 * time.Second).Unix() }
	oidctest.Login(t, browser, app.URL, "bob")
	req, _ = http.NewRequest("GET", app.URL, nil)
	for _, c := range browser.Jar.Cookies(req.URL) {
		req.AddCookie(c)
	}
	req.Header.Set("Origin", app.URL)
	expiring, _, err := websocket.DefaultDialer.Dial(socketURL, req.Header)
	if err != nil {
		t.Fatal(err)
	}
	defer expiring.Close()
	sendTestCommand(t, expiring, "register", map[string]any{"name": "Bob"})
	if data := readType(expiring, "registered"); data["entityId"] == aliceID {
		t.Fatal("Bob claimed Alice")
	}
	expiring.SetReadDeadline(time.Now().Add(4 * time.Second))
	for {
		_, _, err := expiring.ReadMessage()
		if err != nil {
			if !websocket.IsCloseError(err, 4001) {
				t.Fatalf("expected idle expiry close: %v", err)
			}
			break
		}
	}

}

func TestRegisterRequiresServerIdentity(t *testing.T) {
	game := newCommandHandlerTestGame(t)
	var last message.Message
	game.RegisterSender(func(_ string, m message.Message) { last = m })
	handler := NewClientCommandHandler(game, func(string) (model.EntityId, bool) { return model.EntityId{}, false })
	handler.HandleCommand("peer", command.Command{Type: "register", Data: map[string]any{"id": model.NewEntityId().String(), "name": "Intruder"}})
	if game.IsRegistered("peer") || last.Metadata.Type != "registrationFailed" {
		t.Fatal("registration did not fail closed")
	}
}
