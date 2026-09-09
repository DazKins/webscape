package persistence

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"
	"webscape/server/snapshot"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func postgresFixture(t *testing.T) (context.Context, *Postgres, *pgx.Conn, string) {
	t.Helper()
	dsn := os.Getenv("WEBSCAPE_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set WEBSCAPE_TEST_DATABASE_URL for PostgreSQL integration tests")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)
	key := "test-" + uuid.NewString()
	p, err := OpenPostgres(ctx, dsn, key)
	if err != nil {
		t.Fatal(err)
	}
	admin, err := pgx.Connect(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		p.Close(ctx)
		admin.Exec(ctx, "DELETE FROM webscape_worlds WHERE world_key=$1", key)
		admin.Close(ctx)
	})
	return ctx, p, admin, dsn
}
func encodeState(t *testing.T, s snapshot.State) []byte {
	t.Helper()
	data, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	return data
}
func testState() snapshot.State {
	return snapshot.State{Version: 1, ContentHash: "test", Tick: 1, Entities: map[string]map[string]snapshot.Component{
		"11111111-1111-4111-8111-111111111111": {"inventory": {Version: 1, Data: json.RawMessage(`{"gold":7}`)}, "position": {Version: 1, Data: json.RawMessage(`{"x":1,"y":2}`)}},
		"22222222-2222-4222-8222-222222222222": {"droppeditem": {Version: 1, Data: json.RawMessage(`{"gold":3}`)}},
	}}
}
func TestPostgresAtomicSaveRestartAndExclusiveOwner(t *testing.T) {
	ctx, p, admin, dsn := postgresFixture(t)
	if data, err := p.Load(ctx); err != nil || data != nil {
		t.Fatalf("fresh world: %s %v", data, err)
	}
	if other, err := OpenPostgres(ctx, dsn, p.worldKey); err == nil {
		other.Close(ctx)
		t.Fatal("second owner admitted")
	}
	first := testState()
	if err := p.Save(ctx, encodeState(t, first)); err != nil {
		t.Fatal(err)
	}
	// Force a server-side failure after metadata/upserts are queued. Both the item
	// transfer and removal must roll back; the cached baseline must also stay old.
	_, err := admin.Exec(ctx, `CREATE OR REPLACE FUNCTION webscape_test_reject_delete() RETURNS trigger LANGUAGE plpgsql AS $$
 BEGIN IF OLD.world_key LIKE 'test-%' THEN RAISE EXCEPTION 'test deletion failure'; END IF; RETURN OLD; END $$;
 CREATE TRIGGER webscape_test_reject_delete BEFORE DELETE ON webscape_components FOR EACH ROW EXECUTE FUNCTION webscape_test_reject_delete()`)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		admin.Exec(ctx, "DROP TRIGGER IF EXISTS webscape_test_reject_delete ON webscape_components; DROP FUNCTION IF EXISTS webscape_test_reject_delete()")
	})
	changed := testState()
	changed.Tick = 2
	changed.Entities["11111111-1111-4111-8111-111111111111"]["inventory"] = snapshot.Component{Version: 1, Data: json.RawMessage(`{"gold":10}`)}
	delete(changed.Entities, "22222222-2222-4222-8222-222222222222")
	if err := p.Save(ctx, encodeState(t, changed)); err == nil {
		t.Fatal("expected mid-transaction failure")
	}
	var gold, tick int
	if err := admin.QueryRow(ctx, "SELECT (data->>'gold')::int FROM webscape_components WHERE world_key=$1 AND component_id='inventory'", p.worldKey).Scan(&gold); err != nil {
		t.Fatal(err)
	}
	if err := admin.QueryRow(ctx, "SELECT tick::integer FROM webscape_worlds WHERE world_key=$1", p.worldKey).Scan(&tick); err != nil {
		t.Fatal(err)
	}
	if gold != 7 || tick != 1 {
		t.Fatal("partial checkpoint committed")
	}
	if _, err := admin.Exec(ctx, "DROP TRIGGER webscape_test_reject_delete ON webscape_components; DROP FUNCTION webscape_test_reject_delete()"); err != nil {
		t.Fatal(err)
	}
	// Retrying the same checkpoint must still write the failed delta.
	if err := p.Save(ctx, encodeState(t, changed)); err != nil {
		t.Fatal(err)
	}
	p.Close(ctx)
	reopened, err := OpenPostgres(ctx, dsn, p.worldKey)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close(ctx)
	data, err := reopened.Load(ctx)
	if err != nil {
		t.Fatal(err)
	}
	got, err := snapshot.Decode(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Entities) != 1 || got.Tick != 2 {
		t.Fatal("deleted entity or tick not restored")
	}
	if !sameComponent(got.Entities["11111111-1111-4111-8111-111111111111"]["inventory"], changed.Entities["11111111-1111-4111-8111-111111111111"]["inventory"]) {
		t.Fatal("inventory transfer not restored")
	}
	if _, err := admin.Exec(ctx, "UPDATE webscape_worlds SET schema_version=99 WHERE world_key=$1", p.worldKey); err != nil {
		t.Fatal(err)
	}
	if _, err := reopened.Load(ctx); err == nil {
		t.Fatal("unknown storage version accepted")
	}
	if _, err := admin.Exec(ctx, "SELECT pg_terminate_backend($1)", reopened.conn.PgConn().PID()); err != nil {
		t.Fatal(err)
	}
	if err := reopened.Save(ctx, encodeState(t, changed)); err == nil {
		t.Fatal("lost connection resumed without ownership lock")
	}
}

