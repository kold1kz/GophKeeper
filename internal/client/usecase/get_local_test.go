package usecase

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	clientcrypto "gophkeeper/internal/client/crypto"
	"gophkeeper/internal/client/local"
)

func TestGetLocalItemUseCase_Success_Text(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	store := local.NewStore(filepath.Join(dir, "state.json"))

	salt := []byte("0123456789abcdef")
	key, err := clientcrypto.DeriveKey("mpass", salt)
	if err != nil {
		t.Fatalf("DeriveKey returned error: %v", err)
	}

	ciphertext, nonce, err := clientcrypto.Encrypt(key, []byte("hello world"))
	if err != nil {
		t.Fatalf("Encrypt returned error: %v", err)
	}

	now := time.Now().UTC().Round(0)
	err = store.Save(&local.ClientState{
		KeySalt: salt,
		Items: []local.LocalItem{
			{
				ID:               "item-1",
				Type:             "ITEM_TYPE_TEXT",
				Title:            "note1",
				Meta:             "meta1",
				PayloadEncrypted: ciphertext,
				PayloadNonce:     nonce,
				Version:          1,
				CreatedAt:        now,
				UpdatedAt:        now,
			},
		},
	})
	if err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	uc := NewGetLocalItemUseCase(store)

	output := captureStdout(t, func() {
		if err := uc.Execute("item-1", "mpass"); err != nil {
			t.Fatalf("Execute returned error: %v", err)
		}
	})

	if !strings.Contains(output, "note1") {
		t.Fatalf("expected title in output, got %q", output)
	}
	if !strings.Contains(output, "hello world") {
		t.Fatalf("expected decrypted payload in output, got %q", output)
	}
}

func TestGetLocalItemUseCase_ItemNotFound(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	store := local.NewStore(filepath.Join(dir, "state.json"))

	err := store.Save(&local.ClientState{
		KeySalt: []byte("0123456789abcdef"),
	})
	if err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	uc := NewGetLocalItemUseCase(store)

	err = uc.Execute("missing", "mpass")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "item not found") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetLocalItemUseCase_WrongMasterPassword(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	store := local.NewStore(filepath.Join(dir, "state.json"))

	salt := []byte("0123456789abcdef")
	key, err := clientcrypto.DeriveKey("right-pass", salt)
	if err != nil {
		t.Fatalf("DeriveKey returned error: %v", err)
	}

	ciphertext, nonce, err := clientcrypto.Encrypt(key, []byte("secret"))
	if err != nil {
		t.Fatalf("Encrypt returned error: %v", err)
	}

	now := time.Now().UTC().Round(0)
	err = store.Save(&local.ClientState{
		KeySalt: salt,
		Items: []local.LocalItem{
			{
				ID:               "item-1",
				Type:             "ITEM_TYPE_TEXT",
				Title:            "note1",
				PayloadEncrypted: ciphertext,
				PayloadNonce:     nonce,
				CreatedAt:        now,
				UpdatedAt:        now,
			},
		},
	})
	if err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	uc := NewGetLocalItemUseCase(store)

	err = uc.Execute("item-1", "wrong-pass")
	if err == nil {
		t.Fatal("expected decrypt error")
	}
}
