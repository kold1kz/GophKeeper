package repository

import (
	"context"
	"time"

	"gophkeeper/internal/model"
)

// VaultRepository определяет контракт для работы с хранилищем пользователя.
//
// Позволяет создавать, получать, обновлять, удалять и синхронизировать данные.
type VaultRepository interface {
	CreateItem(ctx context.Context, item *model.VaultItem) error
	GetItemByID(ctx context.Context, userID string, itemID string) (*model.VaultItem, error)
	ListItems(ctx context.Context, userID string, includeDeleted bool, limit, offset int) ([]*model.VaultItem, error)
	UpdateItem(ctx context.Context, item *model.VaultItem) (bool, error)
	SoftDeleteItem(ctx context.Context, userID string, itemID string, deletedAt time.Time) (bool, error)
	SyncItems(ctx context.Context, userID string, since time.Time, includeDeleted bool) ([]*model.VaultItem, error)
}
