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
