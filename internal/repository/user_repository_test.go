package repository

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v4"
)

func TestPostgresUserRepository_FindByUsername_Success(t *testing.T) {
	t.Parallel()

	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("pgxmock.NewPool returned error: %v", err)
	}
	defer db.Close()

	repo := NewPostgresUserRepository(db)
	userID := "550e8400-e29b-41d4-a716-446655440000"

	rows := pgxmock.NewRows([]string{"id", "login", "password_hash"}).
		AddRow(userID, "u1", "hash1")

	db.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, login, password_hash
		FROM users
		WHERE login = $1
	`)).
		WithArgs("u1").
		WillReturnRows(rows)

	user, err := repo.FindByUsername(context.Background(), "u1")
	if err != nil {
		t.Fatalf("FindByUsername returned error: %v", err)
	}
	if user == nil {
		t.Fatal("expected user, got nil")
	}
	if user.ID != userID {
		t.Fatalf("expected ID %s, got %s", userID, user.ID)
	}
	if user.Login != "u1" {
		t.Fatalf("expected login u1, got %q", user.Login)
	}
	if user.Password != "hash1" {
		t.Fatalf("expected password hash1, got %q", user.Password)
	}

	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet pgxmock expectations: %v", err)
	}
}

func TestPostgresUserRepository_FindByUsername_NotFound(t *testing.T) {
	t.Parallel()

	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("pgxmock.NewPool returned error: %v", err)
	}
	defer db.Close()

	repo := NewPostgresUserRepository(db)

	db.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, login, password_hash
		FROM users
		WHERE login = $1
	`)).
		WithArgs("missing").
		WillReturnError(pgx.ErrNoRows)

	user, err := repo.FindByUsername(context.Background(), "missing")
	if err != nil {
		t.Fatalf("FindByUsername returned error: %v", err)
	}
	if user != nil {
		t.Fatalf("expected nil user, got %+v", user)
	}

	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet pgxmock expectations: %v", err)
	}
}

func TestPostgresUserRepository_FindByUsername_QueryError(t *testing.T) {
	t.Parallel()

	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("pgxmock.NewPool returned error: %v", err)
	}
	defer db.Close()

	repo := NewPostgresUserRepository(db)

	db.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, login, password_hash
		FROM users
		WHERE login = $1
	`)).
		WithArgs("u1").
		WillReturnError(errors.New("db error"))

	user, err := repo.FindByUsername(context.Background(), "u1")
	if err == nil {
		t.Fatal("expected error")
	}
	if user != nil {
		t.Fatalf("expected nil user, got %+v", user)
	}

	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet pgxmock expectations: %v", err)
	}
}

func TestPostgresUserRepository_Create_Success(t *testing.T) {
	t.Parallel()

	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("pgxmock.NewPool returned error: %v", err)
	}
	defer db.Close()

	repo := NewPostgresUserRepository(db)
	userID := "550e8400-e29b-41d4-a716-446655440000"
	rows := pgxmock.NewRows([]string{"id"}).AddRow(userID)

	db.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO users (login, password_hash)
		VALUES ($1, $2)
		RETURNING id
	`)).
		WithArgs("u1", "hash1").
		WillReturnRows(rows)

	id, err := repo.Create(context.Background(), "u1", "hash1")
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if id != userID {
		t.Fatalf("expected id %s, got %s", userID, id)
	}

	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet pgxmock expectations: %v", err)
	}
}

func TestPostgresUserRepository_Create_Error(t *testing.T) {
	t.Parallel()

	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("pgxmock.NewPool returned error: %v", err)
	}
	defer db.Close()

	repo := NewPostgresUserRepository(db)

	db.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO users (login, password_hash)
		VALUES ($1, $2)
		RETURNING id
	`)).
		WithArgs("u1", "hash1").
		WillReturnError(errors.New("insert error"))

	id, err := repo.Create(context.Background(), "u1", "hash1")
	if err == nil {
		t.Fatal("expected error")
	}
	if id != "" {
		t.Fatalf("expected empty id, got %s", id)
	}

	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet pgxmock expectations: %v", err)
	}
}
