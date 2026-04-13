package repository

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"gophkeeper/internal/model"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func TestNewPostgresVaultRepository(t *testing.T) {
	t.Parallel()

	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New returned error: %v", err)
	}
	defer db.Close()

	repo := NewPostgresVaultRepository(db)
	if repo == nil {
		t.Fatal("expected repo")
	}
}

func TestPostgresVaultRepository_CreateItem_Success(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New returned error: %v", err)
	}
	defer db.Close()

	repo := NewPostgresVaultRepository(db)

	now := time.Now().UTC().Round(0)
	item := &model.VaultItem{
		ID:               "item-1",
		UserID:           1,
		Type:             model.ItemTypeText,
		Title:            "note1",
		Meta:             "meta1",
		PayloadEncrypted: []byte("cipher"),
		PayloadNonce:     []byte("nonce"),
		PayloadHash:      "hash1",
		Version:          1,
		ClientUpdatedAt:  &now,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	mock.ExpectExec(regexp.QuoteMeta(`
		INSERT INTO vault_items (
			id, user_id, type, title, meta,
			payload_encrypted, payload_nonce, payload_hash,
			version, client_updated_at, created_at, updated_at, deleted_at
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
	`)).
		WithArgs(
			item.ID,
			item.UserID,
			item.Type,
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
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.CreateItem(context.Background(), item)
	if err != nil {
		t.Fatalf("CreateItem returned error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestPostgresVaultRepository_CreateItem_Error(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New returned error: %v", err)
	}
	defer db.Close()

	repo := NewPostgresVaultRepository(db)

	item := &model.VaultItem{
		ID:               "item-1",
		UserID:           1,
		Type:             model.ItemTypeText,
		Title:            "note1",
		PayloadEncrypted: []byte("cipher"),
		Version:          1,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	mock.ExpectExec(regexp.QuoteMeta(`
		INSERT INTO vault_items (
			id, user_id, type, title, meta,
			payload_encrypted, payload_nonce, payload_hash,
			version, client_updated_at, created_at, updated_at, deleted_at
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
	`)).
		WillReturnError(errors.New("insert error"))

	err = repo.CreateItem(context.Background(), item)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestPostgresVaultRepository_GetItemByID_Success(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New returned error: %v", err)
	}
	defer db.Close()

	repo := NewPostgresVaultRepository(db)

	now := time.Now().UTC().Round(0)

	rows := sqlmock.NewRows([]string{
		"id", "user_id", "type", "title", "meta",
		"payload_encrypted", "payload_nonce", "payload_hash",
		"version", "client_updated_at", "created_at", "updated_at", "deleted_at",
	}).AddRow(
		"item-1", int64(1), string(model.ItemTypeText), "note1", "meta1",
		[]byte("cipher"), []byte("nonce"), "hash1",
		int64(1), now, now, now, nil,
	)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT
			id, user_id, type, title, meta,
			payload_encrypted, payload_nonce, payload_hash,
			version, client_updated_at, created_at, updated_at, deleted_at
		FROM vault_items
		WHERE id = $1 AND user_id = $2
	`)).
		WithArgs("item-1", int64(1)).
		WillReturnRows(rows)

	item, err := repo.GetItemByID(context.Background(), 1, "item-1")
	if err != nil {
		t.Fatalf("GetItemByID returned error: %v", err)
	}
	if item == nil {
		t.Fatal("expected item")
	}
	if item.ID != "item-1" {
		t.Fatalf("expected item-1, got %q", item.ID)
	}
}

func TestPostgresVaultRepository_GetItemByID_NotFound(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New returned error: %v", err)
	}
	defer db.Close()

	repo := NewPostgresVaultRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT
			id, user_id, type, title, meta,
			payload_encrypted, payload_nonce, payload_hash,
			version, client_updated_at, created_at, updated_at, deleted_at
		FROM vault_items
		WHERE id = $1 AND user_id = $2
	`)).
		WithArgs("missing", int64(1)).
		WillReturnError(sql.ErrNoRows)

	item, err := repo.GetItemByID(context.Background(), 1, "missing")
	if err != nil {
		t.Fatalf("GetItemByID returned error: %v", err)
	}
	if item != nil {
		t.Fatal("expected nil item")
	}
}

