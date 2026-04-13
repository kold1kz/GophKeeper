package service

import (
	"context"
	"strings"
	"time"

	"gophkeeper/internal/model"
	"gophkeeper/internal/repository"

	"github.com/google/uuid"
)

type VaultService interface {
	CreateItem(ctx context.Context, item *model.VaultItem) (*model.VaultItem, error)
	GetItem(ctx context.Context, userID int64, itemID string) (*model.VaultItem, error)
	ListItems(ctx context.Context, userID int64, includeDeleted bool, limit, offset int) ([]*model.VaultItem, error)
	UpdateItem(ctx context.Context, item *model.VaultItem) (*model.VaultItem, error)
	DeleteItem(ctx context.Context, userID int64, itemID string) (time.Time, error)
	SyncItems(ctx context.Context, userID int64, since time.Time, includeDeleted bool) ([]*model.VaultItem, error)
}

type vaultService struct {
	repo repository.VaultRepository
}

func NewVaultService(repo repository.VaultRepository) VaultService {
	return &vaultService{repo: repo}
}

func (s *vaultService) CreateItem(ctx context.Context, item *model.VaultItem) (*model.VaultItem, error) {
	if item == nil {
		return nil, ErrInvalidItemData
	}
	if item.UserID == 0 || item.Type == "" || len(item.PayloadEncrypted) == 0 {
		return nil, ErrInvalidItemData
	}

	item.Title = strings.TrimSpace(item.Title)
	if item.Title == "" {
		return nil, ErrInvalidItemData
	}

	now := time.Now()
	item.ID = uuid.NewString()
	item.Version = 1
	item.CreatedAt = now
	item.UpdatedAt = now

	if err := s.repo.CreateItem(ctx, item); err != nil {
		return nil, err
	}

	return item, nil
}

func (s *vaultService) GetItem(ctx context.Context, userID int64, itemID string) (*model.VaultItem, error) {
	if userID == 0 || itemID == "" {
		return nil, ErrInvalidItemData
	}

	item, err := s.repo.GetItemByID(ctx, userID, itemID)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, ErrItemNotFound
	}
	if item.DeletedAt != nil {
		return nil, ErrItemDeleted
	}

	return item, nil
}

func (s *vaultService) ListItems(ctx context.Context, userID int64, includeDeleted bool, limit, offset int) ([]*model.VaultItem, error) {
	if userID == 0 {
		return nil, ErrInvalidItemData
	}
	if limit < 0 || offset < 0 {
		return nil, ErrInvalidItemData
	}

	return s.repo.ListItems(ctx, userID, includeDeleted, limit, offset)
}

func (s *vaultService) UpdateItem(ctx context.Context, item *model.VaultItem) (*model.VaultItem, error) {
	if item == nil {
		return nil, ErrInvalidItemData
	}
	if item.UserID == 0 || item.ID == "" || len(item.PayloadEncrypted) == 0 {
		return nil, ErrInvalidItemData
	}

	item.Title = strings.TrimSpace(item.Title)
	if item.Title == "" {
		return nil, ErrInvalidItemData
	}

	current, err := s.repo.GetItemByID(ctx, item.UserID, item.ID)
	if err != nil {
		return nil, err
	}
	if current == nil {
		return nil, ErrItemNotFound
	}
	if current.DeletedAt != nil {
		return nil, ErrItemDeleted
	}
	if item.Version != current.Version {
		return nil, ErrVersionConflict
	}

	item.Type = current.Type
	item.CreatedAt = current.CreatedAt
	item.UpdatedAt = time.Now()
	item.Version = current.Version + 1

	ok, err := s.repo.UpdateItem(ctx, item)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrItemNotFound
	}

	item.ClientUpdatedAt = item.ClientUpdatedAt
	item.DeletedAt = current.DeletedAt

	return item, nil
}

func (s *vaultService) DeleteItem(ctx context.Context, userID int64, itemID string) (time.Time, error) {
	if userID == 0 || itemID == "" {
		return time.Time{}, ErrInvalidItemData
	}

	deletedAt := time.Now()
	ok, err := s.repo.SoftDeleteItem(ctx, userID, itemID, deletedAt)
	if err != nil {
		return time.Time{}, err
	}
	if !ok {
		return time.Time{}, ErrItemNotFound
	}

	return deletedAt, nil
}

func (s *vaultService) SyncItems(ctx context.Context, userID int64, since time.Time, includeDeleted bool) ([]*model.VaultItem, error) {
	if userID == 0 {
		return nil, ErrInvalidItemData
	}

	return s.repo.SyncItems(ctx, userID, since, includeDeleted)
}
