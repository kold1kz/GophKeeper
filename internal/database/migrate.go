// Package database содержит код для работы с базой данных PostgreSQL.
package database

import (
	"database/sql"
	"fmt"
	migrate "github.com/golang-migrate/migrate/v4"
	postgresmigrate "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// AutoMigrate применяет миграции к базе данных.
//
// Функция использует файловый источник миграций "file://migrations"
// и применяет все ещё не выполненные миграции к базе PostgreSQL.
//
// Если миграций для применения нет, ошибка ErrNoChange не считается ошибкой выполнения.
func AutoMigrate(db *sql.DB) error {
	if db == nil {
		return fmt.Errorf("database is nil")
	}

	driver, err := postgresmigrate.WithInstance(db, &postgresmigrate.Config{})
	if err != nil {
		return fmt.Errorf("create postgres migrate driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("create migrate instance: %w", err)
	}
	defer func() {
		_, _ = m.Close()
	}()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("apply migrations: %w", err)
	}

	return nil
}
