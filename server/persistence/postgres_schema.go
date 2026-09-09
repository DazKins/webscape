package persistence

import (
	"context"
	"fmt"
	"webscape/server/snapshot"

	"github.com/jackc/pgx/v5"
)

func (p *Postgres) initializeSchema(ctx context.Context) error {
	tx, err := p.conn.Begin(ctx)
	if err != nil {
		return dbError("begin schema initialization", err)
	}
	defer tx.Rollback(ctx)
	// Different world keys can start concurrently against an empty database.
	if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock(hashtextextended('webscape:schema',0))"); err != nil {
		return dbError("lock schema", err)
	}
	if _, err := tx.Exec(ctx, `CREATE TABLE IF NOT EXISTS webscape_worlds (
 world_key text PRIMARY KEY,
 schema_version integer NOT NULL CHECK(schema_version>0),
 snapshot_version integer NOT NULL CHECK(snapshot_version>0),
 content_hash text NOT NULL,
 tick numeric(20,0) NOT NULL CHECK(tick>=0),
 saved_at timestamptz NOT NULL DEFAULT now()
 );
 CREATE TABLE IF NOT EXISTS webscape_components (
 world_key text NOT NULL REFERENCES webscape_worlds(world_key) ON DELETE CASCADE,
 entity_id uuid NOT NULL,
 component_id text NOT NULL,
 version integer NOT NULL CHECK(version>0),
 data jsonb NOT NULL,
 PRIMARY KEY(world_key,entity_id,component_id)
 )`); err != nil {
		return dbError("initialize schema", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return dbError("commit schema", err)
	}
	return nil
}

// migrateLegacyWorld imports this key's v1 blob in the caller's transaction. The
// original blob is retained, marked v2 so an old binary cannot load stale state.
func (p *Postgres) migrateLegacyWorld(ctx context.Context, tx pgx.Tx) (*snapshot.State, error) {
	var exists bool
	if err := tx.QueryRow(ctx, "SELECT to_regclass('webscape_snapshots') IS NOT NULL").Scan(&exists); err != nil {
		return nil, dbError("locate legacy schema", err)
	}
	if !exists {
		return nil, nil
	}
	var version int
	var data []byte
	err := tx.QueryRow(ctx, "SELECT schema_version,snapshot FROM webscape_snapshots WHERE world_key=$1", p.worldKey).Scan(&version, &data)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, dbError("read legacy snapshot", err)
	}
	if version != 1 {
		return nil, fmt.Errorf("unsupported legacy snapshot schema version %d", version)
	}
	state, err := snapshot.Decode(data)
	if err != nil {
		return nil, err
	}
	if err := p.writeCheckpoint(ctx, tx, state, nil); err != nil {
		return nil, err
	}
	if _, err := tx.Exec(ctx, "UPDATE webscape_snapshots SET schema_version=2 WHERE world_key=$1", p.worldKey); err != nil {
		return nil, dbError("mark migrated snapshot", err)
	}
	return &state, nil
}
