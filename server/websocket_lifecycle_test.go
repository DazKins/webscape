package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
	"webscape/server/auth"
	"webscape/server/command"
	"webscape/server/config"
	"webscape/server/game"
	"webscape/server/game/model"
	"webscape/server/message"

	"github.com/gorilla/websocket"
)

type lifecycleFixture struct {
	t        *testing.T
	ws       *wsServer
	game     *game.Game
	app      *httptest.Server
	sessions map[string]*auth.Session
	leaves   chan string
}

func newLifecycleFixture(t *testing.T, settings config.ConnectionConfig) *lifecycleFixture {
	t.Helper()
	f := &lifecycleFixture{t: t, ws: NewWsServer(settings), game: newCommandHandlerTestGame(t), leaves: make(chan string, 32)}
	alice := model.NewEntityId()
	f.sessions = map[string]*auth.Session{
		"home": {PlayerID: alice, Expires: time.Now().Add(time.Hour), Done: make(chan struct{})},
		"away": {PlayerID: alice, Expires: time.Now().Add(time.Hour), Done: make(chan struct{})},
		"bob":  {PlayerID: model.NewEntityId(), Expires: time.Now().Add(time.Hour), Done: make(chan struct{})},
	}
	f.game.RetainOfflinePlayers()
	f.game.RegisterSender(f.ws.SendToClient)
	f.ws.SetConnectHandler(f.game.HandleConnect)
	f.ws.SetDisconnectHandler(func(id string) { f.game.HandleLeave(id); f.leaves <- id })
	handler := NewClientCommandHandler(f.game, f.ws.PlayerID)
	handler.onActivity = f.ws.RecordActivity
	handler.onTakeover = f.ws.ReplacePlayer
	f.ws.SetIncomingMessageHandler(func(id, raw string) {
		cmd, err := command.Unmarshal(raw)
		if err == nil {
			handler.HandleCommand(id, cmd)
		}
	})
	f.app = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Test-only authentication fixture. Production uses RequireSocket.
		f.ws.HandleWebSocket(w, r, f.sessions[r.URL.Query().Get("session")])
	}))
	t.Cleanup(func() { f.ws.Close(); f.app.Close() })
	return f
}

func (f *lifecycleFixture) connect(session string) (*websocket.Conn, *client) {
	f.t.Helper()
	conn, _, err := websocket.DefaultDialer.Dial(strings.Replace(f.app.URL, "http", "ws", 1)+"?session="+session, nil)
	if err != nil {
		f.t.Fatal(err)
	}
	f.t.Cleanup(func() { conn.Close() })
	readLifecycleMessage(f.t, conn, "world")
	f.ws.mutex.Lock()
	defer f.ws.mutex.Unlock()
	for _, c := range f.ws.clients {
		if c.session == f.sessions[session] {
			return conn, c
		}
	}
	f.t.Fatal("connection missing")
	return nil, nil
}

func readLifecycleMessage(t *testing.T, conn *websocket.Conn, kind string) map[string]any {
	t.Helper()
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	for {
		var msg struct {
			Metadata struct{ Type string }
			Data     map[string]any
		}
		if err := conn.ReadJSON(&msg); err != nil {
			t.Fatal(err)
		}
		if msg.Metadata.Type == kind {
			return msg.Data
		}
	}
}

func readLifecycleClose(t *testing.T, conn *websocket.Conn, code int) {
	t.Helper()
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			if !websocket.IsCloseError(err, code) {
				t.Fatalf("close = %v, want %d", err, code)
			}
			return
		}
	}
}

func (f *lifecycleFixture) register(conn *websocket.Conn) map[string]any {
	f.t.Helper()
	sendTestCommand(f.t, conn, "register", map[string]any{"name": "Alice"})
	data := readLifecycleMessage(f.t, conn, "registered")
	readLifecycleMessage(f.t, conn, "inactivity")
	return data
}

func (f *lifecycleFixture) age(c *client, age time.Duration) {
	f.ws.actions.Lock()
	defer f.ws.actions.Unlock()
	c.lastActivity = time.Now().Add(-age)
	f.ws.expireIdle(c, time.Now())
}

