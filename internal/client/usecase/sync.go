package usecase

import (
	"context"
	"fmt"
	"time"

	grpcclient "gophkeeper/internal/client/grpc"
	"gophkeeper/internal/client/local"
	pb "gophkeeper/proto"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type SyncUseCase struct {
	store *local.Store
	grpc  *grpcclient.Client
}

func NewSyncUseCase(store *local.Store, grpc *grpcclient.Client) *SyncUseCase {
	return &SyncUseCase{
		store: store,
		grpc:  grpc,
	}
}

func (u *SyncUseCase) Execute(ctx context.Context) error {
	state, err := u.store.Load()
	if err != nil {
		return err
	}
	if state.Token == "" {
		return fmt.Errorf("not authenticated")
	}

	authCtx := grpcclient.ContextWithBearerToken(ctx, state.Token)

	var lastSyncAt *timestamppb.Timestamp
	if state.LastSyncAt != nil {
		lastSyncAt = timestamppb.New(*state.LastSyncAt)
	}

	includeDeleted := true

	req := pb.SyncItemsRequest_builder{
		LastSyncAt:     lastSyncAt,
		IncludeDeleted: &includeDeleted,
	}.Build()

	resp, err := u.grpc.Raw().SyncItems(authCtx, req)
	if err != nil {
		return fmt.Errorf("sync items: %w", err)
	}

	localItems := make([]local.LocalItem, 0, len(resp.GetItems()))
	for _, item := range resp.GetItems() {
		var clientUpdatedAt *time.Time
		if ts := item.GetClientUpdatedAt(); ts != nil {
			t := ts.AsTime()
			clientUpdatedAt = &t
		}

		var deletedAt *time.Time
		if ts := item.GetDeletedAt(); ts != nil {
			t := ts.AsTime()
			deletedAt = &t
		}

		localItems = append(localItems, local.LocalItem{
			ID:               item.GetId(),
			Type:             item.GetType().String(),
			Title:            item.GetTitle(),
			Meta:             item.GetMeta(),
			PayloadEncrypted: item.GetPayloadEncrypted(),
			PayloadNonce:     item.GetPayloadNonce(),
			PayloadHash:      item.GetPayloadHash(),
			Version:          item.GetVersion(),
			ClientUpdatedAt:  clientUpdatedAt,
			CreatedAt:        item.GetCreatedAt().AsTime(),
			UpdatedAt:        item.GetUpdatedAt().AsTime(),
			DeletedAt:        deletedAt,
		})
	}

	serverTime := resp.GetServerTime().AsTime()
	local.ApplySyncedItems(state, localItems, serverTime)

	if err := u.store.Save(state); err != nil {
		return err
	}

	return nil
}
