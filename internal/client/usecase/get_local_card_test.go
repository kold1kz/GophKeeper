package usecase

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	clientcrypto "gophkeeper/internal/client/crypto"
	"gophkeeper/internal/client/local"
	"gophkeeper/internal/client/payload"
)

func TestGetLocalItemUseCase_Success_BankCard(t *testing.T) {
	dir := t.TempDir()
	store := local.NewStore(filepath.Join(dir, "state.json"))

	salt := []byte("0123456789abcdef")
	key, err := clientcrypto.DeriveKey("master-password", salt)
	if err != nil {
		t.Fatalf("DeriveKey returned error: %v", err)
	}

	rawPayload, err := json.Marshal(payload.BankCard{
		Number: "4111111111111111",
		Holder: "JOHN DOE",
		Expiry: "12/30",
		CVV:    "123",
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
				ID:               "item-2",
				Type:             "ITEM_TYPE_BANK_CARD",
				Title:            "visa",
				PayloadEncrypted: ciphertext,
				PayloadNonce:     nonce,
			},
		},
	})
	if err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	uc := NewGetLocalItemUseCase(store)

	out := captureStdoutForGetLocal(t, func() {
		err = uc.Execute("item-2", "master-password")
	})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if !strings.Contains(out, "4111111111111111") {
		t.Fatalf("expected card number in output, got %q", out)
	}
	if !strings.Contains(out, "JOHN DOE") {
		t.Fatalf("expected holder in output, got %q", out)
	}
}

func TestGetLocalItemUseCase_InvalidJSONPayload(t *testing.T) {
	dir := t.TempDir()
	store := local.NewStore(filepath.Join(dir, "state.json"))

	salt := []byte("0123456789abcdef")
	key, err := clientcrypto.DeriveKey("master-password", salt)
	if err != nil {
		t.Fatalf("DeriveKey returned error: %v", err)
	}

	ciphertext, nonce, err := clientcrypto.Encrypt(key, []byte(`{"broken"`))
	if err != nil {
		t.Fatalf("Encrypt returned error: %v", err)
	}

	err = store.Save(&local.ClientState{
		KeySalt: salt,
		Items: []local.LocalItem{
			{
				ID:               "item-3",
				Type:             "ITEM_TYPE_LOGIN_PASSWORD",
				Title:            "broken",
				PayloadEncrypted: ciphertext,
				PayloadNonce:     nonce,
			},
		},
	})
	if err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	uc := NewGetLocalItemUseCase(store)

	err = uc.Execute("item-3", "master-password")
	if err == nil {
		t.Fatal("expected error")
	}
}
