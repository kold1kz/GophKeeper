package grpcserver

import (
	"context"
	"errors"
	"testing"
	"time"

	"gophkeeper/internal/model"
	"gophkeeper/internal/service"
	pb "gophkeeper/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestServer_CreateItem_Unauthenticated_Negative(t *testing.T) {
	t.Parallel()

	srv := NewServer(&registerServiceMock{}, &loginServiceMock{}, &vaultServiceMock{})

	title := "title"
	meta := "meta"
	hash := "hash"
	badType := pb.ItemType(1)

	_, err := srv.CreateItem(context.Background(), pb.CreateItemRequest_builder{
		Type:             &badType,
		Title:            &title,
		Meta:             &meta,
		PayloadEncrypted: []byte("cipher"),
		PayloadNonce:     []byte("nonce"),
		PayloadHash:      &hash,
	}.Build())
	if err == nil {
		t.Fatal("expected error")
	}

	st, _ := status.FromError(err)
	if st.Code() != codes.Unauthenticated {
		t.Fatalf("expected Unauthenticated, got %v", st.Code())
	}
}

func TestServer_CreateItem_InvalidType(t *testing.T) {
	t.Parallel()

	srv := NewServer(&registerServiceMock{}, &loginServiceMock{}, &vaultServiceMock{})

	ctx := withUserID(context.Background(), "1")
	title := "title"
	meta := "meta"
	hash := "hash"
	badType := pb.ItemType(999)

	_, err := srv.CreateItem(ctx, pb.CreateItemRequest_builder{
		Type:             &badType,
		Title:            &title,
		Meta:             &meta,
		PayloadEncrypted: []byte("cipher"),
		PayloadNonce:     []byte("nonce"),
		PayloadHash:      &hash,
	}.Build())
	if err == nil {
		t.Fatal("expected error")
	}

	st, _ := status.FromError(err)
	if st.Code() != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument, got %v", st.Code())
	}
}

func TestServer_CreateItem_InvalidItemData(t *testing.T) {
	t.Parallel()

	vaultSvc := &vaultServiceMock{
		createFn: func(ctx context.Context, item *model.VaultItem) (*model.VaultItem, error) {
			return nil, service.ErrInvalidItemData
		},
	}
	srv := NewServer(&registerServiceMock{}, &loginServiceMock{}, vaultSvc)

	ctx := withUserID(context.Background(), "1")
	title := "title"
	meta := "meta"
	hash := "hash"
	itemType := pb.ItemType_ITEM_TYPE_TEXT

	_, err := srv.CreateItem(ctx, pb.CreateItemRequest_builder{
		Type:             &itemType,
		Title:            &title,
		Meta:             &meta,
		PayloadEncrypted: []byte("cipher"),
		PayloadNonce:     []byte("nonce"),
		PayloadHash:      &hash,
	}.Build())
	if err == nil {
		t.Fatal("expected error")
	}

	st, _ := status.FromError(err)
	if st.Code() != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument, got %v", st.Code())
	}
}

func TestServer_CreateItem_InternalError(t *testing.T) {
	t.Parallel()

	vaultSvc := &vaultServiceMock{
		createFn: func(ctx context.Context, item *model.VaultItem) (*model.VaultItem, error) {
			return nil, errors.New("boom")
		},
	}
	srv := NewServer(&registerServiceMock{}, &loginServiceMock{}, vaultSvc)

	ctx := withUserID(context.Background(), "1")
	title := "title"
	meta := "meta"
	hash := "hash"
	itemType := pb.ItemType_ITEM_TYPE_TEXT

	_, err := srv.CreateItem(ctx, pb.CreateItemRequest_builder{
		Type:             &itemType,
		Title:            &title,
		Meta:             &meta,
		PayloadEncrypted: []byte("cipher"),
		PayloadNonce:     []byte("nonce"),
		PayloadHash:      &hash,
	}.Build())
	if err == nil {
		t.Fatal("expected error")
	}

	st, _ := status.FromError(err)
	if st.Code() != codes.Internal {
		t.Fatalf("expected Internal, got %v", st.Code())
	}
}

func TestServer_GetItem_Unauthenticated_Negative(t *testing.T) {
	t.Parallel()

	srv := NewServer(&registerServiceMock{}, &loginServiceMock{}, &vaultServiceMock{})

	id := "item-1"
	_, err := srv.GetItem(context.Background(), pb.GetItemRequest_builder{
		Id: &id,
	}.Build())
	if err == nil {
		t.Fatal("expected error")
	}

	st, _ := status.FromError(err)
	if st.Code() != codes.Unauthenticated {
		t.Fatalf("expected Unauthenticated, got %v", st.Code())
	}
}

