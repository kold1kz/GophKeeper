package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"gophkeeper/internal/model"
)

type PostgresVaultRepository struct {
	db *sql.DB
}

func NewPostgresVaultRepository(db *sql.DB) *PostgresVaultRepository {
	return &PostgresVaultRepository{db: db}
}

func (r *PostgresVaultRepository) CreateItem(ctx context.Context, item *model.VaultItem) error {
	const query = `
		INSERT INTO vault_items (
			id,
			user_id,
			type,
			title,
			meta,
			payload_encrypted,
			payload_nonce,
			payload_hash,
			version,
			client_updated_at,
			created_at,
			updated_at,
			deleted_at
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		item.ID,
		item.UserID,
		string(item.Type),
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
	)
	if err != nil {
		return fmt.Errorf("insert vault item: %w", err)
	}

	return nil
}

func (r *PostgresVaultRepository) GetItemByID(ctx context.Context, userID int64, itemID string) (*model.VaultItem, error) {
	const query = `
		SELECT
			id,
			user_id,
			type,
			title,
			meta,
			payload_encrypted,
			payload_nonce,
			payload_hash,
			version,
			client_updated_at,
			created_at,
			updated_at,
			deleted_at
		FROM vault_items
		WHERE id = $1 AND user_id = $2
	`

	var item model.VaultItem
	var itemType string

	err := r.db.QueryRowContext(ctx, query, itemID, userID).Scan(
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
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("select vault item by id: %w", err)
	}

	item.Type = model.ItemType(itemType)
	return &item, nil
}

func (r *PostgresVaultRepository) ListItems(ctx context.Context, userID int64, includeDeleted bool, limit, offset int) ([]*model.VaultItem, error) {
	query := `
		SELECT
			id,
			user_id,
			type,
			title,
			meta,
			payload_encrypted,
			payload_nonce,
			payload_hash,
			version,
			client_updated_at,
			created_at,
			updated_at,
			deleted_at
		FROM vault_items
		WHERE user_id = $1
	`
	args := []any{userID}

	if !includeDeleted {
		query += ` AND deleted_at IS NULL`
	}

	query += ` ORDER BY updated_at DESC`

	if limit > 0 {
		query += ` LIMIT $2 OFFSET $3`
		args = append(args, limit, offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
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

func (r *PostgresVaultRepository) UpdateItem(ctx context.Context, item *model.VaultItem) (bool, error) {
	const query = `
		UPDATE vault_items
		SET
			title = $1,
			meta = $2,
			payload_encrypted = $3,
			payload_nonce = $4,
			payload_hash = $5,
			version = $6,
			client_updated_at = $7,
			updated_at = $8
		WHERE id = $9 AND user_id = $10 AND deleted_at IS NULL
	`

	res, err := r.db.ExecContext(
		ctx,
		query,
		item.Title,
		item.Meta,
		item.PayloadEncrypted,
		item.PayloadNonce,
		item.PayloadHash,
		item.Version,
		item.ClientUpdatedAt,
		item.UpdatedAt,
		item.ID,
		item.UserID,
	)
	if err != nil {
		return false, fmt.Errorf("update vault item: %w", err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("update vault item rows affected: %w", err)
	}

	return affected > 0, nil
}

func (r *PostgresVaultRepository) SoftDeleteItem(ctx context.Context, userID int64, itemID string, deletedAt time.Time) (bool, error) {
	const query = `
		UPDATE vault_items
		SET
			deleted_at = $1,
			updated_at = $1,
			version = version + 1
		WHERE id = $2 AND user_id = $3 AND deleted_at IS NULL
	`

	res, err := r.db.ExecContext(ctx, query, deletedAt, itemID, userID)
	if err != nil {
		return false, fmt.Errorf("soft delete vault item: %w", err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("soft delete rows affected: %w", err)
	}

	return affected > 0, nil
}

func (r *PostgresVaultRepository) SyncItems(ctx context.Context, userID int64, since time.Time, includeDeleted bool) ([]*model.VaultItem, error) {
	query := `
		SELECT
			id,
			user_id,
			type,
			title,
			meta,
			payload_encrypted,
			payload_nonce,
			payload_hash,
			version,
			client_updated_at,
			created_at,
			updated_at,
			deleted_at
		FROM vault_items
		WHERE user_id = $1 AND updated_at > $2
	`
	args := []any{userID, since}

	if !includeDeleted {
		query += ` AND deleted_at IS NULL`
	}

	query += ` ORDER BY updated_at ASC`

	rows, err := r.db.QueryContext(ctx, query, args...)
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
