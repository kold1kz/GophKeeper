// Package grpcserver содержит реализацию gRPC-сервера,
// обеспечивающего доступ к хранилищу данных пользователей.
//
// Пакет реализует:
// - обработку gRPC-запросов
// - авторизацию пользователей
// - преобразование моделей в protobuf
// - взаимодействие с бизнес-логикой (service layer)
package grpcserver

import (
	"context"
	"errors"
	"time"

	"gophkeeper/internal/service"
	pb "gophkeeper/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// CreateItem создает новый элемент в хранилище пользователя.
//
// Требует наличия userID в контексте.
// Выполняет преобразование запроса в модель,
// вызывает сервисный слой и возвращает результат.
//
// Возможные ошибки:
// - Unauthenticated — пользователь не авторизован
// - InvalidArgument — некорректный тип или данные
// - Internal — внутренняя ошибка сервера
func (s *Server) CreateItem(ctx context.Context, req *pb.CreateItemRequest) (*pb.CreateItemResponse, error) {
	userID, ok := userIDInt64FromContext(ctx)
	if !ok || userID == 0 {
		return nil, status.Error(codes.Unauthenticated, "missing user in context")
	}

	item, err := createItemRequestToModel(userID, req)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid item type")
	}

	created, err := s.vaultService.CreateItem(ctx, item)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidItemData):
			return nil, status.Error(codes.InvalidArgument, "invalid item data")
		default:
			return nil, status.Error(codes.Internal, "create item failed")
		}
	}

	return pb.CreateItemResponse_builder{
		Item: modelItemToProto(created),
	}.Build(), nil
}

// GetItem возвращает элемент по его идентификатору.
//
// Требует авторизации пользователя.
// Делегирует получение данных сервисному слою.
//
// Возможные ошибки:
// - Unauthenticated — пользователь не авторизован
// - InvalidArgument — некорректный идентификатор
// - NotFound — элемент не найден
// - FailedPrecondition — элемент удален
// - Internal — внутренняя ошибка
func (s *Server) GetItem(ctx context.Context, req *pb.GetItemRequest) (*pb.GetItemResponse, error) {
	userID, ok := userIDInt64FromContext(ctx)
	if !ok || userID == 0 {
		return nil, status.Error(codes.Unauthenticated, "missing user in context")
	}

	item, err := s.vaultService.GetItem(ctx, userID, req.GetId())
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidItemData):
			return nil, status.Error(codes.InvalidArgument, "invalid item id")
		case errors.Is(err, service.ErrItemNotFound):
			return nil, status.Error(codes.NotFound, "item not found")
		case errors.Is(err, service.ErrItemDeleted):
			return nil, status.Error(codes.FailedPrecondition, "item deleted")
		default:
			return nil, status.Error(codes.Internal, "get item failed")
		}
	}

	return pb.GetItemResponse_builder{
		Item: modelItemToProto(item),
	}.Build(), nil
}

// ListItems возвращает список элементов пользователя.
//
// Поддерживает пагинацию и фильтрацию удаленных элементов.
//
// Возможные ошибки:
// - Unauthenticated — пользователь не авторизован
// - InvalidArgument — некорректные параметры
// - Internal — ошибка получения данных
func (s *Server) ListItems(ctx context.Context, req *pb.ListItemsRequest) (*pb.ListItemsResponse, error) {
	userID, ok := userIDInt64FromContext(ctx)
	if !ok || userID == 0 {
		return nil, status.Error(codes.Unauthenticated, "missing user in context")
	}

	items, err := s.vaultService.ListItems(
		ctx,
		userID,
		req.GetIncludeDeleted(),
		int(req.GetLimit()),
		int(req.GetOffset()),
	)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidItemData):
			return nil, status.Error(codes.InvalidArgument, "invalid list params")
		default:
			return nil, status.Error(codes.Internal, "list items failed")
		}
	}

	return pb.ListItemsResponse_builder{
		Items: modelItemsToProto(items),
	}.Build(), nil
}

