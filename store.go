package main

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Contact struct {
	Name    string
	Email   string
	Message string
}
type Store interface {
	Projects(context.Context) ([]Project, error)
	SaveContact(context.Context, Contact) error
	Ping(context.Context) error
	Mode() string
	Close()
}

// MemoryStore is explicit development mode: messages do NOT survive restart.
// A failed PostgreSQL connection never silently falls back to this store.
type MemoryStore struct {
	mu       sync.Mutex
	projects []Project
	contacts []Contact
}

func NewMemoryStore(projects []Project) *MemoryStore { return &MemoryStore{projects: projects} }
func (s *MemoryStore) Projects(ctx context.Context) ([]Project, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	// Return independent slices so callers cannot mutate shared configuration.
	b, _ := json.Marshal(s.projects)
	var p []Project
	_ = json.Unmarshal(b, &p)
	return p, nil
}
func (s *MemoryStore) SaveContact(ctx context.Context, c Contact) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.contacts) >= 1000 {
		return fmt.Errorf("development message capacity reached")
	}
	s.contacts = append(s.contacts, c)
	return nil
}
func (s *MemoryStore) Ping(ctx context.Context) error { return ctx.Err() }
func (s *MemoryStore) Mode() string                   { return "memory" }
func (s *MemoryStore) Close()                         {}

type PostgresStore struct{ pool *pgxpool.Pool }

func NewPostgresStore(ctx context.Context, dsn string, projects []Project) (*PostgresStore, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("invalid PostgreSQL configuration")
	}
	cfg.MaxConns = 5
	cfg.MinConns = 0
	cfg.MaxConnLifetime = 30 * time.Minute
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("initialize PostgreSQL: %w", err)
	}
	initialized := false
	defer func() {
		if !initialized {
			pool.Close()
		}
	}()
	fail := func(e error) (*PostgresStore, error) { return nil, e }
	if err = pool.Ping(ctx); err != nil {
		return fail(fmt.Errorf("connect PostgreSQL: %w", err))
	}
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fail(err)
	}
	defer tx.Rollback(ctx)
	// Serializes startup schema initialization across replicas. This small initial
	// migration is additive; use versioned migrations for future schema changes.
	if _, err = tx.Exec(ctx, "SELECT pg_advisory_xact_lock(6192042026)"); err != nil {
		return fail(err)
	}
	for _, ddl := range []string{
		`CREATE TABLE IF NOT EXISTS portfolio_projects (id text PRIMARY KEY, position integer NOT NULL, payload jsonb NOT NULL)`,
		`CREATE TABLE IF NOT EXISTS portfolio_contacts (id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY, name text NOT NULL CHECK (char_length(name) BETWEEN 2 AND 80), email text NOT NULL CHECK (char_length(email) <= 254), message text NOT NULL CHECK (char_length(message) BETWEEN 20 AND 2000), created_at timestamptz NOT NULL DEFAULT now())`,
		`CREATE INDEX IF NOT EXISTS portfolio_contacts_created_at_idx ON portfolio_contacts (created_at)`,
	} {
		if _, err = tx.Exec(ctx, ddl); err != nil {
			return fail(fmt.Errorf("initialize schema: %w", err))
		}
	}
	// profile.json is the source of truth for portfolio content, synchronized on
	// startup. Runtime contact data is stored separately and never reseeded.
	for i, p := range projects {
		b, e := json.Marshal(p)
		if e != nil {
			return fail(e)
		}
		_, err = tx.Exec(ctx, `INSERT INTO portfolio_projects(id,position,payload) VALUES($1,$2,$3::jsonb) ON CONFLICT(id) DO UPDATE SET position=excluded.position,payload=excluded.payload`, p.ID, i, string(b))
		if err != nil {
			return fail(fmt.Errorf("seed projects: %w", err))
		}
	}
	ids := make([]string, 0, len(projects))
	for _, p := range projects {
		ids = append(ids, p.ID)
	}
	if _, err = tx.Exec(ctx, `DELETE FROM portfolio_projects WHERE NOT (id = ANY($1::text[]))`, ids); err != nil {
		return fail(err)
	}
	if err = tx.Commit(ctx); err != nil {
		return fail(err)
	}
	initialized = true
	return &PostgresStore{pool: pool}, nil
}
func (s *PostgresStore) Projects(ctx context.Context) ([]Project, error) {
	rows, err := s.pool.Query(ctx, `SELECT payload FROM portfolio_projects ORDER BY position,id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]Project, 0)
	for rows.Next() {
		var b []byte
		var p Project
		if err = rows.Scan(&b); err != nil {
			return nil, err
		}
		if err = json.Unmarshal(b, &p); err != nil {
			return nil, err
		}
		result = append(result, p)
	}
	return result, rows.Err()
}
func (s *PostgresStore) SaveContact(ctx context.Context, c Contact) error {
	_, err := s.pool.Exec(ctx, `INSERT INTO portfolio_contacts(name,email,message) VALUES($1,$2,$3)`, c.Name, c.Email, c.Message)
	return err
}
func (s *PostgresStore) Ping(ctx context.Context) error { return s.pool.Ping(ctx) }
func (s *PostgresStore) Mode() string                   { return "postgresql" }
func (s *PostgresStore) Close()                         { s.pool.Close() }
func (s *PostgresStore) DeleteExpired(ctx context.Context) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM portfolio_contacts WHERE created_at < now() - interval '30 days'`)
	return err
}
