package database

import (
	"strings"
	"testing"
)

func TestAutoMigrate_NilDB(t *testing.T) {
	err := AutoMigrate("")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "database dsn is empty") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAutoMigrate_BadDB(t *testing.T) {
	err := AutoMigrate("postgres://user:pass@127.0.0.1:1/db?sslmode=disable")
	if err == nil {
		t.Fatal("expected error")
	}
}
