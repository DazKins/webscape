package persistence

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/jackc/pgx/v5"
)

// Postgres owns one connection and a session advisory lock. A second server may
// not write the same world. A lost connection fails closed; it is never silently
// replaced by an unlocked connection. Use a direct/session-pooled connection.
type Postgres struct {
	mu       sync.Mutex
	conn     *pgx.Conn
	worldKey string
}

func OpenPostgres(ctx context.Context, connectionString, worldKey string) (*Postgres, error) {
	config, err := pgx.ParseConfig(connectionString)
	if err != nil {
		return nil, errors.New("invalid PostgreSQL connection configuration")
	}
	conn, err := pgx.ConnectConfig(ctx, config)
	if err != nil {
		return nil, dbError("connect", err)
	}
	p := &Postgres{conn: conn, worldKey: worldKey}
	success := false
	defer func() {
		if !success {
			conn.Close(context.Background())
		}
	}()
	var locked bool
	if err := conn.QueryRow(ctx, "SELECT pg_try_advisory_lock(hashtextextended($1, 0))", "webscape:"+worldKey).Scan(&locked); err != nil {
		return nil, dbError("lock world", err)
	}
	if !locked {
		return nil, errors.New("another server already owns persistence.worldKey")
	}
	// Schema v1: an atomic whole-world record. JSON save versions are independently
	// validated by the game. Future SQL migrations belong in this adapter only.
	if _, err := conn.Exec(ctx, `CREATE TABLE IF NOT EXISTS webscape_snapshots (
 world_key text PRIMARY KEY,
 schema_version integer NOT NULL CHECK (schema_version > 0),
 snapshot jsonb NOT NULL,
 saved_at timestamptz NOT NULL DEFAULT now()
 )`); err != nil {
		return nil, dbError("initialize schema", err)
	}
	success = true
	return p, nil
}

func (p *Postgres) Load(ctx context.Context) ([]byte, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	var version int
	var data []byte
	err := p.conn.QueryRow(ctx, "SELECT schema_version, snapshot FROM webscape_snapshots WHERE world_key=$1", p.worldKey).Scan(&version, &data)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, dbError("load", err)
	}
	if version != 1 {
		return nil, fmt.Errorf("unsupported PostgreSQL snapshot schema version %d", version)
	}
	return data, nil
}

func (p *Postgres) Save(ctx context.Context, data []byte) error {
	if !json.Valid(data) {
		return errors.New("snapshot is not valid JSON")
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	// One statement is one transaction: entity additions, deletions and item
	// transfers either all commit or leave the previous snapshot intact.
	_, err := p.conn.Exec(ctx, `INSERT INTO webscape_snapshots (world_key,schema_version,snapshot)
 VALUES ($1,1,$2::jsonb) ON CONFLICT (world_key) DO UPDATE
 SET snapshot=EXCLUDED.snapshot, schema_version=1, saved_at=now()`, p.worldKey, string(data))
	if err != nil {
		return dbError("save", err)
	}
	return nil
}

func (p *Postgres) Close(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.conn.Close(ctx)
}

// Driver errors can include connection strings, credentials or SQL values. Keep
// operational context and SQLSTATE, without logging user-supplied secrets.
func dbError(operation string, err error) error {
	if errors.Is(err, context.DeadlineExceeded) {
		return fmt.Errorf("PostgreSQL %s timed out", operation)
	}
	if errors.Is(err, context.Canceled) {
		return fmt.Errorf("PostgreSQL %s canceled", operation)
	}
	var state interface{ SQLState() string }
	if errors.As(err, &state) {
		return fmt.Errorf("PostgreSQL %s failed (SQLSTATE %s)", operation, state.SQLState())
	}
	return fmt.Errorf("PostgreSQL %s failed; check database availability, credentials and TLS configuration", operation)
}
