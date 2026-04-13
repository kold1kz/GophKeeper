package grpcserver

import (
	"context"
	"testing"
	"time"

	"gophkeeper/internal/model"
	"gophkeeper/internal/service"
	pb "gophkeeper/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestServer_GetItem_Success(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC().Round(0)

	vaultSvc := &vaultServiceMock{
		getFn: func(ctx context.Context, userID int64, itemID string) (*model.VaultItem, error) {
			return &model.VaultItem{
				ID:               itemID,
				UserID:           userID,
				Type:             model.ItemTypeText,
				Title:            "note1",
				Meta:             "meta1",
				PayloadEncrypted: []byte("cipher"),
				PayloadNonce:     []byte("nonce"),
				PayloadHash:      "hash1",
				Version:          1,
				CreatedAt:        now,
				UpdatedAt:        now,
			}, nil
		},
	}
	srv := NewServer(&registerServiceMock{}, &loginServiceMock{}, vaultSvc)

	ctx := withUserID(context.Background(), "1")

	id := "item-1"
	resp, err := srv.GetItem(ctx, pb.GetItemRequest_builder{
		Id: &id,
	}.Build())
	if err != nil {
		t.Fatalf("GetItem returned error: %v", err)
	}
	if resp.GetItem() == nil {
		t.Fatal("expected item")
	}
	if resp.GetItem().GetId() != "item-1" {
		t.Fatalf("expected item-1, got %q", resp.GetItem().GetId())
	}
}

func TestServer_GetItem_NotFound(t *testing.T) {
	t.Parallel()

	vaultSvc := &vaultServiceMock{
		getFn: func(ctx context.Context, userID int64, itemID string) (*model.VaultItem, error) {
			return nil, service.ErrItemNotFound
		},
	}
	srv := NewServer(&registerServiceMock{}, &loginServiceMock{}, vaultSvc)

	ctx := withUserID(context.Background(), "1")
	id := "missing"

	_, err := srv.GetItem(ctx, pb.GetItemRequest_builder{
		Id: &id,
	}.Build())
	if err == nil {
		t.Fatal("expected error")
	}

	st, _ := status.FromError(err)
	if st.Code() != codes.NotFound {
		t.Fatalf("expected NotFound, got %v", st.Code())
	}
}

func TestServer_ListItems_Success(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC().Round(0)

	vaultSvc := &vaultServiceMock{
		listFn: func(ctx context.Context, userID int64, includeDeleted bool, limit, offset int) ([]*model.VaultItem, error) {
			return []*model.VaultItem{
				{
					ID:               "item-1",
					UserID:           userID,
					Type:             model.ItemTypeText,
					Title:            "note1",
					Meta:             "meta1",
					PayloadEncrypted: []byte("cipher"),
					PayloadNonce:     []byte("nonce"),
					PayloadHash:      "hash1",
					Version:          1,
					CreatedAt:        now,
					UpdatedAt:        now,
				},
			}, nil
		},
	}
	srv := NewServer(&registerServiceMock{}, &loginServiceMock{}, vaultSvc)

	ctx := withUserID(context.Background(), "1")
	includeDeleted := false
	limit := int32(10)
	offset := int32(0)

	resp, err := srv.ListItems(ctx, pb.ListItemsRequest_builder{
		IncludeDeleted: &includeDeleted,
		Limit:          &limit,
		Offset:         &offset,
	}.Build())
	if err != nil {
		t.Fatalf("ListItems returned error: %v", err)
	}
	if len(resp.GetItems()) != 1 {
		t.Fatalf("expected 1 item, got %d", len(resp.GetItems()))
	}
}

func TestServer_DeleteItem_Success(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC().Round(0)

	vaultSvc := &vaultServiceMock{
		deleteFn: func(ctx context.Context, userID int64, itemID string) (time.Time, error) {
			return now, nil
		},
	}
	srv := NewServer(&registerServiceMock{}, &loginServiceMock{}, vaultSvc)

	ctx := withUserID(context.Background(), "1")
	id := "item-1"

	resp, err := srv.DeleteItem(ctx, pb.DeleteItemRequest_builder{
		Id: &id,
	}.Build())
	if err != nil {
		t.Fatalf("DeleteItem returned error: %v", err)
	}
	if resp.GetId() != "item-1" {
		t.Fatalf("expected item-1, got %q", resp.GetId())
	}
	if resp.GetDeletedAt() == nil {
		t.Fatal("expected DeletedAt")
	}
}

