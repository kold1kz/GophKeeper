package repository

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func TestPostgresUserRepository_FindByUsername_Success(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New returned error: %v", err)
	}
	defer db.Close()

	repo := NewPostgresUserRepository(db)

	rows := sqlmock.NewRows([]string{
		"id",
		"login",
		"password_hash",
	}).AddRow(1, "u1", "hash1")

	mock.ExpectQuery(regexp.QuoteMeta(`
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
	if user.ID != 1 {
		t.Fatalf("expected ID 1, got %d", user.ID)
	}
	if user.Login != "u1" {
		t.Fatalf("expected login u1, got %q", user.Login)
	}
	if user.Password != "hash1" {
		t.Fatalf("expected password hash1, got %q", user.Password)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sqlmock expectations: %v", err)
	}
}

func TestPostgresUserRepository_FindByUsername_NotFound(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New returned error: %v", err)
	}
	defer db.Close()

	repo := NewPostgresUserRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, login, password_hash
		FROM users
		WHERE login = $1
	`)).
		WithArgs("missing").
		WillReturnError(sql.ErrNoRows)

	user, err := repo.FindByUsername(context.Background(), "missing")
	if err != nil {
		t.Fatalf("FindByUsername returned error: %v", err)
	}
	if user != nil {
		t.Fatalf("expected nil user, got %+v", user)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sqlmock expectations: %v", err)
	}
}

func TestPostgresUserRepository_FindByUsername_QueryError(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New returned error: %v", err)
	}
	defer db.Close()

	repo := NewPostgresUserRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta(`
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

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sqlmock expectations: %v", err)
	}
}

func TestPostgresUserRepository_Create_Success(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New returned error: %v", err)
	}
	defer db.Close()

	repo := NewPostgresUserRepository(db)

	rows := sqlmock.NewRows([]string{"id"}).AddRow(7)

	mock.ExpectQuery(regexp.QuoteMeta(`
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
	if id != 7 {
		t.Fatalf("expected id 7, got %d", id)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sqlmock expectations: %v", err)
	}
}

func TestPostgresUserRepository_Create_Error(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New returned error: %v", err)
	}
	defer db.Close()

	repo := NewPostgresUserRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta(`
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
	if id != 0 {
		t.Fatalf("expected id 0, got %d", id)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sqlmock expectations: %v", err)
	}
}
