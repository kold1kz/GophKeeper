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
	GetItem(ctx context.Context, userID string, itemID string) (*model.VaultItem, error)
	ListItems(ctx context.Context, userID string, includeDeleted bool, limit, offset int) ([]*model.VaultItem, error)
	UpdateItem(ctx context.Context, item *model.VaultItem) (*model.VaultItem, error)
	DeleteItem(ctx context.Context, userID string, itemID string) (time.Time, error)
	SyncItems(ctx context.Context, userID string, since time.Time, includeDeleted bool) ([]*model.VaultItem, error)
}

// VaultService предоставляет бизнес-логику работы с хранилищем пользовательских данных.
//
// Сервис отвечает за создание, получение, обновление, удаление
// и синхронизацию элементов хранилища.
type vaultService struct {
	repo repository.VaultRepository
}

// NewVaultService создает сервис для работы с пользовательским хранилищем.
func NewVaultService(repo repository.VaultRepository) VaultService {
	return &vaultService{repo: repo}
}

// CreateItem создает новый элемент в хранилище.
//
// Выполняет базовую валидацию данных перед сохранением.
// Возвращает ErrInvalidItemData, если данные элемента некорректны.
func (s *vaultService) CreateItem(ctx context.Context, item *model.VaultItem) (*model.VaultItem, error) {
	if item == nil {
		return nil, ErrInvalidItemData
	}
	if item.UserID == "" || item.Type == "" || len(item.PayloadEncrypted) == 0 {
		return nil, ErrInvalidItemData
	}

	item.Title = strings.TrimSpace(item.Title)
	if item.Title == "" {
		return nil, ErrInvalidItemData
	}

	now := time.Now()
	item.ID = uuid.New()
	item.Version = 1
	item.CreatedAt = now
	item.UpdatedAt = now

	if err := s.repo.CreateItem(ctx, item); err != nil {
		return nil, err
	}

	return item, nil
}

// GetItem возвращает элемент по идентификатору и идентификатору пользователя.
//
// Возвращает ErrInvalidItemData, если входные параметры некорректны.
// Возвращает ErrItemNotFound, если элемент не найден.
// Возвращает ErrItemDeleted, если элемент был удален.
func (s *vaultService) GetItem(ctx context.Context, userID string, itemID string) (*model.VaultItem, error) {
	if userID == "" || itemID == "" {
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

// ListItems возвращает список элементов пользователя.
//
// Поддерживает фильтрацию удаленных элементов и пагинацию.
// Возвращает ErrInvalidItemData, если параметры запроса некорректны.
func (s *vaultService) ListItems(ctx context.Context, userID string, includeDeleted bool, limit, offset int) ([]*model.VaultItem, error) {
	if userID == "" {
		return nil, ErrInvalidItemData
	}
	if limit < 0 || offset < 0 {
		return nil, ErrInvalidItemData
	}

	return s.repo.ListItems(ctx, userID, includeDeleted, limit, offset)
}

// UpdateItem обновляет существующий элемент в хранилище.
//
// Использует версию элемента для предотвращения конфликтов записи.
// Возвращает:
// - ErrInvalidItemData при некорректных данных,
// - ErrItemNotFound если элемент не найден,
// - ErrItemDeleted если элемент удален,
// - ErrVersionConflict при конфликте версий.
func (s *vaultService) UpdateItem(ctx context.Context, item *model.VaultItem) (*model.VaultItem, error) {
	if item == nil {
		return nil, ErrInvalidItemData
	}
	if item.UserID == "" || item.ID == uuid.Nil || len(item.PayloadEncrypted) == 0 {
		return nil, ErrInvalidItemData
	}

	item.Title = strings.TrimSpace(item.Title)
	if item.Title == "" {
		return nil, ErrInvalidItemData
	}

	current, err := s.repo.GetItemByID(ctx, item.UserID, item.ID.String())
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

// DeleteItem выполняет мягкое удаление элемента пользователя.
//
// Возвращает время удаления элемента.
// Возвращает ErrInvalidItemData при некорректных параметрах.
// Возвращает ErrItemNotFound, если элемент не найден.
func (s *vaultService) DeleteItem(ctx context.Context, userID string, itemID string) (time.Time, error) {
	if userID == "" || itemID == "" {
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

// SyncItems возвращает список элементов, измененных после указанного времени.
//
// Используется для синхронизации данных между несколькими клиентами пользователя.
// Возвращает ErrInvalidItemData при некорректных параметрах запроса.
func (s *vaultService) SyncItems(ctx context.Context, userID string, since time.Time, includeDeleted bool) ([]*model.VaultItem, error) {
	if userID == "" {
		return nil, ErrInvalidItemData
	}

	return s.repo.SyncItems(ctx, userID, since, includeDeleted)
}
