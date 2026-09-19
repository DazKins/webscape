package persistence

import "context"

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
