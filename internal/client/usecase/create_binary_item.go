package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	clientcrypto "gophkeeper/internal/client/crypto"
	grpcclient "gophkeeper/internal/client/grpc"
	"gophkeeper/internal/client/local"
	"gophkeeper/internal/client/payload"
	pb "gophkeeper/proto"
)

type CreateBinaryItemUseCase struct {
	store *local.Store
	grpc  *grpcclient.Client
}

func NewCreateBinaryItemUseCase(store *local.Store, grpc *grpcclient.Client) *CreateBinaryItemUseCase {
	return &CreateBinaryItemUseCase{
		store: store,
		grpc:  grpc,
	}
}

func (u *CreateBinaryItemUseCase) Execute(
	ctx context.Context,
	title, filePath, meta, masterPassword string,
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
	if filePath == "" {
		return "", fmt.Errorf("file path is required")
	}
	if masterPassword == "" {
		return "", fmt.Errorf("master password is required")
	}

	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("read file: %w", err)
	}

	salt, err := ensureKeySalt(state)
	if err != nil {
		return "", err
	}

	key, err := clientcrypto.DeriveKey(masterPassword, salt)
	if err != nil {
		return "", err
	}

	rawPayload, err := json.Marshal(payload.BinaryFile{
		FileName: filepath.Base(filePath),
		MimeType: "",
		Data:     fileData,
	})
	if err != nil {
		return "", fmt.Errorf("marshal binary payload: %w", err)
	}

	ciphertext, nonce, err := clientcrypto.Encrypt(key, rawPayload)
	if err != nil {
		return "", err
	}

	hash := clientcrypto.Hash(ciphertext)
	authCtx := grpcclient.ContextWithBearerToken(ctx, state.Token)

	req := pb.CreateItemRequest_builder{
		Type:             pb.ItemType_ITEM_TYPE_BINARY.Enum(),
		Title:            &title,
		Meta:             &meta,
		PayloadEncrypted: ciphertext,
		PayloadNonce:     nonce,
		PayloadHash:      &hash,
	}.Build()

	resp, err := u.grpc.Raw().CreateItem(authCtx, req)
	if err != nil {
		return "", fmt.Errorf("create binary item: %w", err)
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
