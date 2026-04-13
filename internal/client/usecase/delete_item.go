package usecase

import (
	"context"
	"fmt"

	grpcclient "gophkeeper/internal/client/grpc"
	"gophkeeper/internal/client/local"
	pb "gophkeeper/proto"
)

type DeleteItemUseCase struct {
	store *local.Store
	grpc  *grpcclient.Client
}

func NewDeleteItemUseCase(store *local.Store, grpc *grpcclient.Client) *DeleteItemUseCase {
	return &DeleteItemUseCase{
		store: store,
		grpc:  grpc,
	}
}

func (u *DeleteItemUseCase) Execute(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("item id is required")
	}

	state, err := u.store.Load()
	if err != nil {
		return err
	}
	if state.Token == "" {
		return fmt.Errorf("not authenticated")
	}

	authCtx := grpcclient.ContextWithBearerToken(ctx, state.Token)

	req := pb.DeleteItemRequest_builder{
		Id: &id,
	}.Build()

	resp, err := u.grpc.Raw().DeleteItem(authCtx, req)
	if err != nil {
		return fmt.Errorf("delete item: %w", err)
	}

	fmt.Println("deleted item:", resp.GetId())
	return nil
}
