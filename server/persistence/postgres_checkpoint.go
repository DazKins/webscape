package persistence

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"webscape/server/snapshot"

	"github.com/jackc/pgx/v5"
)

func (p *Postgres) loadCheckpoint(ctx context.Context) (*snapshot.State, error) {
	tx, err := p.conn.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead})
	if err != nil {
		return nil, dbError("begin load", err)
	}
	defer tx.Rollback(ctx)
	var version int
	err = tx.QueryRow(ctx, "SELECT version FROM webscape_player_saves WHERE world_key=$1", p.worldKey).Scan(&version)
	if err == pgx.ErrNoRows {
		state, err := p.migrateLegacyWorld(ctx, tx)
		if err != nil {
			return nil, err
		}
		if err := tx.Commit(ctx); err != nil {
			return nil, dbError("commit migration", err)
		}
		return state, nil
	}
	if err != nil {
		return nil, dbError("load save version", err)
	}
	if version != 2 {
		return nil, fmt.Errorf("unsupported player save version %d", version)
	}
	state := snapshot.State{Version: version, Players: map[string]map[string]snapshot.Component{}}
	rows, err := tx.Query(ctx, "SELECT player_id::text,components FROM webscape_players WHERE world_key=$1", p.worldKey)
	if err != nil {
		return nil, dbError("load players", err)
	}
	for rows.Next() {
		var id string
		var data []byte
		if err := rows.Scan(&id, &data); err != nil {
			rows.Close()
			return nil, dbError("read player", err)
		}
		var components map[string]snapshot.Component
		if err := json.Unmarshal(data, &components); err != nil {
			rows.Close()
			return nil, fmt.Errorf("invalid player components")
		}
		state.Players[id] = components
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, dbError("read players", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, dbError("commit load", err)
	}
	return &state, nil
}

// Each changed player is an atomic row; the whole checkpoint is transactional.
// Absence is not deletion: a disconnected player must never lose their save.
func (p *Postgres) writeCheckpoint(ctx context.Context, tx pgx.Tx, current snapshot.State, previous *snapshot.State) error {
	batch := &pgx.Batch{}
	batch.Queue(`INSERT INTO webscape_player_saves(world_key,version) VALUES($1,2)
 ON CONFLICT(world_key) DO UPDATE SET saved_at=now()`, p.worldKey)
	for id, components := range current.Players {
		if previous != nil && samePlayer(components, previous.Players[id]) {
			continue
		}
		data, err := json.Marshal(components)
		if err != nil {
			return err
		}
		batch.Queue(`INSERT INTO webscape_players(world_key,player_id,components) VALUES($1,$2::uuid,$3::jsonb)
 ON CONFLICT(world_key,player_id) DO UPDATE SET components=EXCLUDED.components,updated_at=now()`, p.worldKey, id, string(data))
	}
	if err := tx.SendBatch(ctx, batch).Close(); err != nil {
		return dbError("write players", err)
	}
	return nil
}
func samePlayer(a, b map[string]snapshot.Component) bool {
	if len(a) != len(b) {
		return false
	}
	for id, c := range a {
		if !sameComponent(c, b[id]) {
			return false
		}
	}
	return true
}

// jsonb reorders keys; compare canonical JSON without rounding numbers.
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
