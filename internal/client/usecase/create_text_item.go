package usecase

import (
	"context"
	"fmt"

	clientcrypto "gophkeeper/internal/client/crypto"
	grpcclient "gophkeeper/internal/client/grpc"
	"gophkeeper/internal/client/local"
	pb "gophkeeper/proto"
)

type CreateTextItemUseCase struct {
	store *local.Store
	grpc  *grpcclient.Client
}

func NewCreateTextItemUseCase(store *local.Store, grpc *grpcclient.Client) *CreateTextItemUseCase {
	return &CreateTextItemUseCase{
		store: store,
		grpc:  grpc,
	}
}

func (u *CreateTextItemUseCase) Execute(ctx context.Context, title, text, meta, masterPassword string) (string, error) {
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
	if text == "" {
		return "", fmt.Errorf("text is required")
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

	ciphertext, nonce, err := clientcrypto.Encrypt(key, []byte(text))
	if err != nil {
		return "", err
	}

	hash := clientcrypto.Hash(ciphertext)

	authCtx := grpcclient.ContextWithBearerToken(ctx, state.Token)

	req := pb.CreateItemRequest_builder{
		Type:             pb.ItemType_ITEM_TYPE_TEXT.Enum(),
		Title:            &title,
		Meta:             &meta,
		PayloadEncrypted: ciphertext,
		PayloadNonce:     nonce,
		PayloadHash:      &hash,
		ClientUpdatedAt:  nil,
	}.Build()

	resp, err := u.grpc.Raw().CreateItem(authCtx, req)
	if err != nil {
		return "", fmt.Errorf("create text item: %w", err)
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
