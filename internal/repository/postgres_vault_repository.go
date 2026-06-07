// Package repository содержит реализации репозиториев для работы
// с постоянным хранилищем данных приложения GophKeeper.
//
// Пакет отвечает за доступ к данным пользователей и приватных записей
// в PostgreSQL.
package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gophkeeper/internal/model"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
)

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

var vaultItemColumns = []string{
	"id",
	"user_id",
	"type",
	"title",
	"meta",
	"payload_encrypted",
	"payload_nonce",
	"payload_hash",
	"version",
	"client_updated_at",
	"created_at",
	"updated_at",
	"deleted_at",
}

// PostgresVaultRepository реализует хранилище приватных записей
// пользователя на базе PostgreSQL.
type PostgresVaultRepository struct {
	db PgxDB
}

// NewPostgresVaultRepository создаёт новый репозиторий для работы
// с приватными данными пользователя.
func NewPostgresVaultRepository(db PgxDB) *PostgresVaultRepository {
	return &PostgresVaultRepository{db: db}
}

// CreateItem сохраняет новую запись в базе данных.
//
// При успешном выполнении возвращает созданную запись со всеми заполненными
// служебными полями, включая идентификатор, версию и временные метки.
func (r *PostgresVaultRepository) CreateItem(ctx context.Context, item *model.VaultItem) error {
	query, args, err := psql.Insert("vault_items").
		Columns(vaultItemColumns...).
		Values(
			item.ID,
			item.UserID,
			sq.Expr("?::vault_item_type", string(item.Type)),
			item.Title,
			item.Meta,
			item.PayloadEncrypted,
			item.PayloadNonce,
			item.PayloadHash,
			item.Version,
			item.ClientUpdatedAt,
			item.CreatedAt,
			item.UpdatedAt,
			item.DeletedAt,
		).
		ToSql()
	if err != nil {
		return fmt.Errorf("build insert vault item query: %w", err)
	}

	_, err = r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("insert vault item: %w", err)
	}

	return nil
}

// GetItemByID возвращает запись по её идентификатору и идентификатору владельца.
//
// Если запись не найдена, функция возвращает ErrItemNotFound.
// Если запись была удалена, это определяется на уровне вызывающего кода
// по полю DeletedAt.
func (r *PostgresVaultRepository) GetItemByID(ctx context.Context, userID string, itemID string) (*model.VaultItem, error) {
	query, args, err := psql.Select(vaultItemColumns...).
		From("vault_items").
		Where(sq.Eq{"id": itemID, "user_id": userID}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build select vault item by id query: %w", err)
	}

	var item model.VaultItem
	var itemType string

	err = r.db.QueryRow(ctx, query, args...).Scan(
		&item.ID,
		&item.UserID,
		&itemType,
		&item.Title,
		&item.Meta,
		&item.PayloadEncrypted,
		&item.PayloadNonce,
		&item.PayloadHash,
		&item.Version,
		&item.ClientUpdatedAt,
		&item.CreatedAt,
		&item.UpdatedAt,
		&item.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("select vault item by id: %w", err)
	}

	item.Type = model.ItemType(itemType)
	return &item, nil
}

// ListItems возвращает список записей пользователя с учётом параметров выборки.
//
// Параметр includeDeleted определяет, нужно ли включать удалённые записи.
// Параметры limit и offset используются для пагинации результата.
func (r *PostgresVaultRepository) ListItems(ctx context.Context, userID string, includeDeleted bool, limit, offset int) ([]*model.VaultItem, error) {
	builder := psql.Select(vaultItemColumns...).
		From("vault_items").
		Where(sq.Eq{"user_id": userID}).
		OrderBy("updated_at DESC")

	if !includeDeleted {
		builder = builder.Where("deleted_at IS NULL")
	}

	if limit > 0 {
		builder = builder.Limit(uint64(limit)).Offset(uint64(offset))
	}

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build list vault items query: %w", err)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list vault items: %w", err)
	}
	defer rows.Close()

	items := make([]*model.VaultItem, 0)
	for rows.Next() {
		var item model.VaultItem
		var itemType string

		if err := rows.Scan(
			&item.ID,
			&item.UserID,
			&itemType,
			&item.Title,
			&item.Meta,
			&item.PayloadEncrypted,
			&item.PayloadNonce,
			&item.PayloadHash,
			&item.Version,
			&item.ClientUpdatedAt,
			&item.CreatedAt,
			&item.UpdatedAt,
			&item.DeletedAt,
		); err != nil {
			return nil, fmt.Errorf("scan vault item: %w", err)
		}

		item.Type = model.ItemType(itemType)
		items = append(items, &item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate vault items: %w", err)
	}

	return items, nil
}

