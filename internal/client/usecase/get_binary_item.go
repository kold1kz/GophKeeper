package usecase

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	clientcrypto "gophkeeper/internal/client/crypto"
	"gophkeeper/internal/client/local"
	"gophkeeper/internal/client/payload"
)

type GetBinaryItemUseCase struct {
	store *local.Store
}

func NewGetBinaryItemUseCase(store *local.Store) *GetBinaryItemUseCase {
	return &GetBinaryItemUseCase{
		store: store,
	}
}

func (u *GetBinaryItemUseCase) Execute(id, masterPassword, outputDir string) error {
	if id == "" {
		return fmt.Errorf("item id is required")
	}
	if masterPassword == "" {
		return fmt.Errorf("master password is required")
	}
	if outputDir == "" {
		return fmt.Errorf("output dir is required")
	}

	state, err := u.store.Load()
	if err != nil {
		return err
	}
	if len(state.KeySalt) == 0 {
		return fmt.Errorf("key salt is missing")
	}

	key, err := clientcrypto.DeriveKey(masterPassword, state.KeySalt)
	if err != nil {
		return err
	}

	for _, item := range state.Items {
		if item.ID != id {
			continue
		}
		if item.DeletedAt != nil {
			return fmt.Errorf("item is deleted")
		}

		plaintext, err := clientcrypto.Decrypt(key, item.PayloadEncrypted, item.PayloadNonce)
		if err != nil {
			return err
		}

		var data payload.BinaryFile
		if err := json.Unmarshal(plaintext, &data); err != nil {
			return fmt.Errorf("unmarshal binary payload: %w", err)
		}

		if data.FileName == "" {
			return fmt.Errorf("binary payload has empty file name")
		}

		if err := os.MkdirAll(outputDir, 0o755); err != nil {
			return fmt.Errorf("create output dir: %w", err)
		}

		outPath := filepath.Join(outputDir, data.FileName)
		if err := os.WriteFile(outPath, data.Data, 0o644); err != nil {
			return fmt.Errorf("write output file: %w", err)
		}

		fmt.Println("file written to:", outPath)
		return nil
	}

	return fmt.Errorf("item not found")
}
