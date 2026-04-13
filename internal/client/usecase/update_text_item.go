package usecase

import (
	"context"
	"fmt"

	clientcrypto "gophkeeper/internal/client/crypto"
	grpcclient "gophkeeper/internal/client/grpc"
	"gophkeeper/internal/client/local"
	pb "gophkeeper/proto"
)

type UpdateTextItemUseCase struct {
	store *local.Store
	grpc  *grpcclient.Client
}

func NewUpdateTextItemUseCase(store *local.Store, grpc *grpcclient.Client) *UpdateTextItemUseCase {
	return &UpdateTextItemUseCase{
		store: store,
		grpc:  grpc,
	}
}

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