func TestServer_GetItem_InvalidItemID(t *testing.T) {
	t.Parallel()

	vaultSvc := &vaultServiceMock{
		getFn: func(ctx context.Context, userID string, itemID string) (*model.VaultItem, error) {
			return nil, service.ErrInvalidItemData
		},
	}
	srv := NewServer(&registerServiceMock{}, &loginServiceMock{}, vaultSvc)

	ctx := withUserID(context.Background(), "1")
	id := ""

	_, err := srv.GetItem(ctx, pb.GetItemRequest_builder{
		Id: &id,
	}.Build())
	if err == nil {
		t.Fatal("expected error")
	}

	st, _ := status.FromError(err)
	if st.Code() != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument, got %v", st.Code())
	}
}

func TestServer_GetItem_Deleted(t *testing.T) {
	t.Parallel()

	vaultSvc := &vaultServiceMock{
		getFn: func(ctx context.Context, userID string, itemID string) (*model.VaultItem, error) {
			return nil, service.ErrItemDeleted
		},
	}
	srv := NewServer(&registerServiceMock{}, &loginServiceMock{}, vaultSvc)

	ctx := withUserID(context.Background(), "1")
	id := "item-1"

	_, err := srv.GetItem(ctx, pb.GetItemRequest_builder{
		Id: &id,
	}.Build())
	if err == nil {
		t.Fatal("expected error")
	}

	st, _ := status.FromError(err)
	if st.Code() != codes.FailedPrecondition {
		t.Fatalf("expected FailedPrecondition, got %v", st.Code())
	}
}

func TestServer_GetItem_InternalError(t *testing.T) {
	t.Parallel()

	vaultSvc := &vaultServiceMock{
		getFn: func(ctx context.Context, userID string, itemID string) (*model.VaultItem, error) {
			return nil, errors.New("boom")
		},
	}
	srv := NewServer(&registerServiceMock{}, &loginServiceMock{}, vaultSvc)

	ctx := withUserID(context.Background(), "1")
	id := "item-1"

	_, err := srv.GetItem(ctx, pb.GetItemRequest_builder{
		Id: &id,
	}.Build())
	if err == nil {
		t.Fatal("expected error")
	}

	st, _ := status.FromError(err)
	if st.Code() != codes.Internal {
		t.Fatalf("expected Internal, got %v", st.Code())
	}
}

func TestServer_ListItems_Unauthenticated_Negative(t *testing.T) {
	t.Parallel()

	srv := NewServer(&registerServiceMock{}, &loginServiceMock{}, &vaultServiceMock{})

	_, err := srv.ListItems(context.Background(), pb.ListItemsRequest_builder{}.Build())
	if err == nil {
		t.Fatal("expected error")
	}

	st, _ := status.FromError(err)
	if st.Code() != codes.Unauthenticated {
		t.Fatalf("expected Unauthenticated, got %v", st.Code())
	}
}

func TestServer_ListItems_InvalidParams(t *testing.T) {
	t.Parallel()

	vaultSvc := &vaultServiceMock{
		listFn: func(ctx context.Context, userID string, includeDeleted bool, limit, offset int) ([]*model.VaultItem, error) {
			return nil, service.ErrInvalidItemData
		},
	}
	srv := NewServer(&registerServiceMock{}, &loginServiceMock{}, vaultSvc)

	ctx := withUserID(context.Background(), "1")
	includeDeleted := false
	limit := int32(-1)
	offset := int32(-1)

	_, err := srv.ListItems(ctx, pb.ListItemsRequest_builder{
		IncludeDeleted: &includeDeleted,
		Limit:          &limit,
		Offset:         &offset,
	}.Build())
	if err == nil {
		t.Fatal("expected error")
	}

	st, _ := status.FromError(err)
	if st.Code() != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument, got %v", st.Code())
	}
}

func TestServer_ListItems_InternalError(t *testing.T) {
	t.Parallel()

	vaultSvc := &vaultServiceMock{
		listFn: func(ctx context.Context, userID string, includeDeleted bool, limit, offset int) ([]*model.VaultItem, error) {
			return nil, errors.New("boom")
		},
	}
	srv := NewServer(&registerServiceMock{}, &loginServiceMock{}, vaultSvc)

	ctx := withUserID(context.Background(), "1")
	includeDeleted := false
	limit := int32(10)
	offset := int32(0)

	_, err := srv.ListItems(ctx, pb.ListItemsRequest_builder{
		IncludeDeleted: &includeDeleted,
		Limit:          &limit,
		Offset:         &offset,
	}.Build())
	if err == nil {
		t.Fatal("expected error")
	}

	st, _ := status.FromError(err)
	if st.Code() != codes.Internal {
		t.Fatalf("expected Internal, got %v", st.Code())
	}
}

