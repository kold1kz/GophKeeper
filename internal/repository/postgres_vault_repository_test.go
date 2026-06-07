package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"gophkeeper/internal/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v4"
)

const testUserID = "550e8400-e29b-41d4-a716-446655440000"

var testItemID = uuid.MustParse("7d444840-9dc0-11d1-b245-5ffdce74fad2")

func TestNewPostgresVaultRepository(t *testing.T) {
	t.Parallel()

	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("pgxmock.NewPool returned error: %v", err)
	}
	defer db.Close()

	repo := NewPostgresVaultRepository(db)
	if repo == nil {
		t.Fatal("expected repo")
	}
}

func TestPostgresVaultRepository_CreateItem_Success(t *testing.T) {
	t.Parallel()

	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("pgxmock.NewPool returned error: %v", err)
	}
	defer db.Close()

	repo := NewPostgresVaultRepository(db)
	now := time.Now().UTC().Round(0)
	item := &model.VaultItem{
		ID:               testItemID,
		UserID:           testUserID,
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

	db.ExpectExec("INSERT INTO vault_items").
		WithArgs(
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
		).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	if err := repo.CreateItem(context.Background(), item); err != nil {
		t.Fatalf("CreateItem returned error: %v", err)
	}

	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestPostgresVaultRepository_CreateItem_Error(t *testing.T) {
	t.Parallel()

	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("pgxmock.NewPool returned error: %v", err)
	}
	defer db.Close()

	repo := NewPostgresVaultRepository(db)
	item := &model.VaultItem{
		ID:               testItemID,
		UserID:           testUserID,
		Type:             model.ItemTypeText,
		Title:            "note1",
		PayloadEncrypted: []byte("cipher"),
		Version:          1,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	db.ExpectExec("INSERT INTO vault_items").
		WillReturnError(errors.New("insert error"))

	if err := repo.CreateItem(context.Background(), item); err == nil {
		t.Fatal("expected error")
	}
}

func TestPostgresVaultRepository_GetItemByID_Success(t *testing.T) {
	t.Parallel()

	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("pgxmock.NewPool returned error: %v", err)
	}
	defer db.Close()

	repo := NewPostgresVaultRepository(db)
	now := time.Now().UTC().Round(0)
	rows := pgxmock.NewRows(vaultItemColumns).AddRow(
		testItemID, testUserID, string(model.ItemTypeText), "note1", "meta1",
		[]byte("cipher"), []byte("nonce"), "hash1",
		int64(1), nil, now, now, nil,
	)

	db.ExpectQuery("SELECT").
		WithArgs(testItemID.String(), testUserID).
		WillReturnRows(rows)

	item, err := repo.GetItemByID(context.Background(), testUserID, testItemID.String())
	if err != nil {
		t.Fatalf("GetItemByID returned error: %v", err)
	}
	if item == nil {
		t.Fatal("expected item")
	}
	if item.ID != testItemID {
		t.Fatalf("expected %s, got %s", testItemID, item.ID)
	}
}

func TestPostgresVaultRepository_GetItemByID_NotFound(t *testing.T) {
	t.Parallel()

	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("pgxmock.NewPool returned error: %v", err)
	}
	defer db.Close()

	repo := NewPostgresVaultRepository(db)

	db.ExpectQuery("SELECT").
		WithArgs(testItemID.String(), testUserID).
		WillReturnError(pgx.ErrNoRows)

	item, err := repo.GetItemByID(context.Background(), testUserID, testItemID.String())
	if err != nil {
		t.Fatalf("GetItemByID returned error: %v", err)
	}
	if item != nil {
		t.Fatal("expected nil item")
	}
}

func TestPostgresVaultRepository_ListItems_Success(t *testing.T) {
	t.Parallel()

	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("pgxmock.NewPool returned error: %v", err)
	}
	defer db.Close()

	repo := NewPostgresVaultRepository(db)
	now := time.Now().UTC().Round(0)
	rows := pgxmock.NewRows(vaultItemColumns).AddRow(
		testItemID, testUserID, string(model.ItemTypeText), "note1", "meta1",
		[]byte("cipher"), []byte("nonce"), "hash1",
		int64(1), nil, now, now, nil,
	)

	db.ExpectQuery("SELECT").
		WithArgs(testUserID).
		WillReturnRows(rows)

	items, err := repo.ListItems(context.Background(), testUserID, false, 10, 0)
	if err != nil {
		t.Fatalf("ListItems returned error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
}

func TestPostgresVaultRepository_UpdateItem_Success(t *testing.T) {
	t.Parallel()

	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("pgxmock.NewPool returned error: %v", err)
	}
	defer db.Close()

	repo := NewPostgresVaultRepository(db)
	now := time.Now().UTC().Round(0)
	item := &model.VaultItem{
		ID:               testItemID,
		UserID:           testUserID,
		Title:            "updated",
		Meta:             "meta2",
		PayloadEncrypted: []byte("cipher2"),
		PayloadNonce:     []byte("nonce2"),
		PayloadHash:      "hash2",
		Version:          2,
		ClientUpdatedAt:  &now,
		UpdatedAt:        now,
	}

	db.ExpectExec("UPDATE vault_items").
		WithArgs(
			item.Title,
			item.Meta,
			item.PayloadEncrypted,
			item.PayloadNonce,
			item.PayloadHash,
			item.Version,
			item.ClientUpdatedAt,
			item.UpdatedAt,
			item.ID.String(),
			item.UserID,
		).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

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

	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("pgxmock.NewPool returned error: %v", err)
	}
	defer db.Close()

	repo := NewPostgresVaultRepository(db)
	now := time.Now().UTC().Round(0)

	db.ExpectExec("UPDATE vault_items").
		WithArgs(now, now, testItemID.String(), testUserID).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	ok, err := repo.SoftDeleteItem(context.Background(), testUserID, testItemID.String(), now)
	if err != nil {
		t.Fatalf("SoftDeleteItem returned error: %v", err)
	}
	if !ok {
		t.Fatal("expected deleted=true")
	}
}

func TestPostgresVaultRepository_SyncItems_Success(t *testing.T) {
	t.Parallel()

	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("pgxmock.NewPool returned error: %v", err)
	}
	defer db.Close()

	repo := NewPostgresVaultRepository(db)
	now := time.Now().UTC().Round(0)
	rows := pgxmock.NewRows(vaultItemColumns).AddRow(
		testItemID, testUserID, string(model.ItemTypeText), "note1", "meta1",
		[]byte("cipher"), []byte("nonce"), "hash1",
		int64(1), nil, now, now, nil,
	)

	since := now.Add(-time.Hour)
	db.ExpectQuery("SELECT").
		WithArgs(testUserID, since).
		WillReturnRows(rows)

	items, err := repo.SyncItems(context.Background(), testUserID, since, true)
	if err != nil {
		t.Fatalf("SyncItems returned error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
}
