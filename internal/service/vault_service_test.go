package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"gophkeeper/internal/model"
)

type vaultRepoMock struct {
	createFn     func(ctx context.Context, item *model.VaultItem) error
	getByIDFn    func(ctx context.Context, userID int64, itemID string) (*model.VaultItem, error)
	listFn       func(ctx context.Context, userID int64, includeDeleted bool, limit, offset int) ([]*model.VaultItem, error)
	updateFn     func(ctx context.Context, item *model.VaultItem) (bool, error)
	softDeleteFn func(ctx context.Context, userID int64, itemID string, deletedAt time.Time) (bool, error)
	syncFn       func(ctx context.Context, userID int64, since time.Time, includeDeleted bool) ([]*model.VaultItem, error)
}

func (m *vaultRepoMock) CreateItem(ctx context.Context, item *model.VaultItem) error {
	return m.createFn(ctx, item)
}

func (m *vaultRepoMock) GetItemByID(ctx context.Context, userID int64, itemID string) (*model.VaultItem, error) {
	return m.getByIDFn(ctx, userID, itemID)
}

func (m *vaultRepoMock) ListItems(ctx context.Context, userID int64, includeDeleted bool, limit, offset int) ([]*model.VaultItem, error) {
	return m.listFn(ctx, userID, includeDeleted, limit, offset)
}

func (m *vaultRepoMock) UpdateItem(ctx context.Context, item *model.VaultItem) (bool, error) {
	return m.updateFn(ctx, item)
}

func (m *vaultRepoMock) SoftDeleteItem(ctx context.Context, userID int64, itemID string, deletedAt time.Time) (bool, error) {
	return m.softDeleteFn(ctx, userID, itemID, deletedAt)
}

func (m *vaultRepoMock) SyncItems(ctx context.Context, userID int64, since time.Time, includeDeleted bool) ([]*model.VaultItem, error) {
	return m.syncFn(ctx, userID, since, includeDeleted)
}

func TestVaultService_CreateItem_Success(t *testing.T) {
	t.Parallel()

	repo := &vaultRepoMock{
		createFn: func(ctx context.Context, item *model.VaultItem) error {
			return nil
		},
	}

	svc := NewVaultService(repo)

	item, err := svc.CreateItem(context.Background(), &model.VaultItem{
		UserID:           1,
		Type:             model.ItemTypeText,
		Title:            "note1",
		PayloadEncrypted: []byte("ciphertext"),
	})
	if err != nil {
		t.Fatalf("CreateItem returned error: %v", err)
	}

	if item.ID == "" {
		t.Fatal("expected generated item ID")
	}
	if item.Version != 1 {
		t.Fatalf("expected version 1, got %d", item.Version)
	}
	if item.CreatedAt.IsZero() {
		t.Fatal("expected CreatedAt to be set")
	}
}

func TestVaultService_CreateItem_InvalidData(t *testing.T) {
	t.Parallel()

	repo := &vaultRepoMock{
		createFn: func(ctx context.Context, item *model.VaultItem) error {
			return nil
		},
	}

	svc := NewVaultService(repo)

	_, err := svc.CreateItem(context.Background(), &model.VaultItem{
		UserID: 0,
		Type:   model.ItemTypeText,
		Title:  "note1",
	})
	if !errors.Is(err, ErrInvalidItemData) {
		t.Fatalf("expected ErrInvalidItemData, got %v", err)
	}
}

func TestVaultService_GetItem_NotFound(t *testing.T) {
	t.Parallel()

	repo := &vaultRepoMock{
		getByIDFn: func(ctx context.Context, userID int64, itemID string) (*model.VaultItem, error) {
			return nil, nil
		},
	}

	svc := NewVaultService(repo)

	_, err := svc.GetItem(context.Background(), 1, "item-id")
	if !errors.Is(err, ErrItemNotFound) {
		t.Fatalf("expected ErrItemNotFound, got %v", err)
	}
}

func TestVaultService_GetItem_Deleted(t *testing.T) {
	t.Parallel()

	now := time.Now()

	repo := &vaultRepoMock{
		getByIDFn: func(ctx context.Context, userID int64, itemID string) (*model.VaultItem, error) {
			return &model.VaultItem{
				ID:        itemID,
				UserID:    userID,
				DeletedAt: &now,
			}, nil
		},
	}

	svc := NewVaultService(repo)

	_, err := svc.GetItem(context.Background(), 1, "item-id")
	if !errors.Is(err, ErrItemDeleted) {
		t.Fatalf("expected ErrItemDeleted, got %v", err)
	}
}

