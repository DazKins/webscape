package server

import (
	"context"
	"net"
	"os"
	"sync"
	"testing"
	"testing/fstest"
	"time"
	"webscape/server/config"
	"webscape/server/game/world"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5"
)

func TestServerRestartsWithPostgresProgress(t *testing.T) {
	dsn := os.Getenv("WEBSCAPE_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set WEBSCAPE_TEST_DATABASE_URL for server restart integration test")
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
	defer admin.Close(context.Background())
	defer admin.Exec(context.Background(), "DELETE FROM webscape_snapshots WHERE world_key=$1", cfg.Persistence.WorldKey)
	start := func() (string, func()) {
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
		ctx, cancel := context.WithCancel(context.Background())
		done := make(chan error, 1)
		go func() {
			done <- Start(ctx, fstest.MapFS{"index.html": {Data: []byte("test")}}, w, addr, 1, 100*time.Millisecond, false, cfg.Persistence)
		}()
		var once sync.Once
		stop := func() {
			once.Do(func() {
				cancel()
				select {
				case err := <-done:
					if err != nil {
						t.Errorf("server shutdown: %v", err)
					}
				case <-time.After(15 * time.Second):
					t.Error("server did not stop")
				}
			})
		}
		t.Cleanup(stop)
		return "ws://" + addr + "/ws", stop
	}
	connect := func(address, id string) *websocket.Conn {
		t.Helper()
		var ws *websocket.Conn
		var err error
		for deadline := time.Now().Add(10 * time.Second); time.Now().Before(deadline); {
			ws, _, err = websocket.DefaultDialer.Dial(address, nil)
			if err == nil {
				break
			}
			time.Sleep(20 * time.Millisecond)
		}
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { ws.Close() })
		if err := ws.WriteJSON(map[string]any{"type": "register", "data": map[string]any{"id": id, "name": "Restart Player"}}); err != nil {
			t.Fatal(err)
		}
		return ws
	}
	type wireItem struct {
		ID       string `json:"id"`
		Quantity int    `json:"quantity"`
	}
	inventory := func(ws *websocket.Conn, id string) []wireItem {
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
			for _, entity := range msg.Data.Entities {
				if entity.EntityID == id && entity.ComponentID == "inventory" {
					return entity.Data.Items
				}
			}
		}
	}
	address, stop := start()
	id := uuid.NewString()
	ws := connect(address, id)
	original := inventory(ws, id)
	if len(original) < 2 {
		t.Fatal("missing starter items")
	}
	dropped := original[0].ID
	if err := ws.WriteJSON(map[string]any{"type": "drop", "data": map[string]any{"itemId": dropped}}); err != nil {
		t.Fatal(err)
	}
	after := inventory(ws, id)
	if len(after) != len(original)-1 {
		t.Fatal("drop did not update inventory")
	}
	// Stop with the player connected: shutdown must snapshot the active player and
	// explicitly close the upgraded WebSocket before releasing the database owner.
	stop()
	address, stop = start()
	ws = connect(address, id)
	restored := inventory(ws, id)
	if len(restored) != len(after) {
		t.Fatal("restart reset inventory")
	}
	for i, item := range restored {
		if item != after[i] {
			t.Fatal("restart changed item identity or quantity")
		}
	}
	var drops int
	err = admin.QueryRow(context.Background(), `SELECT count(*) FROM webscape_snapshots,
 jsonb_each(snapshot->'entities') AS entity
 WHERE world_key=$1 AND entity.value ? 'droppeditem'`, cfg.Persistence.WorldKey).Scan(&drops)
	if err != nil || drops != 1 {
		t.Fatalf("saved ground drops=%d, err=%v", drops, err)
	}
	stop()
}
