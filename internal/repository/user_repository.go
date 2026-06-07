// Package repository содержит реализации доступа к данным.
//
// Пакет реализует взаимодействие с базой данных PostgreSQL
// для работы с пользователями и элементами хранилища.
//
// Все операции с данными инкапсулированы в репозиториях,
// которые используются сервисным слоем.
package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// ErrUserAlreadyExists возвращается при попытке зарегистрировать
// пользователя с уже существующим логином.
var ErrUserAlreadyExists = errors.New("user already exists")

type User struct {
	ID       string
	Login    string
	Password string
}

// UserRepository определяет контракт для работы с пользователями.
//
// Используется сервисным слоем для абстракции от конкретной реализации БД.
type UserRepository interface {
	FindByUsername(ctx context.Context, login string) (*User, error)
	Create(ctx context.Context, login, passwordHash string) (string, error)
}

// PostgresUserRepository реализует работу с пользователями в PostgreSQL.
//
// Предоставляет методы для поиска пользователя и создания нового.
type PostgresUserRepository struct {
	db PgxDB
}

// NewPostgresUserRepository создает новый репозиторий пользователей.
func NewPostgresUserRepository(db PgxDB) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

// FindByUsername возвращает пользователя по логину.
//
// Возвращает:
// - пользователя при успехе
// - ErrUserNotFound, если пользователь не найден
// - ошибку при сбое запроса к базе данных
func (r *PostgresUserRepository) FindByUsername(ctx context.Context, login string) (*User, error) {
	const query = `
		SELECT id, login, password_hash
		FROM users
		WHERE login = $1
	`

	var user User
	err := r.db.QueryRow(ctx, query, login).Scan(
		&user.ID,
		&user.Login,
		&user.Password,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("query user by login: %w", err)
	}

	return &user, nil
}

// Create создает нового пользователя в базе данных.
//
// Возвращает:
// - созданного пользователя при успехе
// - ErrUserAlreadyExists, если пользователь уже существует
// - ошибку при сбое записи в базу данных
func (r *PostgresUserRepository) Create(ctx context.Context, login, passwordHash string) (string, error) {
	const query = `
		INSERT INTO users (login, password_hash)
		VALUES ($1, $2)
		RETURNING id
	`

	var id string
	err := r.db.QueryRow(ctx, query, login, passwordHash).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return "", ErrUserAlreadyExists
		}
		return "", fmt.Errorf("insert user: %w", err)
	}

	return id, nil
}
