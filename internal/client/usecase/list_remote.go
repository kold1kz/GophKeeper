// Package usecase содержит прикладные сценарии работы CLI-клиента GophKeeper.
package usecase

import (
	"context"
	"fmt"

	grpcclient "gophkeeper/internal/client/grpc"
	"gophkeeper/internal/client/local"
	pb "gophkeeper/proto"
)

// ListRemoteItemsUseCase описывает сценарий вывода удалённых записей с сервера.
type ListRemoteItemsUseCase struct {
	store *local.Store
	grpc  *grpcclient.Client
}

// NewListRemoteItemsUseCase создаёт новый use case для вывода удалённых записей.
func NewListRemoteItemsUseCase(store *local.Store, grpc *grpcclient.Client) *ListRemoteItemsUseCase {
	return &ListRemoteItemsUseCase{
		store: store,
		grpc:  grpc,
	}
}

// Execute выполняет запрос списка записей на сервер и выводит их в консоль.
//
// Для выполнения операции требуется, чтобы пользователь был аутентифицирован.
// Если сервер не возвращает записей, функция печатает сообщение
// "no remote items".
func (u *ListRemoteItemsUseCase) Execute(ctx context.Context) error {
	state, err := u.store.Load()
	if err != nil {
		return err
	}
	if state.Token == "" {
		return fmt.Errorf("not authenticated")
	}

	authCtx := grpcclient.ContextWithBearerToken(ctx, state.Token)

	includeDeleted := false
	limit := int32(100)
	offset := int32(0)

	req := pb.ListItemsRequest_builder{
		IncludeDeleted: &includeDeleted,
		Limit:          &limit,
		Offset:         &offset,
	}.Build()

	resp, err := u.grpc.Raw().ListItems(authCtx, req)
	if err != nil {
		return fmt.Errorf("list remote items: %w", err)
	}

	items := resp.GetItems()
	if len(items) == 0 {
		fmt.Println("no remote items")
		return nil
	}

	for _, item := range items {
		fmt.Println("ID:", item.GetId())
		fmt.Println("Type:", item.GetType().String())
		fmt.Println("Title:", item.GetTitle())
		fmt.Println("Meta:", item.GetMeta())
		fmt.Println("Version:", item.GetVersion())
		fmt.Println("Updated:", item.GetUpdatedAt().AsTime())
		fmt.Println("---")
	}

	return nil
}
