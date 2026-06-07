// Package database содержит код для работы с базой данных PostgreSQL.
//
// Пакет отвечает за:
//   - создание подключения к базе данных;
//   - проверку доступности базы;
//   - запуск миграций;
//   - предоставление обёртки над *sql.DB.
package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// Pinger описывает объект, который умеет выполнять проверку доступности базы.
//
// Интерфейс используется там, где нужна только операция Ping.
type Pinger interface {
	Ping(ctx context.Context) error
}

// DB представляет собой обёртку над подключением к базе данных.
type DB struct {
	pool *pgxpool.Pool
}

// NewDB создаёт новое подключение к PostgreSQL по переданному DSN.
//
// Если DSN пустой, функция возвращает объект DB без активного подключения.
//
// После успешного открытия подключения функция:
//   - выполняет Ping для проверки доступности базы;
//   - запускает миграции;
//   - возвращает готовый объект DB.
func NewDB(dsn string) (*DB, error) {
	if dsn == "" {
		return &DB{pool: nil}, nil
	}

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	config.MaxConns = 20
	config.MinConns = 5
	config.MaxConnLifetime = 30 * time.Minute
	config.MaxConnIdleTime = 10 * time.Minute
	config.HealthCheckPeriod = time.Minute

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	if err := AutoMigrate(dsn); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to apply migrations: %w", err)
	}

	log.Printf("Database connected and migrations applied")
	return &DB{pool: pool}, nil
}

// Ping проверяет доступность базы данных.
//
// Если подключение не настроено, функция возвращает ошибку.
func (d *DB) Ping() error {
	if d == nil || d.pool == nil {
		return fmt.Errorf("database not configured")
	}
	return d.pool.Ping(context.Background())
}

// Close закрывает подключение к базе данных.
//
// Если подключение отсутствует, функция завершает работу без ошибки.
func (d *DB) Close() error {
	if d == nil || d.pool == nil {
		return nil
	}
	d.pool.Close()
	return nil
}

// GetPool возвращает внутренний пул подключений pgx.
func (d *DB) GetPool() *pgxpool.Pool {
	return d.pool
}

// IsConfigured сообщает, настроено ли подключение к базе данных.
func (d *DB) IsConfigured() bool {
	return d.pool != nil
}
