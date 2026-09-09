package persistence

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"webscape/server/snapshot"

	"github.com/jackc/pgx/v5"
)

// Postgres owns one connection and a session advisory lock. A second server may
// not write the same world. A lost connection fails closed; it is never silently
// replaced by an unlocked connection. Use a direct/session-pooled connection.
type Postgres struct {
	mu       sync.Mutex
	conn     *pgx.Conn
	worldKey string
	baseline *snapshot.State
	loaded   bool
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
	if err := p.initializeSchema(ctx); err != nil {
		return nil, err
	}

	success = true
	return p, nil
}

func (p *Postgres) Load(ctx context.Context) ([]byte, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	state, err := p.loadCheckpoint(ctx)
	if err != nil {
		return nil, err
	}
	p.baseline, p.loaded = state, true
	if state == nil {
		return nil, nil
	}
	return json.Marshal(state)
}

func (p *Postgres) Save(ctx context.Context, data []byte) error {
	state, err := snapshot.Decode(data)
	if err != nil {
		return err
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if !p.loaded {
		p.baseline, err = p.loadCheckpoint(ctx)
		if err != nil {
			return err
		}
		p.loaded = true
	}
	tx, err := p.conn.Begin(ctx)
	if err != nil {
		return dbError("begin save", err)
	}
	defer tx.Rollback(ctx)
	if err := p.writeCheckpoint(ctx, tx, state, p.baseline); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return dbError("commit save", err)
	}
	// Never advance the delta baseline before a successful commit.
	p.baseline = &state
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
