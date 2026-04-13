package database

import (
	"database/sql"
	"strings"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestAutoMigrate_NilDB(t *testing.T) {
	err := AutoMigrate(nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "database is nil") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAutoMigrate_BadDB(t *testing.T) {
	db, err := sql.Open("pgx", "postgres://user:pass@127.0.0.1:1/db?sslmode=disable")
	if err != nil {
		t.Fatalf("sql.Open returned error: %v", err)
	}
	defer db.Close()

	err = AutoMigrate(db)
	if err == nil {
		t.Fatal("expected error")
	}
}
