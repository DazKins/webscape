package persistence

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"webscape/server/snapshot"

	"github.com/jackc/pgx/v5"
)

func (p *Postgres) loadCheckpoint(ctx context.Context) (*snapshot.State, error) {
	tx, err := p.conn.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead})
	if err != nil {
		return nil, dbError("begin load", err)
	}
	defer tx.Rollback(ctx)
	state := snapshot.State{Entities: map[string]map[string]snapshot.Component{}}
	var schema int
	var tick string
	err = tx.QueryRow(ctx, "SELECT schema_version,snapshot_version,content_hash,tick::text FROM webscape_worlds WHERE world_key=$1", p.worldKey).Scan(&schema, &state.Version, &state.ContentHash, &tick)
	if err == pgx.ErrNoRows {
		legacy, err := p.migrateLegacyWorld(ctx, tx)
		if err != nil {
			return nil, err
		}
		if err := tx.Commit(ctx); err != nil {
			return nil, dbError("commit legacy load", err)
		}
		return legacy, nil
	}
	if err != nil {
		return nil, dbError("load world metadata", err)
	}
	if schema != 2 {
		return nil, fmt.Errorf("unsupported PostgreSQL snapshot schema version %d", schema)
	}
	state.Tick, err = strconv.ParseUint(tick, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid saved tick")
	}
	rows, err := tx.Query(ctx, "SELECT entity_id::text,component_id,version,data FROM webscape_components WHERE world_key=$1", p.worldKey)
	if err != nil {
		return nil, dbError("load components", err)
	}
	for rows.Next() {
		var entity, id string
		var c snapshot.Component
		if err := rows.Scan(&entity, &id, &c.Version, &c.Data); err != nil {
			rows.Close()
			return nil, dbError("decode component row", err)
		}
		if state.Entities[entity] == nil {
			state.Entities[entity] = map[string]snapshot.Component{}
		}
		state.Entities[entity][id] = c
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, dbError("read components", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, dbError("commit load", err)
	}
	return &state, nil
}

// writeCheckpoint only touches changed component rows. Metadata, additions,
// changes and deletions share one transaction, preserving atomic item transfers.
func (p *Postgres) writeCheckpoint(ctx context.Context, tx pgx.Tx, current snapshot.State, previous *snapshot.State) error {
	batch := &pgx.Batch{}
	batch.Queue(`INSERT INTO webscape_worlds(world_key,schema_version,snapshot_version,content_hash,tick)
 VALUES($1,2,$2,$3,$4::numeric) ON CONFLICT(world_key) DO UPDATE
 SET snapshot_version=EXCLUDED.snapshot_version,content_hash=EXCLUDED.content_hash,tick=EXCLUDED.tick,saved_at=now()`, p.worldKey, current.Version, current.ContentHash, strconv.FormatUint(current.Tick, 10))
	for entity, components := range current.Entities {
		for id, c := range components {
			if previous != nil {
				if old, ok := previous.Entities[entity][id]; ok && sameComponent(c, old) {
					continue
				}
			}
			batch.Queue(`INSERT INTO webscape_components(world_key,entity_id,component_id,version,data)
 VALUES($1,$2::uuid,$3,$4,$5::jsonb) ON CONFLICT(world_key,entity_id,component_id) DO UPDATE
 SET version=EXCLUDED.version,data=EXCLUDED.data`, p.worldKey, entity, id, c.Version, string(c.Data))
		}
	}
	if previous != nil {
		for entity, components := range previous.Entities {
			for id := range components {
				if _, present := current.Entities[entity][id]; !present {
					batch.Queue("DELETE FROM webscape_components WHERE world_key=$1 AND entity_id=$2::uuid AND component_id=$3", p.worldKey, entity, id)
				}
			}
		}
	}
	results := tx.SendBatch(ctx, batch)
	if err := results.Close(); err != nil {
		return dbError("write checkpoint", err)
	}
	return nil
}

// PostgreSQL jsonb can reorder keys. Compare decoded JSON with exact number text
// so the first save after a load does not rewrite every unchanged component.
func sameComponent(a, b snapshot.Component) bool {
	if a.Version != b.Version {
		return false
	}
	canonical := func(data []byte) []byte {
		var value any
		decoder := json.NewDecoder(bytes.NewReader(data))
		decoder.UseNumber()
		if decoder.Decode(&value) != nil {
			return nil
		}
		result, _ := json.Marshal(value)
		return result
	}
	return bytes.Equal(canonical(a.Data), canonical(b.Data))
}
