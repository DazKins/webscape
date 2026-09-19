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
		admin.Exec(ctx, "DELETE FROM webscape_player_saves WHERE world_key=$1", key)
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

const firstPlayer = "11111111-1111-4111-8111-111111111111"
const secondPlayer = "22222222-2222-4222-8222-222222222222"

func testState() snapshot.State {
	return snapshot.State{Version: 2, Players: map[string]map[string]snapshot.Component{
		firstPlayer:  {"player": {Version: 1, Data: json.RawMessage(`{}`)}, "inventory": {Version: 1, Data: json.RawMessage(`{"gold":7}`)}},
		secondPlayer: {"player": {Version: 1, Data: json.RawMessage(`{}`)}, "inventory": {Version: 1, Data: json.RawMessage(`{"gold":3}`)}},
	}}
}
func TestPostgresAtomicSaveRestartAndExclusiveOwner(t *testing.T) {
	ctx, p, admin, dsn := postgresFixture(t)
	if data, err := p.Load(ctx); err != nil || data != nil {
		t.Fatalf("fresh save: %s %v", data, err)
	}
	if other, err := OpenPostgres(ctx, dsn, p.worldKey); err == nil {
		other.Close(ctx)
		t.Fatal("second owner admitted")
	}
	first := testState()
	if err := p.Save(ctx, encodeState(t, first)); err != nil {
		t.Fatal(err)
	}
	// A constraint failing on the second player must roll back all player updates.
	_, err := admin.Exec(ctx, `ALTER TABLE webscape_players ADD CONSTRAINT test_reject_gold CHECK ((components->'inventory'->'data'->>'gold')::int < 20) NOT VALID`)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { admin.Exec(ctx, "ALTER TABLE webscape_players DROP CONSTRAINT IF EXISTS test_reject_gold") })
	changed := testState()
	changed.Players[firstPlayer]["inventory"] = snapshot.Component{Version: 1, Data: json.RawMessage(`{"gold":10}`)}
	changed.Players[secondPlayer]["inventory"] = snapshot.Component{Version: 1, Data: json.RawMessage(`{"gold":20}`)}
	if err := p.Save(ctx, encodeState(t, changed)); err == nil {
		t.Fatal("expected transaction failure")
	}
	var gold int
	if err := admin.QueryRow(ctx, "SELECT (components->'inventory'->'data'->>'gold')::int FROM webscape_players WHERE world_key=$1 AND player_id=$2::uuid", p.worldKey, firstPlayer).Scan(&gold); err != nil {
		t.Fatal(err)
	}
	if gold != 7 {
		t.Fatal("partial checkpoint committed")
	}
	if _, err := admin.Exec(ctx, "ALTER TABLE webscape_players DROP CONSTRAINT test_reject_gold"); err != nil {
		t.Fatal(err)
	}
	if err := p.Save(ctx, encodeState(t, changed)); err != nil {
		t.Fatal(err)
	}
	delete(changed.Players, secondPlayer)
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
	if len(got.Players) != 2 || !sameComponent(got.Players[firstPlayer]["inventory"], changed.Players[firstPlayer]["inventory"]) {
		t.Fatal("lost saved or absent player")
	}
	if _, err := admin.Exec(ctx, "UPDATE webscape_player_saves SET version=99 WHERE world_key=$1", p.worldKey); err != nil {
		t.Fatal(err)
	}
	if _, err := reopened.Load(ctx); err == nil {
		t.Fatal("unknown storage version accepted")
	}
	if _, err := admin.Exec(ctx, "SELECT pg_terminate_backend($1)", reopened.conn.PgConn().PID()); err != nil {
		t.Fatal(err)
	}
	if err := reopened.Save(ctx, encodeState(t, changed)); err == nil {
		t.Fatal("lost ownership accepted")
	}
}
func TestPostgresDoesNotRewriteUnchangedPlayers(t *testing.T) {
	ctx, p, admin, dsn := postgresFixture(t)
	state := testState()
	state.Players[firstPlayer]["position"] = snapshot.Component{Version: 1, Data: json.RawMessage(`{"x":1,"y":2}`)}
	if err := p.Save(ctx, encodeState(t, state)); err != nil {
		t.Fatal(err)
	}
	rowVersion := func() string {
		t.Helper()
		var v string
		if err := admin.QueryRow(ctx, "SELECT xmin::text FROM webscape_players WHERE world_key=$1 AND player_id=$2::uuid", p.worldKey, firstPlayer).Scan(&v); err != nil {
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
	state.Players[firstPlayer]["position"] = snapshot.Component{Version: 1, Data: json.RawMessage(`{"y":2,"x":1}`)}
	if err := p.Save(ctx, encodeState(t, state)); err != nil {
		t.Fatal(err)
	}
	if rowVersion() != before {
		t.Fatal("unchanged player rewritten")
	}
	state.Players[firstPlayer]["position"] = snapshot.Component{Version: 1, Data: json.RawMessage(`{"y":2,"x":3}`)}
	if err := p.Save(ctx, encodeState(t, state)); err != nil {
		t.Fatal(err)
	}
	if rowVersion() == before {
		t.Fatal("changed player not written")
	}
}
func TestPostgresMigratesLegacySaves(t *testing.T) {
	for _, format := range []string{"blob", "components"} {
		t.Run(format, func(t *testing.T) {
			ctx, p, admin, _ := postgresFixture(t)
			state := testState()
			entities := state.Players
			entities["33333333-3333-4333-8333-333333333333"] = map[string]snapshot.Component{"obsolete": {Version: 99, Data: json.RawMessage(`{}`)}}
			if format == "blob" {
				_, err := admin.Exec(ctx, `CREATE TABLE IF NOT EXISTS webscape_snapshots(world_key text PRIMARY KEY,schema_version integer NOT NULL,snapshot jsonb NOT NULL,saved_at timestamptz NOT NULL DEFAULT now())`)
				if err != nil {
					t.Fatal(err)
				}
				data, _ := json.Marshal(map[string]any{"version": 1, "entities": entities, "tick": 999, "contentHash": "old"})
				if _, err := admin.Exec(ctx, "INSERT INTO webscape_snapshots(world_key,schema_version,snapshot) VALUES($1,1,$2::jsonb)", p.worldKey, string(data)); err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { admin.Exec(ctx, "DELETE FROM webscape_snapshots WHERE world_key=$1", p.worldKey) })
			} else {
				_, err := admin.Exec(ctx, `CREATE TABLE IF NOT EXISTS webscape_worlds(world_key text PRIMARY KEY,schema_version integer NOT NULL,snapshot_version integer NOT NULL,content_hash text NOT NULL,tick numeric(20,0) NOT NULL,saved_at timestamptz NOT NULL DEFAULT now());CREATE TABLE IF NOT EXISTS webscape_components(world_key text NOT NULL REFERENCES webscape_worlds(world_key) ON DELETE CASCADE,entity_id uuid NOT NULL,component_id text NOT NULL,version integer NOT NULL,data jsonb NOT NULL,PRIMARY KEY(world_key,entity_id,component_id))`)
				if err != nil {
					t.Fatal(err)
				}
				if _, err := admin.Exec(ctx, "INSERT INTO webscape_worlds(world_key,schema_version,snapshot_version,content_hash,tick) VALUES($1,2,1,'old',999)", p.worldKey); err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { admin.Exec(ctx, "DELETE FROM webscape_worlds WHERE world_key=$1", p.worldKey) })
				for id, cs := range entities {
					for key, c := range cs {
						if _, err := admin.Exec(ctx, "INSERT INTO webscape_components VALUES($1,$2::uuid,$3,$4,$5::jsonb)", p.worldKey, id, key, c.Version, string(c.Data)); err != nil {
							t.Fatal(err)
						}
					}
				}
			}
			data, err := p.Load(ctx)
			if err != nil {
				t.Fatal(err)
			}
			got, err := snapshot.Decode(data)
			if err != nil || len(got.Players) != 2 {
				t.Fatalf("migration lost players: %v", err)
			}
			var version int
			table := "webscape_worlds"
			if format == "blob" {
				table = "webscape_snapshots"
			}
			if err := admin.QueryRow(ctx, "SELECT schema_version FROM "+table+" WHERE world_key=$1", p.worldKey).Scan(&version); err != nil || version != 3 {
				t.Fatalf("legacy not fenced: %v", err)
			}
			got.Players[firstPlayer]["inventory"] = snapshot.Component{Version: 1, Data: json.RawMessage(`{"gold":99}`)}
			if err := p.Save(ctx, encodeState(t, got)); err != nil {
				t.Fatal(err)
			}
			data, err = p.Load(ctx)
			if err != nil {
				t.Fatal(err)
			}
			current, err := snapshot.Decode(data)
			if err != nil || !samePlayer(current.Players[firstPlayer], got.Players[firstPlayer]) {
				t.Fatal("legacy reapplied")
			}
		})
	}
}
func TestConnectionErrorsDoNotExposeCredentials(t *testing.T) {
	_, err := OpenPostgres(context.Background(), "postgres://user:super-secret@host:invalid/db", "test")
	if err == nil || strings.Contains(err.Error(), "super-secret") {
		t.Fatalf("unsafe parse error: %v", err)
	}
}