func TestWebSocketIdleWarningActivityAndRejoin(t *testing.T) {
	f := newLifecycleFixture(t, config.DefaultConnections())
	conn, c := f.connect("home")
	initial := f.register(conn)
	f.age(c, 14*time.Minute)
	warning := readLifecycleMessage(t, conn, "inactivity")
	if warning["warning"] != true || warning["remainingMs"].(float64) > 60000 {
		t.Fatalf("warning: %v", warning)
	}
	sendTestCommand(t, conn, "activity", map[string]any{})
	reset := readLifecycleMessage(t, conn, "inactivity")
	if reset["warning"] != false || reset["remainingMs"] != float64(900000) {
		t.Fatalf("reset: %v", reset)
	}
	// A gameplay intent also dismisses the warning.
	f.age(c, 14*time.Minute)
	readLifecycleMessage(t, conn, "inactivity")
	sendTestCommand(t, conn, "move", map[string]any{"x": 0, "y": 0})
	if readLifecycleMessage(t, conn, "inactivity")["warning"] != false {
		t.Fatal("gameplay did not count as activity")
	}
	f.age(c, 15*time.Minute)
	readLifecycleClose(t, conn, message.CloseInactive)
	select {
	case id := <-f.leaves:
		if id != c.id {
			t.Fatal("wrong cleanup")
		}
	case <-time.After(time.Second):
		t.Fatal("no cleanup")
	}
	if f.game.IsRegistered(c.id) || !c.session.Active() {
		t.Fatal("idle cleanup removed authentication or left player online")
	}
	next, _ := f.connect("away")
	// The original character name survives, even when a different guest name is supplied.
	sendTestCommand(t, next, "register", map[string]any{"name": "Changed"})
	restored := readLifecycleMessage(t, next, "registered")
	if restored["entityId"] != initial["entityId"] || restored["name"] != initial["name"] {
		t.Fatalf("character was not restored: %v", restored)
	}
}

func TestWebSocketTakeoverIsExplicitAuthenticatedAndCleansUpOnce(t *testing.T) {
	f := newLifecycleFixture(t, config.DefaultConnections())
	home, old := f.connect("home")
	f.register(home)
	away, current := f.connect("away")
	sendTestCommand(t, away, "register", map[string]any{"name": "Alice"})
	if readLifecycleMessage(t, away, "registrationFailed")["code"] != "playerAlreadyActive" {
		t.Fatal("missing actionable conflict")
	}
	for _, data := range []map[string]any{
		{"name": "Alice", "id": old.session.PlayerID.String(), "takeover": true},
		{"name": "Alice", "takeover": "yes"},
		{"name": "", "takeover": true},
	} {
		sendTestCommand(t, away, "register", data)
		readLifecycleMessage(t, away, "registrationFailed")
		if !f.game.IsRegistered(old.id) {
			t.Fatal("invalid takeover evicted player")
		}
	}
	bob, other := f.connect("bob")
	sendTestCommand(t, bob, "register", map[string]any{"name": "Bob", "takeover": true})
	readLifecycleMessage(t, bob, "registered")
	if !f.game.IsRegistered(old.id) {
		t.Fatal("another account evicted Alice")
	}
	sendTestCommand(t, away, "register", map[string]any{"name": "Alice", "takeover": true})
	readLifecycleMessage(t, away, "registered")
	readLifecycleClose(t, home, message.CloseReplaced)
	if f.game.IsRegistered(old.id) || !f.game.IsRegistered(current.id) || !f.game.IsRegistered(other.id) {
		t.Fatal("incorrect ownership after takeover")
	}
	// Simulate stale queued intent and a duplicate transport cleanup explicitly.
	f.ws.handleMessage(old.id, `{"type":"register","data":{"name":"Alice","takeover":true}}`)
	f.ws.actions.Lock()
	f.ws.remove(old)
	f.ws.actions.Unlock()
	if !f.game.IsRegistered(current.id) {
		t.Fatal("old socket removed replacement")
	}
	select {
	case id := <-f.leaves:
		if id != old.id {
			t.Fatal("wrong cleanup")
		}
	case <-time.After(time.Second):
		t.Fatal("missing cleanup")
	}
	select {
	case <-f.leaves:
		t.Fatal("cleanup ran more than once")
	default:
	}
	if !old.session.Active() {
		t.Fatal("takeover revoked account sign-in")
	}
}

