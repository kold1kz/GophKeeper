package database

import (
	"strings"
	"testing"
)

func TestNewDB_BadDSN(t *testing.T) {
	t.Parallel()

	_, err := NewDB("://bad-dsn")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestNewDB_UnreachableDSN(t *testing.T) {
	t.Parallel()

	// несуществующий postgres, ожидаем ошибку на Ping
	_, err := NewDB("postgres://user:pass@127.0.0.1:1/db?sslmode=disable")
	if err == nil {
		t.Fatal("expected error")
	}

	if !strings.Contains(err.Error(), "failed to ping database") &&
		!strings.Contains(err.Error(), "failed to connect") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewDB_EmptyDSN(t *testing.T) {
	t.Parallel()

	db, err := NewDB("")
	if err != nil {
		t.Fatalf("NewDB returned error: %v", err)
	}
	if db == nil {
		t.Fatal("expected db")
	}
	if db.IsConfigured() {
		t.Fatal("expected db to be not configured")
	}
}

func TestPing_NotConfigured(t *testing.T) {
	t.Parallel()

	db := &DB{}

	err := db.Ping()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestClose_NotConfigured(t *testing.T) {
	t.Parallel()

	db := &DB{}

	if err := db.Close(); err != nil {
		t.Fatalf("Close returned error: %v", err)
	}
}

func TestGetPool_NotConfigured(t *testing.T) {
	t.Parallel()

	db := &DB{}

	if db.GetPool() != nil {
		t.Fatal("expected nil pgx pool")
	}
}

func TestIsConfigured_False(t *testing.T) {
	t.Parallel()

	db := &DB{}

	if db.IsConfigured() {
		t.Fatal("expected false")
	}
}
