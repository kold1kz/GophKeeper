// Package database содержит код для работы с базой данных PostgreSQL.
//
// Пакет отвечает за:
//   - создание подключения к базе данных;
//   - проверку доступности базы;
//   - запуск миграций;
//   - предоставление обёртки над *sql.DB.
package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// Pinger описывает объект, который умеет выполнять проверку доступности базы.
//
// Интерфейс используется там, где нужна только операция Ping.
type Pinger interface {
	Ping() error
}

// DB представляет собой обёртку над подключением к базе данных.
type DB struct {
	db *sql.DB
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
		return &DB{db: nil}, nil
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	if err := AutoMigrate(db); err != nil {
		return nil, fmt.Errorf("failed to apply migrations: %w", err)
	}

	log.Printf("Database connected and migrations applied")
	return &DB{db: db}, nil
}

// Ping проверяет доступность базы данных.
//
// Если подключение не настроено, функция возвращает ошибку.
func (d *DB) Ping() error {
	if d == nil || d.db == nil {
		return fmt.Errorf("database not configured")
	}
	return d.db.Ping()
}

// Close закрывает подключение к базе данных.
//
// Если подключение отсутствует, функция завершает работу без ошибки.
func (d *DB) Close() error {
	if d == nil || d.db == nil {
		return nil
	}
	return d.db.Close()
}

// GetDB возвращает внутренний объект *sql.DB.
func (d *DB) GetDB() *sql.DB {
	return d.db
}

// IsConfigured сообщает, настроено ли подключение к базе данных.
func (d *DB) IsConfigured() bool {
	return d.db != nil
}