func TestServer_DeleteItem_NotFound(t *testing.T) {
	t.Parallel()

	vaultSvc := &vaultServiceMock{
		deleteFn: func(ctx context.Context, userID int64, itemID string) (time.Time, error) {
			return time.Time{}, service.ErrItemNotFound
		},
	}
	srv := NewServer(&registerServiceMock{}, &loginServiceMock{}, vaultSvc)

	ctx := withUserID(context.Background(), "1")
	id := "missing"

	_, err := srv.DeleteItem(ctx, pb.DeleteItemRequest_builder{
		Id: &id,
	}.Build())
	if err == nil {
		t.Fatal("expected error")
	}

	st, _ := status.FromError(err)
	if st.Code() != codes.NotFound {
		t.Fatalf("expected NotFound, got %v", st.Code())
	}
}

func TestServer_SyncItems_Success(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC().Round(0)

	vaultSvc := &vaultServiceMock{
		syncFn: func(ctx context.Context, userID int64, since time.Time, includeDeleted bool) ([]*model.VaultItem, error) {
			return []*model.VaultItem{
				{
					ID:               "item-1",
					UserID:           userID,
					Type:             model.ItemTypeText,
					Title:            "note1",
					Meta:             "meta1",
					PayloadEncrypted: []byte("cipher"),
					PayloadNonce:     []byte("nonce"),
					PayloadHash:      "hash1",
					Version:          1,
					CreatedAt:        now,
					UpdatedAt:        now,
				},
			}, nil
		},
	}
	srv := NewServer(&registerServiceMock{}, &loginServiceMock{}, vaultSvc)

	ctx := withUserID(context.Background(), "1")
	includeDeleted := true

	resp, err := srv.SyncItems(ctx, pb.SyncItemsRequest_builder{
		IncludeDeleted: &includeDeleted,
	}.Build())
	if err != nil {
		t.Fatalf("SyncItems returned error: %v", err)
	}
	if len(resp.GetItems()) != 1 {
		t.Fatalf("expected 1 item, got %d", len(resp.GetItems()))
	}
	if resp.GetServerTime() == nil {
		t.Fatal("expected ServerTime")
	}
}

func TestServer_UpdateItem_Success(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC().Round(0)

	vaultSvc := &vaultServiceMock{
		updateFn: func(ctx context.Context, item *model.VaultItem) (*model.VaultItem, error) {
			return &model.VaultItem{
				ID:               item.ID,
				UserID:           item.UserID,
				Type:             model.ItemTypeText,
				Title:            item.Title,
				Meta:             item.Meta,
				PayloadEncrypted: item.PayloadEncrypted,
				PayloadNonce:     item.PayloadNonce,
				PayloadHash:      item.PayloadHash,
				Version:          item.Version + 1,
				CreatedAt:        now.Add(-time.Hour),
				UpdatedAt:        now,
			}, nil
		},
	}
	srv := NewServer(&registerServiceMock{}, &loginServiceMock{}, vaultSvc)

	ctx := withUserID(context.Background(), "1")

	id := "item-1"
	title := "updated-title"
	meta := "updated-meta"
	hash := "hash1"
	version := int64(2)

	resp, err := srv.UpdateItem(ctx, pb.UpdateItemRequest_builder{
		Id:               &id,
		Title:            &title,
		Meta:             &meta,
		PayloadEncrypted: []byte("cipher"),
		PayloadNonce:     []byte("nonce"),
		PayloadHash:      &hash,
		Version:          &version,
	}.Build())
	if err != nil {
		t.Fatalf("UpdateItem returned error: %v", err)
	}

	if resp.GetItem() == nil {
		t.Fatal("expected item")
	}
	if resp.GetItem().GetId() != "item-1" {
		t.Fatalf("expected item-1, got %q", resp.GetItem().GetId())
	}
}

func TestServer_UpdateItem_VersionConflict(t *testing.T) {
	t.Parallel()

	vaultSvc := &vaultServiceMock{
		updateFn: func(ctx context.Context, item *model.VaultItem) (*model.VaultItem, error) {
			return nil, service.ErrVersionConflict
		},
	}
	srv := NewServer(&registerServiceMock{}, &loginServiceMock{}, vaultSvc)

	ctx := withUserID(context.Background(), "1")

	id := "item-1"
	title := "updated-title"
	version := int64(2)

	_, err := srv.UpdateItem(ctx, pb.UpdateItemRequest_builder{
		Id:               &id,
		Title:            &title,
		PayloadEncrypted: []byte("cipher"),
		PayloadNonce:     []byte("nonce"),
		Version:          &version,
	}.Build())
	if err == nil {
		t.Fatal("expected error")
	}

	st, _ := status.FromError(err)
	if st.Code() != codes.Aborted {
		t.Fatalf("expected Aborted, got %v", st.Code())
	}
}
