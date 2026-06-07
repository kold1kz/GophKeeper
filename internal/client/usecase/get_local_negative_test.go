package usecase

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	clientcrypto "gophkeeper/internal/client/crypto"
	"gophkeeper/internal/client/local"
)

func captureStdoutGetLocal(t *testing.T, fn func()) string {
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

func TestGetLocalItemUseCase_UnknownType(t *testing.T) {

	dir := t.TempDir()
	store := local.NewStore(filepath.Join(dir, "state.json"))

	salt := []byte("0123456789abcdef")
	key, err := clientcrypto.DeriveKey("master-password", salt)
	if err != nil {
		t.Fatalf("DeriveKey returned error: %v", err)
	}

	ciphertext, nonce, err := clientcrypto.Encrypt(key, []byte(`{"x":"y"}`))
	if err != nil {
		t.Fatalf("Encrypt returned error: %v", err)
	}

	err = store.Save(&local.ClientState{
		KeySalt: salt,
		Items: []local.LocalItem{
			{
				ID:               "item-1",
				Type:             "ITEM_TYPE_UNKNOWN",
				Title:            "unknown",
				PayloadEncrypted: ciphertext,
				PayloadNonce:     nonce,
			},
		},
	})
	if err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	uc := NewGetLocalItemUseCase(store)

	out := captureStdoutGetLocal(t, func() {
		err = uc.Execute("item-1", "master-password")
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if !strings.Contains(out, `Payload (raw): {"x":"y"}`) {
		t.Fatalf("unexpected output: %q", out)
	}
}
