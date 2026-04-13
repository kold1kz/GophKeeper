package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

var ErrUserAlreadyExists = errors.New("user already exists")

type User struct {
	ID       int
	Login    string
	Password string
}

type UserRepository interface {
	FindByUsername(ctx context.Context, login string) (*User, error)
	Create(ctx context.Context, login, passwordHash string) (int, error)
}

type PostgresUserRepository struct {
	db *sql.DB
}

func NewPostgresUserRepository(db *sql.DB) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

func (r *PostgresUserRepository) FindByUsername(ctx context.Context, login string) (*User, error) {
	const query = `
		SELECT id, login, password_hash
		FROM users
		WHERE login = $1
	`

	var user User
	err := r.db.QueryRowContext(ctx, query, login).Scan(
		&user.ID,
		&user.Login,
		&user.Password,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("query user by login: %w", err)
	}

	return &user, nil
}

func (r *PostgresUserRepository) Create(ctx context.Context, login, passwordHash string) (int, error) {
	const query = `
		INSERT INTO users (login, password_hash)
		VALUES ($1, $2)
		RETURNING id
	`

	var id int
	err := r.db.QueryRowContext(ctx, query, login, passwordHash).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return 0, ErrUserAlreadyExists
		}
		return 0, fmt.Errorf("insert user: %w", err)
	}

	return id, nil
}