func TestServer_UpdateItem_Unauthenticated_Negative(t *testing.T) {
	t.Parallel()

	srv := NewServer(&registerServiceMock{}, &loginServiceMock{}, &vaultServiceMock{})

	id := "item-1"
	version := int64(1)

	_, err := srv.UpdateItem(context.Background(), pb.UpdateItemRequest_builder{
		Id:               &id,
		PayloadEncrypted: []byte("cipher"),
		PayloadNonce:     []byte("nonce"),
		Version:          &version,
	}.Build())
	if err == nil {
		t.Fatal("expected error")
	}

	st, _ := status.FromError(err)
	if st.Code() != codes.Unauthenticated {
		t.Fatalf("expected Unauthenticated, got %v", st.Code())
	}
}

func TestServer_UpdateItem_InvalidData(t *testing.T) {
	t.Parallel()

	vaultSvc := &vaultServiceMock{
		updateFn: func(ctx context.Context, item *model.VaultItem) (*model.VaultItem, error) {
			return nil, service.ErrInvalidItemData
		},
	}
	srv := NewServer(&registerServiceMock{}, &loginServiceMock{}, vaultSvc)

	ctx := withUserID(context.Background(), "1")
	id := "item-1"
	version := int64(1)

	_, err := srv.UpdateItem(ctx, pb.UpdateItemRequest_builder{
		Id:               &id,
		PayloadEncrypted: []byte("cipher"),
		PayloadNonce:     []byte("nonce"),
		Version:          &version,
	}.Build())
	if err == nil {
		t.Fatal("expected error")
	}

	st, _ := status.FromError(err)
	if st.Code() != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument, got %v", st.Code())
	}
}

func TestServer_UpdateItem_NotFound(t *testing.T) {
	t.Parallel()

	vaultSvc := &vaultServiceMock{
		updateFn: func(ctx context.Context, item *model.VaultItem) (*model.VaultItem, error) {
			return nil, service.ErrItemNotFound
		},
	}
	srv := NewServer(&registerServiceMock{}, &loginServiceMock{}, vaultSvc)

	ctx := withUserID(context.Background(), "1")
	id := "item-1"
	version := int64(1)

	_, err := srv.UpdateItem(ctx, pb.UpdateItemRequest_builder{
		Id:               &id,
		PayloadEncrypted: []byte("cipher"),
		PayloadNonce:     []byte("nonce"),
		Version:          &version,
	}.Build())
	if err == nil {
		t.Fatal("expected error")
	}

	st, _ := status.FromError(err)
	if st.Code() != codes.NotFound {
		t.Fatalf("expected NotFound, got %v", st.Code())
	}
}

func TestServer_UpdateItem_Deleted(t *testing.T) {
	t.Parallel()

	vaultSvc := &vaultServiceMock{
		updateFn: func(ctx context.Context, item *model.VaultItem) (*model.VaultItem, error) {
			return nil, service.ErrItemDeleted
		},
	}
	srv := NewServer(&registerServiceMock{}, &loginServiceMock{}, vaultSvc)

	ctx := withUserID(context.Background(), "1")
	id := "item-1"
	version := int64(1)

	_, err := srv.UpdateItem(ctx, pb.UpdateItemRequest_builder{
		Id:               &id,
		PayloadEncrypted: []byte("cipher"),
		PayloadNonce:     []byte("nonce"),
		Version:          &version,
	}.Build())
	if err == nil {
		t.Fatal("expected error")
	}

	st, _ := status.FromError(err)
	if st.Code() != codes.FailedPrecondition {
		t.Fatalf("expected FailedPrecondition, got %v", st.Code())
	}
}

func TestServer_UpdateItem_InternalError(t *testing.T) {
	t.Parallel()

	vaultSvc := &vaultServiceMock{
		updateFn: func(ctx context.Context, item *model.VaultItem) (*model.VaultItem, error) {
			return nil, errors.New("boom")
		},
	}
	srv := NewServer(&registerServiceMock{}, &loginServiceMock{}, vaultSvc)

	ctx := withUserID(context.Background(), "1")
	id := "item-1"
	version := int64(1)

	_, err := srv.UpdateItem(ctx, pb.UpdateItemRequest_builder{
		Id:               &id,
		PayloadEncrypted: []byte("cipher"),
		PayloadNonce:     []byte("nonce"),
		Version:          &version,
	}.Build())
	if err == nil {
		t.Fatal("expected error")
	}

	st, _ := status.FromError(err)
	if st.Code() != codes.Internal {
		t.Fatalf("expected Internal, got %v", st.Code())
	}
}

