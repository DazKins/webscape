package persistence

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// Run against an isolated test database using WEBSCAPE_TEST_DATABASE_URL.
func TestPostgresAtomicSaveRestartAndExclusiveOwner(t *testing.T) {
	dsn := os.Getenv("WEBSCAPE_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set WEBSCAPE_TEST_DATABASE_URL for PostgreSQL integration tests")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	key := "test-" + uuid.NewString()
	store, err := OpenPostgres(ctx, dsn, key)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close(ctx)
	admin, err := pgx.Connect(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close(ctx)
	defer admin.Exec(ctx, "DELETE FROM webscape_snapshots WHERE world_key=$1", key)
	if data, err := store.Load(ctx); err != nil || data != nil {
		t.Fatalf("fresh world: %s %v", data, err)
	}
	if other, err := OpenPostgres(ctx, dsn, key); err == nil {
		other.Close(ctx)
		t.Fatal("second owner admitted")
	}
	first := []byte(`{"version":1,"entities":{"player":{"gold":7},"drop":{"gold":3}}}`)
	if err := store.Save(ctx, first); err != nil {
		t.Fatal(err)
	}
	if err := store.Save(ctx, []byte("broken")); err == nil {
		t.Fatal("accepted invalid JSON")
	}
	// A canceled transaction cannot destroy the committed record.
	canceled, stop := context.WithCancel(ctx)
	stop()
	if err := store.Save(canceled, []byte(`{"version":1,"entities":{}}`)); err == nil {
		t.Fatal("canceled write succeeded")
	}
	store.Close(ctx)
	store, err = OpenPostgres(ctx, dsn, key)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close(ctx)
	data, err := store.Load(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	json.Unmarshal(data, &got)
	if len(got["entities"].(map[string]any)) != 2 {
		t.Fatal("failed write replaced committed entities")
	}
	if err := store.Save(ctx, []byte(`{"version":1,"entities":{"player":{"gold":10}}}`)); err != nil {
		t.Fatal(err)
	}
	store.Close(ctx)
	store, err = OpenPostgres(ctx, dsn, key)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close(ctx)
	data, err = store.Load(ctx)
	if err != nil {
		t.Fatal(err)
	}
	json.Unmarshal(data, &got)
	if len(got["entities"].(map[string]any)) != 1 {
		t.Fatal("deleted drop resurrected")
	}
	if _, err := admin.Exec(ctx, "UPDATE webscape_snapshots SET schema_version=99 WHERE world_key=$1", key); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Load(ctx); err == nil {
		t.Fatal("unknown schema version accepted")
	}
	if _, err := admin.Exec(ctx, "SELECT pg_terminate_backend($1)", store.conn.PgConn().PID()); err != nil {
		t.Fatal(err)
	}
	if err := store.Save(ctx, first); err == nil {
		t.Fatal("lost connection silently resumed writes without its ownership lock")
	}

}

func TestConnectionErrorsDoNotExposeCredentials(t *testing.T) {
	_, err := OpenPostgres(context.Background(), "postgres://user:super-secret@host:invalid/db", "test")
	if err == nil || strings.Contains(err.Error(), "super-secret") {
		t.Fatalf("unsafe parse error: %v", err)
	}
}
