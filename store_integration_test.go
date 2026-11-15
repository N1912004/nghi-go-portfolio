package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// A disposable schema isolates this test from real portfolio data. Only run
// against a dedicated test database whose user can CREATE/DROP schemas.
func TestPostgresIntegration(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set: PostgreSQL integration not executed")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	admin, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close()
	b := make([]byte, 8)
	if _, err = rand.Read(b); err != nil {
		t.Fatal(err)
	}
	schema := "portfolio_test_" + hex.EncodeToString(b)
	if _, err = admin.Exec(ctx, "CREATE SCHEMA "+schema); err != nil {
		t.Fatal(err)
	}
	defer func() {
		cleanup, c := context.WithTimeout(context.Background(), 5*time.Second)
		defer c()
		if _, e := admin.Exec(cleanup, "DROP SCHEMA "+schema+" CASCADE"); e != nil {
			t.Error(e)
		}
	}()
	u, err := url.Parse(dsn)
	if err != nil {
		t.Fatal(err)
	}
	q := u.Query()
	q.Set("search_path", schema)
	u.RawQuery = q.Encode()
	p, err := loadProfile()
	if err != nil {
		t.Fatal(err)
	}
	s, err := NewPostgresStore(ctx, u.String(), p.Projects)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	projects, err := s.Projects(ctx)
	if err != nil || len(projects) != len(p.Projects) {
		t.Fatal(err, len(projects))
	}
	if err = s.SaveContact(ctx, Contact{Name: "Test Recruiter", Email: "test@example.com", Message: "A real PostgreSQL persistence test message."}); err != nil {
		t.Fatal(err)
	}
	var count int
	if err = s.pool.QueryRow(ctx, "SELECT count(*) FROM portfolio_contacts").Scan(&count); err != nil || count != 1 {
		t.Fatal(err, count)
	}
	if _, err = s.pool.Exec(ctx, "UPDATE portfolio_contacts SET created_at = now() - interval '31 days'"); err != nil {
		t.Fatal(err)
	}
	if err = s.DeleteExpired(ctx); err != nil {
		t.Fatal(err)
	}
	if err = s.pool.QueryRow(ctx, "SELECT count(*) FROM portfolio_contacts").Scan(&count); err != nil || count != 0 {
		t.Fatal(err, count)
	}
	if err = s.Ping(ctx); err != nil {
		t.Fatal(err)
	}
}
