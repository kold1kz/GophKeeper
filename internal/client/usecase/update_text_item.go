// Package usecase содержит прикладные сценарии работы CLI-клиента GophKeeper.
package usecase

import (
	"context"
	"fmt"

	clientcrypto "gophkeeper/internal/client/crypto"
	grpcclient "gophkeeper/internal/client/grpc"
	"gophkeeper/internal/client/local"
	pb "gophkeeper/proto"
)

// UpdateTextItemUseCase описывает сценарий обновления текстовой записи.
type UpdateTextItemUseCase struct {
	store *local.Store
	grpc  *grpcclient.Client
}

// NewUpdateTextItemUseCase создаёт новый use case для обновления
// текстовой записи.
func NewUpdateTextItemUseCase(store *local.Store, grpc *grpcclient.Client) *UpdateTextItemUseCase {
	return &UpdateTextItemUseCase{
		store: store,
		grpc:  grpc,
	}
}

// Execute обновляет текстовую запись на сервере.
//
// Функция:
//   - проверяет входные данные;
//   - загружает локальное состояние;
//   - ищет запись в локальном кеше;
//   - шифрует новый текст;
//   - отправляет обновление на сервер.
//
// Для обновления требуется актуальная локальная версия записи,
// поэтому при отсутствии записи в локальном состоянии
// пользователю предлагается сначала выполнить синхронизацию.
func (u *UpdateTextItemUseCase) Execute(ctx context.Context, id, title, text, meta, masterPassword string) error {
	if id == "" {
		return fmt.Errorf("item id is required")
	}
	if title == "" {
		return fmt.Errorf("title is required")
	}
	if text == "" {
		return fmt.Errorf("text is required")
	}
	if masterPassword == "" {
		return fmt.Errorf("master password is required")
	}

	state, err := u.store.Load()
	if err != nil {
		return err
	}
	if state.Token == "" {
		return fmt.Errorf("not authenticated")
	}
	if len(state.KeySalt) == 0 {
		return fmt.Errorf("key salt is missing")
	}

	var localItem *local.LocalItem
	for i := range state.Items {
		if state.Items[i].ID == id {
			localItem = &state.Items[i]
			break
		}
	}
	if localItem == nil {
		return fmt.Errorf("item not found in local state; run sync first")
	}
	if localItem.DeletedAt != nil {
		return fmt.Errorf("item is deleted")
	}

	key, err := clientcrypto.DeriveKey(masterPassword, state.KeySalt)
	if err != nil {
		return err
	}

	ciphertext, nonce, err := clientcrypto.Encrypt(key, []byte(text))
	if err != nil {
		return err
	}

	hash := clientcrypto.Hash(ciphertext)

	authCtx := grpcclient.ContextWithBearerToken(ctx, state.Token)
	version := localItem.Version

	req := pb.UpdateItemRequest_builder{
		Id:               &id,
		Title:            &title,
		Meta:             &meta,
		PayloadEncrypted: ciphertext,
		PayloadNonce:     nonce,
		PayloadHash:      &hash,
		ClientUpdatedAt:  nil,
		Version:          &version,
	}.Build()

	resp, err := u.grpc.Raw().UpdateItem(authCtx, req)
	if err != nil {
		return fmt.Errorf("update item: %w", err)
	}

	item := resp.GetItem()
	if item == nil {
		return fmt.Errorf("empty item in response")
	}

	fmt.Println("updated item:", item.GetId())
	return nil
}
