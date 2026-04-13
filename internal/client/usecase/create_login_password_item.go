// Package usecase содержит прикладные сценарии работы CLI-клиента GophKeeper.
package usecase

import (
	"context"
	"encoding/json"
	"fmt"

	clientcrypto "gophkeeper/internal/client/crypto"
	grpcclient "gophkeeper/internal/client/grpc"
	"gophkeeper/internal/client/local"
	"gophkeeper/internal/client/payload"
	pb "gophkeeper/proto"
)

// CreateLoginPasswordItemUseCase описывает сценарий создания записи.
type CreateLoginPasswordItemUseCase struct {
	store *local.Store
	grpc  *grpcclient.Client
}

// NewCreateLoginPasswordItemUseCase создает новый use case
// для создания записи с логином и паролем.
func NewCreateLoginPasswordItemUseCase(store *local.Store, grpc *grpcclient.Client) *CreateLoginPasswordItemUseCase {
	return &CreateLoginPasswordItemUseCase{
		store: store,
		grpc:  grpc,
	}
}

// Execute создает новую запись с логином и паролем.
//
// Полезная нагрузка сериализуется в JSON, шифруется мастер-паролем
// и отправляется на сервер.
//
// Возвращает идентификатор созданной записи или ошибку.
func (u *CreateLoginPasswordItemUseCase) Execute(
	ctx context.Context,
	title, login, password, meta, masterPassword string,
) (string, error) {
	state, err := u.store.Load()
	if err != nil {
		return "", err
	}
	if state.Token == "" {
		return "", fmt.Errorf("not authenticated")
	}
	if title == "" {
		return "", fmt.Errorf("title is required")
	}
	if login == "" {
		return "", fmt.Errorf("login is required")
	}
	if password == "" {
		return "", fmt.Errorf("password is required")
	}
	if masterPassword == "" {
		return "", fmt.Errorf("master password is required")
	}

	salt, err := ensureKeySalt(state)
	if err != nil {
		return "", err
	}

	key, err := clientcrypto.DeriveKey(masterPassword, salt)
	if err != nil {
		return "", err
	}

	rawPayload, err := json.Marshal(payload.LoginPassword{
		Login:    login,
		Password: password,
	})
	if err != nil {
		return "", fmt.Errorf("marshal login/password payload: %w", err)
	}

	ciphertext, nonce, err := clientcrypto.Encrypt(key, rawPayload)
	if err != nil {
		return "", err
	}

	hash := clientcrypto.Hash(ciphertext)

	authCtx := grpcclient.ContextWithBearerToken(ctx, state.Token)

	req := pb.CreateItemRequest_builder{
		Type:             pb.ItemType_ITEM_TYPE_LOGIN_PASSWORD.Enum(),
		Title:            &title,
		Meta:             &meta,
		PayloadEncrypted: ciphertext,
		PayloadNonce:     nonce,
		PayloadHash:      &hash,
		ClientUpdatedAt:  nil,
	}.Build()

	resp, err := u.grpc.Raw().CreateItem(authCtx, req)
	if err != nil {
		return "", fmt.Errorf("create login/password item: %w", err)
	}

	if err := u.store.Save(state); err != nil {
		return "", err
	}

	item := resp.GetItem()
	if item == nil {
		return "", fmt.Errorf("empty item in response")
	}

	return item.GetId(), nil
}