func TestPostgresVaultRepository_ListItems_Success(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New returned error: %v", err)
	}
	defer db.Close()

	repo := NewPostgresVaultRepository(db)

	now := time.Now().UTC().Round(0)

	rows := sqlmock.NewRows([]string{
		"id", "user_id", "type", "title", "meta",
		"payload_encrypted", "payload_nonce", "payload_hash",
		"version", "client_updated_at", "created_at", "updated_at", "deleted_at",
	}).AddRow(
		"item-1", int64(1), string(model.ItemTypeText), "note1", "meta1",
		[]byte("cipher"), []byte("nonce"), "hash1",
		int64(1), now, now, now, nil,
	)

	mock.ExpectQuery("SELECT").
		WillReturnRows(rows)

	items, err := repo.ListItems(context.Background(), 1, false, 10, 0)
	if err != nil {
		t.Fatalf("ListItems returned error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
}

func TestPostgresVaultRepository_UpdateItem_Success(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New returned error: %v", err)
	}
	defer db.Close()

	repo := NewPostgresVaultRepository(db)

	now := time.Now().UTC().Round(0)
	item := &model.VaultItem{
		ID:               "item-1",
		UserID:           1,
		Title:            "updated",
		Meta:             "meta2",
		PayloadEncrypted: []byte("cipher2"),
		PayloadNonce:     []byte("nonce2"),
		PayloadHash:      "hash2",
		Version:          2,
		ClientUpdatedAt:  &now,
		UpdatedAt:        now,
	}

	mock.ExpectExec(regexp.QuoteMeta(`
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
	`)).
		WithArgs(
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
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	ok, err := repo.UpdateItem(context.Background(), item)
	if err != nil {
		t.Fatalf("UpdateItem returned error: %v", err)
	}
	if !ok {
		t.Fatal("expected updated=true")
	}
}

func TestPostgresVaultRepository_SoftDeleteItem_Success(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New returned error: %v", err)
	}
	defer db.Close()

	repo := NewPostgresVaultRepository(db)

	now := time.Now().UTC().Round(0)

	mock.ExpectExec(regexp.QuoteMeta(`
		UPDATE vault_items
		SET deleted_at = $1, updated_at = $1, version = version + 1
		WHERE id = $2 AND user_id = $3 AND deleted_at IS NULL
	`)).
		WithArgs(now, "item-1", int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	ok, err := repo.SoftDeleteItem(context.Background(), 1, "item-1", now)
	if err != nil {
		t.Fatalf("SoftDeleteItem returned error: %v", err)
	}
	if !ok {
		t.Fatal("expected deleted=true")
	}
}

func TestPostgresVaultRepository_SyncItems_Success(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New returned error: %v", err)
	}
	defer db.Close()

	repo := NewPostgresVaultRepository(db)

	now := time.Now().UTC().Round(0)

	rows := sqlmock.NewRows([]string{
		"id", "user_id", "type", "title", "meta",
		"payload_encrypted", "payload_nonce", "payload_hash",
		"version", "client_updated_at", "created_at", "updated_at", "deleted_at",
	}).AddRow(
		"item-1", int64(1), string(model.ItemTypeText), "note1", "meta1",
		[]byte("cipher"), []byte("nonce"), "hash1",
		int64(1), now, now, now, nil,
	)

	mock.ExpectQuery("SELECT").
		WillReturnRows(rows)

	items, err := repo.SyncItems(context.Background(), 1, now.Add(-time.Hour), true)
	if err != nil {
		t.Fatalf("SyncItems returned error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
}