func TestVaultService_UpdateItem_VersionConflict(t *testing.T) {
	t.Parallel()

	repo := &vaultRepoMock{
		getByIDFn: func(ctx context.Context, userID int64, itemID string) (*model.VaultItem, error) {
			return &model.VaultItem{
				ID:               itemID,
				UserID:           userID,
				Type:             model.ItemTypeText,
				Title:            "old",
				PayloadEncrypted: []byte("old"),
				Version:          2,
				CreatedAt:        time.Now(),
			}, nil
		},
		updateFn: func(ctx context.Context, item *model.VaultItem) (bool, error) {
			return true, nil
		},
	}

	svc := NewVaultService(repo)

	_, err := svc.UpdateItem(context.Background(), &model.VaultItem{
		ID:               "item-id",
		UserID:           1,
		Title:            "new",
		PayloadEncrypted: []byte("new"),
		Version:          1,
	})
	if !errors.Is(err, ErrVersionConflict) {
		t.Fatalf("expected ErrVersionConflict, got %v", err)
	}
}

func TestVaultService_DeleteItem_NotFound(t *testing.T) {
	t.Parallel()

	repo := &vaultRepoMock{
		softDeleteFn: func(ctx context.Context, userID int64, itemID string, deletedAt time.Time) (bool, error) {
			return false, nil
		},
	}

	svc := NewVaultService(repo)

	_, err := svc.DeleteItem(context.Background(), 1, "item-id")
	if !errors.Is(err, ErrItemNotFound) {
		t.Fatalf("expected ErrItemNotFound, got %v", err)
	}
}

func TestVaultService_ListItems_InvalidUserID(t *testing.T) {
	t.Parallel()

	repo := &vaultRepoMock{
		listFn: func(ctx context.Context, userID int64, includeDeleted bool, limit, offset int) ([]*model.VaultItem, error) {
			return nil, nil
		},
	}

	svc := NewVaultService(repo)

	_, err := svc.ListItems(context.Background(), 0, false, 10, 0)
	if !errors.Is(err, ErrInvalidItemData) {
		t.Fatalf("expected ErrInvalidItemData, got %v", err)
	}
}

func TestVaultService_SyncItems_InvalidUserID(t *testing.T) {
	t.Parallel()

	repo := &vaultRepoMock{
		syncFn: func(ctx context.Context, userID int64, since time.Time, includeDeleted bool) ([]*model.VaultItem, error) {
			return nil, nil
		},
	}

	svc := NewVaultService(repo)

	_, err := svc.SyncItems(context.Background(), 0, time.Now(), false)
	if !errors.Is(err, ErrInvalidItemData) {
		t.Fatalf("expected ErrInvalidItemData, got %v", err)
	}
}

func TestVaultService_ListItems_Success(t *testing.T) {
	t.Parallel()

	repo := &vaultRepoMock{
		listFn: func(ctx context.Context, userID int64, includeDeleted bool, limit, offset int) ([]*model.VaultItem, error) {
			return []*model.VaultItem{
				{ID: "1", UserID: userID, Title: "note1"},
			}, nil
		},
	}

	svc := NewVaultService(repo)

	items, err := svc.ListItems(context.Background(), 1, false, 10, 0)
	if err != nil {
		t.Fatalf("ListItems returned error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
}

func TestVaultService_UpdateItem_Success(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC().Round(0)

	repo := &vaultRepoMock{
		getByIDFn: func(ctx context.Context, userID int64, itemID string) (*model.VaultItem, error) {
			return &model.VaultItem{
				ID:               itemID,
				UserID:           userID,
				Type:             model.ItemTypeText,
				Title:            "old-title",
				PayloadEncrypted: []byte("old"),
				Version:          3,
				CreatedAt:        now.Add(-time.Hour),
			}, nil
		},
		updateFn: func(ctx context.Context, item *model.VaultItem) (bool, error) {
			return true, nil
		},
	}

	svc := NewVaultService(repo)

	item, err := svc.UpdateItem(context.Background(), &model.VaultItem{
		ID:               "item-1",
		UserID:           1,
		Title:            "new-title",
		Meta:             "meta1",
		PayloadEncrypted: []byte("new"),
		PayloadNonce:     []byte("nonce"),
		PayloadHash:      "hash1",
		Version:          3,
	})
	if err != nil {
		t.Fatalf("UpdateItem returned error: %v", err)
	}
	if item.Version != 4 {
		t.Fatalf("expected version 4, got %d", item.Version)
	}
	if item.Type != model.ItemTypeText {
		t.Fatalf("expected type to be preserved")
	}
}