func TestPostgresDoesNotRewriteUnchangedComponents(t *testing.T) {
	ctx, p, admin, dsn := postgresFixture(t)
	first := testState()
	if err := p.Save(ctx, encodeState(t, first)); err != nil {
		t.Fatal(err)
	}
	rowVersion := func() string {
		t.Helper()
		var v string
		if err := admin.QueryRow(ctx, "SELECT xmin::text FROM webscape_components WHERE world_key=$1 AND component_id='position'", p.worldKey).Scan(&v); err != nil {
			t.Fatal(err)
		}
		return v
	}
	before := rowVersion()
	p.Close(ctx)
	p, err := OpenPostgres(ctx, dsn, p.worldKey)
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close(ctx)
	if _, err := p.Load(ctx); err != nil {
		t.Fatal(err)
	}
	first.Tick++
	// Reordered JSON from jsonb must compare equal to the original codec payload.
	first.Entities["11111111-1111-4111-8111-111111111111"]["position"] = snapshot.Component{Version: 1, Data: json.RawMessage(`{"y":2,"x":1}`)}
	if err := p.Save(ctx, encodeState(t, first)); err != nil {
		t.Fatal(err)
	}
	if rowVersion() != before {
		t.Fatal("unchanged component was rewritten")
	}
	first.Entities["11111111-1111-4111-8111-111111111111"]["position"] = snapshot.Component{Version: 1, Data: json.RawMessage(`{"y":2,"x":3}`)}
	if err := p.Save(ctx, encodeState(t, first)); err != nil {
		t.Fatal(err)
	}
	if rowVersion() == before {
		t.Fatal("changed component was not written")
	}
	if err := p.Save(ctx, []byte(`{"version":1,"entities":{"invalid":{"x":null}}}`)); err == nil {
		t.Fatal("invalid snapshot accepted")
	}
}

func TestPostgresMigratesLegacySnapshot(t *testing.T) {
	ctx, p, admin, _ := postgresFixture(t)
	if _, err := admin.Exec(ctx, `CREATE TABLE IF NOT EXISTS webscape_snapshots(world_key text PRIMARY KEY,schema_version integer NOT NULL,snapshot jsonb NOT NULL,saved_at timestamptz NOT NULL DEFAULT now())`); err != nil {
		t.Fatal(err)
	}
	first := testState()
	if _, err := admin.Exec(ctx, "INSERT INTO webscape_snapshots(world_key,schema_version,snapshot) VALUES($1,1,$2::jsonb)", p.worldKey, string(encodeState(t, first))); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { admin.Exec(ctx, "DELETE FROM webscape_snapshots WHERE world_key=$1", p.worldKey) })
	data, err := p.Load(ctx)
	if err != nil {
		t.Fatal(err)
	}
	state, err := snapshot.Decode(data)
	if err != nil || len(state.Entities) != 2 {
		t.Fatalf("migration lost entities: %v", err)
	}
	var rows, legacyVersion int
	if err := admin.QueryRow(ctx, "SELECT count(*) FROM webscape_components WHERE world_key=$1", p.worldKey).Scan(&rows); err != nil {
		t.Fatal(err)
	}
	if err := admin.QueryRow(ctx, "SELECT schema_version FROM webscape_snapshots WHERE world_key=$1", p.worldKey).Scan(&legacyVersion); err != nil {
		t.Fatal(err)
	}
	if rows != 3 || legacyVersion != 2 {
		t.Fatal("migration did not preserve components or fence old binaries")
	}
	// The backup cannot replace the current rows on a subsequent load.
	first.Tick = 9
	if err := p.Save(ctx, encodeState(t, first)); err != nil {
		t.Fatal(err)
	}
	data, err = p.Load(ctx)
	if err != nil {
		t.Fatal(err)
	}
	state, err = snapshot.Decode(data)
	if err != nil || state.Tick != 9 {
		t.Fatal("legacy backup replaced current state")
	}
}
func TestConnectionErrorsDoNotExposeCredentials(t *testing.T) {
	_, err := OpenPostgres(context.Background(), "postgres://user:super-secret@host:invalid/db", "test")
	if err == nil || strings.Contains(err.Error(), "super-secret") {
		t.Fatalf("unsafe parse error: %v", err)
	}
}
