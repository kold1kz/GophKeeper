package usecase

import (
	"path/filepath"
	"testing"

	clientcrypto "gophkeeper/internal/client/crypto"
	"gophkeeper/internal/client/local"
)

func TestGetBinaryItemUseCase_WrongType(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	store := local.NewStore(filepath.Join(dir, "state.json"))

	salt := []byte("0123456789abcdef")
	key, err := clientcrypto.DeriveKey("master-password", salt)
	if err != nil {
		t.Fatalf("DeriveKey returned error: %v", err)
	}

	ciphertext, nonce, err := clientcrypto.Encrypt(key, []byte(`{"text":"hello"}`))
	if err != nil {
		t.Fatalf("Encrypt returned error: %v", err)
	}

	err = store.Save(&local.ClientState{
		KeySalt: salt,
		Items: []local.LocalItem{
			{
				ID:               "item-1",
				Type:             "ITEM_TYPE_TEXT",
				Title:            "note",
				PayloadEncrypted: ciphertext,
				PayloadNonce:     nonce,
			},
		},
	})
	if err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	uc := NewGetBinaryItemUseCase(store)

	err = uc.Execute("item-1", "master-password", filepath.Join(dir, "out"))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGetBinaryItemUseCase_InvalidPayloadJSON(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	store := local.NewStore(filepath.Join(dir, "state.json"))

	salt := []byte("0123456789abcdef")
	key, err := clientcrypto.DeriveKey("master-password", salt)
	if err != nil {
		t.Fatalf("DeriveKey returned error: %v", err)
	}

	ciphertext, nonce, err := clientcrypto.Encrypt(key, []byte(`not-json`))
	if err != nil {
		t.Fatalf("Encrypt returned error: %v", err)
	}

	err = store.Save(&local.ClientState{
		KeySalt: salt,
		Items: []local.LocalItem{
			{
				ID:               "item-1",
				Type:             "ITEM_TYPE_BINARY",
				Title:            "file1",
				PayloadEncrypted: ciphertext,
				PayloadNonce:     nonce,
			},
		},
	})
	if err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	uc := NewGetBinaryItemUseCase(store)

	err = uc.Execute("item-1", "master-password", filepath.Join(dir, "out"))
	if err == nil {
		t.Fatal("expected error")
	}
}