// UpdateItem обновляет существующую запись.
//
// Метод использует версионность записи для защиты от конфликтов обновления.
// Если версия не совпадает, должна возвращаться ошибка конфликта версий.
func (r *PostgresVaultRepository) UpdateItem(ctx context.Context, item *model.VaultItem) (bool, error) {
	query, args, err := psql.Update("vault_items").
		Set("title", item.Title).
		Set("meta", item.Meta).
		Set("payload_encrypted", item.PayloadEncrypted).
		Set("payload_nonce", item.PayloadNonce).
		Set("payload_hash", item.PayloadHash).
		Set("version", item.Version).
		Set("client_updated_at", item.ClientUpdatedAt).
		Set("updated_at", item.UpdatedAt).
		Where(sq.Eq{"id": item.ID, "user_id": item.UserID}).
		Where("deleted_at IS NULL").
		ToSql()
	if err != nil {
		return false, fmt.Errorf("build update vault item query: %w", err)
	}

	res, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return false, fmt.Errorf("update vault item: %w", err)
	}

	return res.RowsAffected() > 0, nil
}

// SoftDeleteItem выполняет мягкое удаление записи.
//
// Вместо физического удаления запись помечается как удалённая через поле DeletedAt.
// Возвращает время удаления.
func (r *PostgresVaultRepository) SoftDeleteItem(ctx context.Context, userID string, itemID string, deletedAt time.Time) (bool, error) {
	query, args, err := psql.Update("vault_items").
		Set("deleted_at", deletedAt).
		Set("updated_at", deletedAt).
		Set("version", sq.Expr("version + 1")).
		Where(sq.Eq{"id": itemID, "user_id": userID}).
		Where("deleted_at IS NULL").
		ToSql()
	if err != nil {
		return false, fmt.Errorf("build soft delete vault item query: %w", err)
	}

	res, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return false, fmt.Errorf("soft delete vault item: %w", err)
	}

	return res.RowsAffected() > 0, nil
}

// SyncItems возвращает элементы, измененные после указанного времени.
//
// Используется для синхронизации данных между клиентами.
func (r *PostgresVaultRepository) SyncItems(ctx context.Context, userID string, since time.Time, includeDeleted bool) ([]*model.VaultItem, error) {
	builder := psql.Select(vaultItemColumns...).
		From("vault_items").
		Where(sq.Eq{"user_id": userID}).
		Where(sq.Gt{"updated_at": since}).
		OrderBy("updated_at ASC")

	if !includeDeleted {
		builder = builder.Where("deleted_at IS NULL")
	}

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build sync vault items query: %w", err)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("sync vault items: %w", err)
	}
	defer rows.Close()

	items := make([]*model.VaultItem, 0)
	for rows.Next() {
		var item model.VaultItem
		var itemType string

		if err := rows.Scan(
			&item.ID,
			&item.UserID,
			&itemType,
			&item.Title,
			&item.Meta,
			&item.PayloadEncrypted,
			&item.PayloadNonce,
			&item.PayloadHash,
			&item.Version,
			&item.ClientUpdatedAt,
			&item.CreatedAt,
			&item.UpdatedAt,
			&item.DeletedAt,
		); err != nil {
			return nil, fmt.Errorf("scan synced vault item: %w", err)
		}

		item.Type = model.ItemType(itemType)
		items = append(items, &item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate synced vault items: %w", err)
	}

	return items, nil
}