func TestServer_DeleteItem_Unauthenticated_Negative(t *testing.T) {
	t.Parallel()

	srv := NewServer(&registerServiceMock{}, &loginServiceMock{}, &vaultServiceMock{})

	id := "item-1"
	_, err := srv.DeleteItem(context.Background(), pb.DeleteItemRequest_builder{
		Id: &id,
	}.Build())
	if err == nil {
		t.Fatal("expected error")
	}

	st, _ := status.FromError(err)
	if st.Code() != codes.Unauthenticated {
		t.Fatalf("expected Unauthenticated, got %v", st.Code())
	}
}

func TestServer_DeleteItem_InvalidID(t *testing.T) {
	t.Parallel()

	vaultSvc := &vaultServiceMock{
		deleteFn: func(ctx context.Context, userID string, itemID string) (time.Time, error) {
			return time.Time{}, service.ErrInvalidItemData
		},
	}
	srv := NewServer(&registerServiceMock{}, &loginServiceMock{}, vaultSvc)

	ctx := withUserID(context.Background(), "1")
	id := ""

	_, err := srv.DeleteItem(ctx, pb.DeleteItemRequest_builder{
		Id: &id,
	}.Build())
	if err == nil {
		t.Fatal("expected error")
	}

	st, _ := status.FromError(err)
	if st.Code() != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument, got %v", st.Code())
	}
}

func TestServer_DeleteItem_InternalError(t *testing.T) {
	t.Parallel()

	vaultSvc := &vaultServiceMock{
		deleteFn: func(ctx context.Context, userID string, itemID string) (time.Time, error) {
			return time.Time{}, errors.New("boom")
		},
	}
	srv := NewServer(&registerServiceMock{}, &loginServiceMock{}, vaultSvc)

	ctx := withUserID(context.Background(), "1")
	id := "item-1"

	_, err := srv.DeleteItem(ctx, pb.DeleteItemRequest_builder{
		Id: &id,
	}.Build())
	if err == nil {
		t.Fatal("expected error")
	}

	st, _ := status.FromError(err)
	if st.Code() != codes.Internal {
		t.Fatalf("expected Internal, got %v", st.Code())
	}
}

func TestServer_SyncItems_Unauthenticated_Negative(t *testing.T) {
	t.Parallel()

	srv := NewServer(&registerServiceMock{}, &loginServiceMock{}, &vaultServiceMock{})

	_, err := srv.SyncItems(context.Background(), pb.SyncItemsRequest_builder{}.Build())
	if err == nil {
		t.Fatal("expected error")
	}

	st, _ := status.FromError(err)
	if st.Code() != codes.Unauthenticated {
		t.Fatalf("expected Unauthenticated, got %v", st.Code())
	}
}

func TestServer_SyncItems_InvalidParams(t *testing.T) {
	t.Parallel()

	vaultSvc := &vaultServiceMock{
		syncFn: func(ctx context.Context, userID string, since time.Time, includeDeleted bool) ([]*model.VaultItem, error) {
			return nil, service.ErrInvalidItemData
		},
	}
	srv := NewServer(&registerServiceMock{}, &loginServiceMock{}, vaultSvc)

	ctx := withUserID(context.Background(), "1")
	includeDeleted := true
	lastSyncAt := timestamppb.New(time.Now().Add(-time.Hour))

	_, err := srv.SyncItems(ctx, pb.SyncItemsRequest_builder{
		IncludeDeleted: &includeDeleted,
		LastSyncAt:     lastSyncAt,
	}.Build())
	if err == nil {
		t.Fatal("expected error")
	}

	st, _ := status.FromError(err)
	if st.Code() != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument, got %v", st.Code())
	}
}

func TestServer_SyncItems_InternalError(t *testing.T) {
	t.Parallel()

	vaultSvc := &vaultServiceMock{
		syncFn: func(ctx context.Context, userID string, since time.Time, includeDeleted bool) ([]*model.VaultItem, error) {
			return nil, errors.New("boom")
		},
	}
	srv := NewServer(&registerServiceMock{}, &loginServiceMock{}, vaultSvc)

	ctx := withUserID(context.Background(), "1")
	includeDeleted := false
	lastSyncAt := timestamppb.New(time.Now().Add(-time.Hour))

	_, err := srv.SyncItems(ctx, pb.SyncItemsRequest_builder{
		IncludeDeleted: &includeDeleted,
		LastSyncAt:     lastSyncAt,
	}.Build())
	if err == nil {
		t.Fatal("expected error")
	}

	st, _ := status.FromError(err)
	if st.Code() != codes.Internal {
		t.Fatalf("expected Internal, got %v", st.Code())
	}
}