// UpdateItem обновляет существующий элемент.
//
// Использует версионирование для предотвращения конфликтов.
//
// Возможные ошибки:
// - Unauthenticated — пользователь не авторизован
// - InvalidArgument — некорректные данные
// - NotFound — элемент не найден
// - FailedPrecondition — элемент удален
// - Aborted — конфликт версий
// - Internal — ошибка обновления
func (s *Server) UpdateItem(ctx context.Context, req *pb.UpdateItemRequest) (*pb.UpdateItemResponse, error) {
	userID, ok := userIDInt64FromContext(ctx)
	if !ok || userID == 0 {
		return nil, status.Error(codes.Unauthenticated, "missing user in context")
	}

	item := updateItemRequestToModel(userID, req)

	updated, err := s.vaultService.UpdateItem(ctx, item)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidItemData):
			return nil, status.Error(codes.InvalidArgument, "invalid update data")
		case errors.Is(err, service.ErrItemNotFound):
			return nil, status.Error(codes.NotFound, "item not found")
		case errors.Is(err, service.ErrItemDeleted):
			return nil, status.Error(codes.FailedPrecondition, "item deleted")
		case errors.Is(err, service.ErrVersionConflict):
			return nil, status.Error(codes.Aborted, "version conflict")
		default:
			return nil, status.Error(codes.Internal, "update item failed")
		}
	}

	return pb.UpdateItemResponse_builder{
		Item: modelItemToProto(updated),
	}.Build(), nil
}

// DeleteItem выполняет мягкое удаление элемента.
//
// Возвращает время удаления.
//
// Возможные ошибки:
// - Unauthenticated — пользователь не авторизован
// - InvalidArgument — некорректный идентификатор
// - NotFound — элемент не найден
// - Internal — ошибка удаления
func (s *Server) DeleteItem(ctx context.Context, req *pb.DeleteItemRequest) (*pb.DeleteItemResponse, error) {
	userID, ok := userIDInt64FromContext(ctx)
	if !ok || userID == 0 {
		return nil, status.Error(codes.Unauthenticated, "missing user in context")
	}

	deletedAt, err := s.vaultService.DeleteItem(ctx, userID, req.GetId())
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidItemData):
			return nil, status.Error(codes.InvalidArgument, "invalid item id")
		case errors.Is(err, service.ErrItemNotFound):
			return nil, status.Error(codes.NotFound, "item not found")
		default:
			return nil, status.Error(codes.Internal, "delete item failed")
		}
	}

	id := req.GetId()
	return pb.DeleteItemResponse_builder{
		Id:        &id,
		DeletedAt: timestamppb.New(deletedAt),
	}.Build(), nil
}

// SyncItems синхронизирует данные между клиентом и сервером.
//
// Возвращает список изменений с момента последней синхронизации.
//
// Возможные ошибки:
// - Unauthenticated — пользователь не авторизован
// - InvalidArgument — некорректные параметры
// - Internal — ошибка синхронизации
func (s *Server) SyncItems(ctx context.Context, req *pb.SyncItemsRequest) (*pb.SyncItemsResponse, error) {
	userID, ok := userIDInt64FromContext(ctx)
	if !ok || userID == 0 {
		return nil, status.Error(codes.Unauthenticated, "missing user in context")
	}

	var since time.Time
	if ts := req.GetLastSyncAt(); ts != nil {
		since = ts.AsTime()
	}

	items, err := s.vaultService.SyncItems(ctx, userID, since, req.GetIncludeDeleted())
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidItemData):
			return nil, status.Error(codes.InvalidArgument, "invalid sync params")
		default:
			return nil, status.Error(codes.Internal, "sync items failed")
		}
	}

	now := time.Now()
	return pb.SyncItemsResponse_builder{
		Items:      modelItemsToProto(items),
		ServerTime: timestamppb.New(now),
	}.Build(), nil
}
