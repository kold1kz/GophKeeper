package usecase

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	clientcrypto "gophkeeper/internal/client/crypto"
	"gophkeeper/internal/client/local"
	"gophkeeper/internal/client/payload"
)

func captureStdoutForGetLocal(t *testing.T, fn func()) string {
	t.Helper()

	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe returned error: %v", err)
	}
	os.Stdout = w
	defer func() { os.Stdout = oldStdout }()

	fn()

	_ = w.Close()

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	return buf.String()
}

func TestGetLocalItemUseCase_Success_LoginPassword(t *testing.T) {
	dir := t.TempDir()
	store := local.NewStore(filepath.Join(dir, "state.json"))

	salt := []byte("0123456789abcdef")
	key, err := clientcrypto.DeriveKey("master-password", salt)
	if err != nil {
		t.Fatalf("DeriveKey returned error: %v", err)
	}

	rawPayload, err := json.Marshal(payload.LoginPassword{
		Login:    "alice",
		Password: "secret123",
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
				Type:             "ITEM_TYPE_LOGIN_PASSWORD",
				Title:            "github",
				Meta:             "work",
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
		err = uc.Execute("item-1", "master-password")
	})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if !strings.Contains(out, "github") {
		t.Fatalf("expected title in output, got %q", out)
	}
	if !strings.Contains(out, "alice") {
		t.Fatalf("expected login in output, got %q", out)
	}
	if !strings.Contains(out, "secret123") {
		t.Fatalf("expected password in output, got %q", out)
	}
}
