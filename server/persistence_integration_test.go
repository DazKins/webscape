package server

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"
	"testing/fstest"
	"time"
	"webscape/server/auth"
	"webscape/server/config"
	"webscape/server/game/world"
	"webscape/server/internal/oidctest"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5"
)

type persistenceServerTest struct {
	t        *testing.T
	config   config.Config
	admin    *pgx.Conn
	provider *oidctest.Provider
}

func newPersistenceServerTest(t *testing.T) *persistenceServerTest {
	t.Helper()
	dsn := os.Getenv("WEBSCAPE_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set WEBSCAPE_TEST_DATABASE_URL for runtime persistence tests")
	}
	cfg, err := config.LoadFromFile("../config.json")
	if err != nil {
		t.Fatal(err)
	}
	cfg.Persistence.Driver = "postgres"
	cfg.Persistence.WorldKey = "server-test-" + uuid.NewString()
	cfg.Persistence.Postgres.ConnectionStringEnv = "WEBSCAPE_TEST_DATABASE_URL"
	admin, err := pgx.Connect(context.Background(), dsn)
	if err != nil {
		t.Fatal(err)
	}
	f := &persistenceServerTest{t: t, config: cfg, admin: admin, provider: oidctest.New(t)}
	t.Cleanup(func() {
		admin.Exec(context.Background(), "DELETE FROM webscape_worlds WHERE world_key=$1", cfg.Persistence.WorldKey)
		admin.Close(context.Background())
	})
	return f
}
func (f *persistenceServerTest) start(interval time.Duration) (string, func()) {
	t := f.t
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := listener.Addr().String()
	listener.Close()
	w, err := world.LoadFromGameFolder("../game-project")
	if err != nil {
		t.Fatal(err)
	}
	authConfig := f.provider.Config("http://" + addr)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- Start(ctx, fstest.MapFS{"index.html": {Data: []byte("test")}}, w, addr, 1, interval, true, f.config.Persistence, authConfig)
	}()
	var once sync.Once
	stop := func() {
		once.Do(func() {
			cancel()
			select {
			case err := <-done:
				if err != nil {
					t.Errorf("shutdown: %v", err)
				}
			case <-time.After(15 * time.Second):
				t.Error("server did not stop")
			}
		})
	}
	t.Cleanup(stop)
	return "ws://" + addr + "/ws", stop
}
func (f *persistenceServerTest) connect(address string) *websocket.Conn {
	t := f.t
	t.Helper()
	var ws *websocket.Conn
	var err error
	for deadline := time.Now().Add(10 * time.Second); time.Now().Before(deadline); {
		resp, getErr := http.Get(strings.Replace(address, "ws://", "http://", 1))
		if getErr != nil {
			time.Sleep(20 * time.Millisecond)
			continue
		}
		resp.Body.Close()
		browser := oidctest.Browser()
		app := strings.TrimSuffix(strings.Replace(address, "ws://", "http://", 1), "/ws")
		oidctest.Login(t, browser, app, "restart")
		req, _ := http.NewRequest("GET", app, nil)
		for _, cookie := range browser.Jar.Cookies(req.URL) {
			req.AddCookie(cookie)
		}
		req.Header.Set("Origin", app)
		ws, _, err = websocket.DefaultDialer.Dial(address, req.Header)
		if err == nil {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { ws.Close() })
	return ws
}
func sendTestCommand(t *testing.T, ws *websocket.Conn, kind string, data any) {
	t.Helper()
	if err := ws.WriteJSON(map[string]any{"type": kind, "data": data}); err != nil {
		t.Fatal(err)
	}
}
func registerTestPlayer(t *testing.T, ws *websocket.Conn, id string) {
	sendTestCommand(t, ws, "register", map[string]any{"name": "Restart Player"})
}

type wireItem struct {
	ID       string `json:"id"`
	Quantity int    `json:"quantity"`
}

func testInventory(t *testing.T, ws *websocket.Conn, id string) []wireItem {
	t.Helper()
	ws.SetReadDeadline(time.Now().Add(10 * time.Second))
	for {
		var msg struct {
			Metadata struct {
				Type string `json:"type"`
			} `json:"metadata"`
			Data struct {
				Entities []struct {
					EntityID    string `json:"entityId"`
					ComponentID string `json:"componentId"`
					Data        struct {
						Items []wireItem `json:"items"`
					} `json:"data"`
				} `json:"entities"`
			} `json:"data"`
		}
		if err := ws.ReadJSON(&msg); err != nil {
			t.Fatal(err)
		}
		if msg.Metadata.Type == "registrationFailed" {
			t.Fatal("returning player rejected")
		}
		for _, e := range msg.Data.Entities {
			if e.EntityID == id && e.ComponentID == "inventory" {
				return e.Data.Items
			}
		}
	}
}
func (f *persistenceServerTest) savedAt() time.Time {
	f.t.Helper()
	var saved time.Time
	if err := f.admin.QueryRow(context.Background(), "SELECT saved_at FROM webscape_worlds WHERE world_key=$1", f.config.Persistence.WorldKey).Scan(&saved); err != nil {
		f.t.Fatal(err)
	}
	return saved
}
func (f *persistenceServerTest) waitForDisconnectSave(previous time.Time) {
	f.t.Helper()
	for deadline := time.Now().Add(5 * time.Second); time.Now().Before(deadline); {
		if f.savedAt().After(previous) {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	f.t.Fatal("disconnect did not checkpoint before next tick")
}

func TestServerRestartsWithPostgresProgress(t *testing.T) {
	f := newPersistenceServerTest(t)
	address, stop := f.start(100 * time.Millisecond)
	id := auth.PlayerID(f.provider.URL, "restart").String()
	ws := f.connect(address)
	registerTestPlayer(t, ws, id)
	original := testInventory(t, ws, id)
	if len(original) < 2 {
		t.Fatal("missing starter items")
	}
	sendTestCommand(t, ws, "drop", map[string]any{"itemId": original[0].ID})
	after := testInventory(t, ws, id)
	if len(after) != len(original)-1 {
		t.Fatal("drop did not update inventory")
	}
	stop() // Player is still connected: final save must include active players.
	address, stop = f.start(100 * time.Millisecond)
	ws = f.connect(address)
	registerTestPlayer(t, ws, id)
	if restored := testInventory(t, ws, id); !reflect.DeepEqual(restored, after) {
		t.Fatal("restart changed inventory IDs/quantities")
	}
	var drops int
	if err := f.admin.QueryRow(context.Background(), "SELECT count(*) FROM webscape_components WHERE world_key=$1 AND component_id='droppeditem'", f.config.Persistence.WorldKey).Scan(&drops); err != nil || drops != 1 {
		t.Fatalf("ground drops=%d: %v", drops, err)
	}
	stop()
}

func TestDisconnectedPlayerReconnectsBeforeAndAfterRestart(t *testing.T) {
	f := newPersistenceServerTest(t)
	// A long tick proves disconnect itself flushes accepted mutations, without
	// relying on either a periodic checkpoint or a graceful shutdown save.
	address, stop := f.start(time.Minute)
	id := auth.PlayerID(f.provider.URL, "restart").String()
	ws := f.connect(address)
	registerTestPlayer(t, ws, id)
	original := testInventory(t, ws, id)
	before := f.savedAt()
	sendTestCommand(t, ws, "drop", map[string]any{"itemId": original[0].ID})
	ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	ws.Close()
	f.waitForDisconnectSave(before)
	var count int
	if err := f.admin.QueryRow(context.Background(), "SELECT jsonb_array_length(data->'items') FROM webscape_components WHERE world_key=$1 AND entity_id=$2::uuid AND component_id='inventory'", f.config.Persistence.WorldKey, id).Scan(&count); err != nil || count != len(original)-1 {
		t.Fatalf("disconnect lost accepted command: %d %v", count, err)
	}
	ws = f.connect(address)
	registerTestPlayer(t, ws, id)
	expected := original[1:]
	if restored := testInventory(t, ws, id); !reflect.DeepEqual(restored, expected) {
		t.Fatal("disconnect/reconnect reset inventory")
	}
	before = f.savedAt()
	ws.Close()
	f.waitForDisconnectSave(before)
	stop() // Already offline when the server stops.
	address, stop = f.start(time.Minute)
	ws = f.connect(address)
	registerTestPlayer(t, ws, id)
	if restored := testInventory(t, ws, id); !reflect.DeepEqual(restored, expected) {
		t.Fatal("offline/restart/reconnect reset inventory")
	}
	stop()
}

func TestRejectedCommandsCannotTriggerCheckpoints(t *testing.T) {
	f := newPersistenceServerTest(t)
	address, stop := f.start(time.Minute)
	ws := f.connect(address)
	before := f.savedAt()
	for range 30 {
		if err := ws.WriteMessage(websocket.TextMessage, []byte(`{}`)); err != nil {
			t.Fatal(err)
		}
		sendTestCommand(t, ws, "unknown", map[string]any{})
		sendTestCommand(t, ws, "register", map[string]any{"id": "invalid"})
	}
	// A successful registration reply acts as a barrier for all previous frames.
	id := auth.PlayerID(f.provider.URL, "restart").String()
	registerTestPlayer(t, ws, id)
	ws.SetReadDeadline(time.Now().Add(10 * time.Second))
	for {
		var msg struct {
			Metadata struct {
				Type string `json:"type"`
			} `json:"metadata"`
			Data json.RawMessage `json:"data"`
		}
		if err := ws.ReadJSON(&msg); err != nil {
			t.Fatal(err)
		}
		if msg.Metadata.Type == "registered" {
			break
		}
	}
	if !f.savedAt().Equal(before) {
		t.Fatal("peer commands triggered a checkpoint before the tick")
	}
	// Disconnect still saves the newly registered player.
	ws.Close()
	f.waitForDisconnectSave(before)
	stop()
}