func TestWebSocketHeartbeatDetectsMissingPongWithoutCountingAsActivity(t *testing.T) {
	settings := config.DefaultConnections()
	settings.PingIntervalSeconds = 1
	settings.PongTimeoutSeconds = 2
	f := newLifecycleFixture(t, settings)
	dead, old := f.connect("home")
	f.register(dead)
	dead.SetPingHandler(func(string) error { return nil })
	live, current := f.connect("bob")
	f.register(live)
	f.ws.actions.Lock()
	lastActivity := current.lastActivity
	f.ws.actions.Unlock()
	var pings atomic.Int32
	live.SetPingHandler(func(data string) error {
		pings.Add(1)
		return live.WriteControl(websocket.PongMessage, []byte(data), time.Now().Add(time.Second))
	})
	done := make(chan error, 1)
	go func() {
		for {
			if _, _, err := live.ReadMessage(); err != nil {
				done <- err
				return
			}
		}
	}()
	readLifecycleClose(t, dead, websocket.CloseAbnormalClosure)
	select {
	case id := <-f.leaves:
		if id != old.id {
			t.Fatal("wrong heartbeat cleanup")
		}
	case <-time.After(time.Second):
		t.Fatal("dead player remained registered")
	}
	// Receive several heartbeats beyond the original read deadline.
	deadline := time.After(3 * time.Second)
	for pings.Load() < 3 {
		select {
		case err := <-done:
			t.Fatalf("healthy socket closed: %v", err)
		case <-deadline:
			t.Fatal("pings not received")
		case <-time.After(20 * time.Millisecond):
		}
	}
	f.ws.actions.Lock()
	unchanged := current.lastActivity.Equal(lastActivity)
	f.ws.actions.Unlock()
	if !unchanged || !f.game.IsRegistered(current.id) {
		t.Fatal("pongs altered player activity or healthy player was removed")
	}
	// Heartbeats keep the transport alive, but still cannot defeat AFK expiry.
	f.age(current, 15*time.Minute)
	select {
	case err := <-done:
		if !websocket.IsCloseError(err, message.CloseInactive) {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("healthy AFK socket not closed")
	}
}

func TestWebSocketActivityRequiresRegistrationAndKnownIntent(t *testing.T) {
	f := newLifecycleFixture(t, config.DefaultConnections())
	conn, c := f.connect("home")
	for _, raw := range []string{`{"type":"activity","data":{}}`, `{"type":"move","data":{"x":0,"y":0}}`} {
		f.ws.handleMessage(c.id, raw)
	}
	f.ws.actions.Lock()
	marked := c.registered
	f.ws.actions.Unlock()
	if marked {
		t.Fatal("unregistered input started player session")
	}
	f.register(conn)
	f.ws.actions.Lock()
	initial := c.lastActivity
	f.ws.actions.Unlock()
	for _, raw := range []string{`not json`, `{"type":"ping","data":{}}`, `{"type":"activity","data":{"fake":true}}`} {
		f.ws.handleMessage(c.id, raw)
	}
	f.ws.actions.Lock()
	unchanged := c.lastActivity.Equal(initial)
	f.ws.actions.Unlock()
	if !unchanged {
		t.Fatal("invalid session intent extended AFK deadline")
	}
	// An already expired connection cannot extend itself before the next timer tick.
	f.ws.actions.Lock()
	c.lastActivity = time.Now().Add(-16 * time.Minute)
	f.ws.actions.Unlock()
	payload, _ := json.Marshal(command.Command{Type: "activity", Data: map[string]any{}})
	f.ws.handleMessage(c.id, string(payload))
	readLifecycleClose(t, conn, message.CloseInactive)
	if f.game.IsRegistered(c.id) {
		t.Fatal("late activity resurrected expired player")
	}
}
