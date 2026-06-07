// Package usecase содержит прикладные сценарии работы CLI-клиента GophKeeper.
package usecase

import (
	"encoding/json"
	"fmt"

	clientcrypto "gophkeeper/internal/client/crypto"
	"gophkeeper/internal/client/local"
	"gophkeeper/internal/client/payload"
)

// GetLocalItemUseCase описывает сценарий получения и вывода
// локально сохраненной записи.
type GetLocalItemUseCase struct {
	store *local.Store
}

// NewGetLocalItemUseCase создает новый use case для получения
// локальной записи.
func NewGetLocalItemUseCase(store *local.Store) *GetLocalItemUseCase {
	return &GetLocalItemUseCase{
		store: store,
	}
}

// Execute находит запись в локальном хранилище, расшифровывает ее
// и выводит данные в консоль.
//
// Формат вывода зависит от типа записи.
func (u *GetLocalItemUseCase) Execute(id, masterPassword string) error {
	if id == "" {
		return fmt.Errorf("item id is required")
	}
	if masterPassword == "" {
		return fmt.Errorf("master password is required")
	}

	state, err := u.store.Load()
	if err != nil {
		return err
	}
	if len(state.KeySalt) == 0 {
		return fmt.Errorf("key salt is missing; create or sync encrypted data first")
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

		fmt.Println("ID:", item.ID)
		fmt.Println("Type:", item.Type)
		fmt.Println("Title:", item.Title)
		fmt.Println("Meta:", item.Meta)
		fmt.Println("Version:", item.Version)
		fmt.Println("Created:", item.CreatedAt)
		fmt.Println("Updated:", item.UpdatedAt)

		switch item.Type {
		case "ITEM_TYPE_TEXT":
			fmt.Println("Payload:", string(plaintext))

		case "ITEM_TYPE_LOGIN_PASSWORD":
			var data payload.LoginPassword
			if err := json.Unmarshal(plaintext, &data); err != nil {
				return fmt.Errorf("unmarshal login/password payload: %w", err)
			}
			fmt.Println("Login:", data.Login)
			fmt.Println("Password:", data.Password)

		case "ITEM_TYPE_BANK_CARD":
			var data payload.BankCard
			if err := json.Unmarshal(plaintext, &data); err != nil {
				return fmt.Errorf("unmarshal bank card: %w", err)
			}

			fmt.Println("Number:", data.Number)
			fmt.Println("Holder:", data.Holder)
			fmt.Println("Expiry:", data.Expiry)
			fmt.Println("CVV:", data.CVV)

		case "ITEM_TYPE_BINARY":
			var data payload.BinaryFile
			if err := json.Unmarshal(plaintext, &data); err != nil {
				return fmt.Errorf("unmarshal binary payload: %w", err)
			}

			fmt.Println("FileName:", data.FileName)
			fmt.Println("MimeType:", data.MimeType)
			fmt.Println("Size:", len(data.Data))

		default:
			fmt.Println("Payload (raw):", string(plaintext))
		}

		return nil
	}

	return fmt.Errorf("item not found")
}
