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

type CreateBankCardItemUseCase struct {
	store *local.Store
	grpc  *grpcclient.Client
}

func NewCreateBankCardItemUseCase(store *local.Store, grpc *grpcclient.Client) *CreateBankCardItemUseCase {
	return &CreateBankCardItemUseCase{
		store: store,
		grpc:  grpc,
	}
}

func (u *CreateBankCardItemUseCase) Execute(
	ctx context.Context,
	title, number, holder, expiry, cvv, meta, masterPassword string,
) (string, error) {

	state, err := u.store.Load()
	if err != nil {
		return "", err
	}
	if state.Token == "" {
		return "", fmt.Errorf("not authenticated")
	}

	if title == "" || number == "" || holder == "" || expiry == "" || cvv == "" {
		return "", fmt.Errorf("all card fields are required")
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

	rawPayload, err := json.Marshal(payload.BankCard{
		Number: number,
		Holder: holder,
		Expiry: expiry,
		CVV:    cvv,
	})
	if err != nil {
		return "", fmt.Errorf("marshal bank card: %w", err)
	}

	ciphertext, nonce, err := clientcrypto.Encrypt(key, rawPayload)
	if err != nil {
		return "", err
	}

	hash := clientcrypto.Hash(ciphertext)

	authCtx := grpcclient.ContextWithBearerToken(ctx, state.Token)

	req := pb.CreateItemRequest_builder{
		Type:             pb.ItemType_ITEM_TYPE_BANK_CARD.Enum(),
		Title:            &title,
		Meta:             &meta,
		PayloadEncrypted: ciphertext,
		PayloadNonce:     nonce,
		PayloadHash:      &hash,
	}.Build()

	resp, err := u.grpc.Raw().CreateItem(authCtx, req)
	if err != nil {
		return "", fmt.Errorf("create bank card: %w", err)
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
