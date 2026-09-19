package persistence

import (
	"context"
	"encoding/json"
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
	if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock(hashtextextended('webscape:schema',0))"); err != nil {
		return dbError("lock schema", err)
	}
	if _, err := tx.Exec(ctx, `CREATE TABLE IF NOT EXISTS webscape_player_saves (
 world_key text PRIMARY KEY, version integer NOT NULL CHECK(version>0), saved_at timestamptz NOT NULL DEFAULT now());
 CREATE TABLE IF NOT EXISTS webscape_players (
 world_key text NOT NULL REFERENCES webscape_player_saves(world_key) ON DELETE CASCADE,
 player_id uuid NOT NULL, components jsonb NOT NULL CHECK(jsonb_typeof(components)='object' AND components ? 'player'),
 updated_at timestamptz NOT NULL DEFAULT now(), PRIMARY KEY(world_key,player_id));`); err != nil {
		return dbError("initialize schema", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return dbError("commit schema", err)
	}
	return nil
}

// Import only player records. Legacy rows remain a backup and are fenced against
// old binaries. Each world migrates under its ownership lock in one transaction.
func (p *Postgres) migrateLegacyWorld(ctx context.Context, tx pgx.Tx) (*snapshot.State, error) {
	var worlds, blobs bool
	if err := tx.QueryRow(ctx, `SELECT to_regclass('webscape_worlds') IS NOT NULL,to_regclass('webscape_snapshots') IS NOT NULL`).Scan(&worlds, &blobs); err != nil {
		return nil, dbError("locate legacy tables", err)
	}
	state := snapshot.State{Version: 2, Players: map[string]map[string]snapshot.Component{}}
	if worlds {
		var version, snapshotVersion int
		err := tx.QueryRow(ctx, "SELECT schema_version,snapshot_version FROM webscape_worlds WHERE world_key=$1", p.worldKey).Scan(&version, &snapshotVersion)
		if err != nil && err != pgx.ErrNoRows {
			return nil, dbError("read legacy world", err)
		}
		if err == nil {
			if version != 2 || snapshotVersion != 1 {
				return nil, fmt.Errorf("unsupported legacy world schema version %d", version)
			}
			rows, err := tx.Query(ctx, `SELECT c.entity_id::text,c.component_id,c.version,c.data FROM webscape_components c
 WHERE c.world_key=$1 AND EXISTS(SELECT 1 FROM webscape_components p WHERE p.world_key=c.world_key AND p.entity_id=c.entity_id AND p.component_id='player')`, p.worldKey)
			if err != nil {
				return nil, dbError("read legacy players", err)
			}
			for rows.Next() {
				var id, key string
				var c snapshot.Component
				if err := rows.Scan(&id, &key, &c.Version, &c.Data); err != nil {
					rows.Close()
					return nil, dbError("decode legacy player", err)
				}
				if state.Players[id] == nil {
					state.Players[id] = map[string]snapshot.Component{}
				}
				state.Players[id][key] = c
			}
			rows.Close()
			if err := rows.Err(); err != nil {
				return nil, dbError("read legacy players", err)
			}
			data, err := json.Marshal(state)
			if err != nil {
				return nil, err
			}
			if _, err := snapshot.Decode(data); err != nil {
				return nil, err
			}
			if err := p.writeCheckpoint(ctx, tx, state, nil); err != nil {
				return nil, err
			}
			if _, err := tx.Exec(ctx, "UPDATE webscape_worlds SET schema_version=3 WHERE world_key=$1", p.worldKey); err != nil {
				return nil, dbError("fence legacy world", err)
			}
			return &state, nil
		}
	}
	if !blobs {
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
	state, err = snapshot.Decode(data)
	if err != nil {
		return nil, err
	}
	if err = p.writeCheckpoint(ctx, tx, state, nil); err != nil {
		return nil, err
	}
	if _, err = tx.Exec(ctx, "UPDATE webscape_snapshots SET schema_version=3 WHERE world_key=$1", p.worldKey); err != nil {
		return nil, dbError("fence legacy snapshot", err)
	}
	return &state, nil
}
