package usecase

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	clientcrypto "gophkeeper/internal/client/crypto"
	"gophkeeper/internal/client/local"
	"gophkeeper/internal/client/payload"
)

func TestGetBinaryItemUseCase_Success(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	store := local.NewStore(filepath.Join(dir, "state.json"))

	salt := []byte("0123456789abcdef")
	key, err := clientcrypto.DeriveKey("master-password", salt)
	if err != nil {
		t.Fatalf("DeriveKey returned error: %v", err)
	}

	rawPayload, err := json.Marshal(payload.BinaryFile{
		FileName: "test.txt",
		MimeType: "text/plain",
		Data:     []byte("hello file"),
	})
	if err != nil {
		t.Fatalf("json.Marshal returned error: %v", err)
	}

	ciphertext, nonce, err := clientcrypto.Encrypt(key, rawPayload)
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

	outputDir := filepath.Join(dir, "out")

	uc := NewGetBinaryItemUseCase(store)

	err = uc.Execute("item-1", "master-password", outputDir)
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(outputDir, "test.txt"))
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}

	if string(data) != "hello file" {
		t.Fatalf("expected file content %q, got %q", "hello file", string(data))
	}
}

func TestGetBinaryItemUseCase_ItemNotFound(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	store := local.NewStore(filepath.Join(dir, "state.json"))

	err := store.Save(&local.ClientState{
		KeySalt: []byte("0123456789abcdef"),
	})
	if err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	uc := NewGetBinaryItemUseCase(store)

	err = uc.Execute("missing", "master-password", dir)
	if err == nil {
		t.Fatal("expected error")
	}
}
