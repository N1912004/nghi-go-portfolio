package main

import (
	"context"
	"io"
	"log/slog"
	"strings"
	"testing"
)

func TestProductionRequiresDatabase(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("PUBLIC_URL", "https://example.com")
	t.Setenv("DATABASE_URL", "")
	t.Setenv("PORT", "8080")
	err := run(slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err == nil || !strings.Contains(err.Error(), "DATABASE_URL") {
		t.Fatalf("expected closed startup without DB, got %v", err)
	}
}
func TestInvalidPortFails(t *testing.T) {
	t.Setenv("PORT", "not-a-port")
	if err := run(slog.New(slog.NewTextHandler(io.Discard, nil))); err == nil {
		t.Fatal("invalid port accepted")
	}
}
func TestMalformedPostgresConfig(t *testing.T) {
	if s, err := NewPostgresStore(context.Background(), "postgres://%", nil); err == nil || s != nil {
		t.Fatal("invalid database config accepted")
	}
}
