package local

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestStore_Load_NotExists_ReturnsEmptyState(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	store := NewStore(filepath.Join(dir, "state.json"))

	state, err := store.Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if state == nil {
		t.Fatal("expected non-nil state")
	}
	if state.Token != "" {
		t.Fatalf("expected empty token, got %q", state.Token)
	}
	if len(state.Items) != 0 {
		t.Fatalf("expected no items, got %d", len(state.Items))
	}
}

func TestStore_SaveLoad_RoundTrip(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	store := NewStore(filepath.Join(dir, ".gophkeeper", "state.json"))

	now := time.Now().UTC().Round(0)
	state := &ClientState{
		Token:      "token-123",
		KeySalt:    []byte("salt-123"),
		LastSyncAt: &now,
		Items: []LocalItem{
			{
				ID:               "item-1",
				Type:             "ITEM_TYPE_TEXT",
				Title:            "note1",
				Meta:             "meta1",
				PayloadEncrypted: []byte("cipher"),
				PayloadNonce:     []byte("nonce"),
				PayloadHash:      "hash",
				Version:          1,
				CreatedAt:        now,
				UpdatedAt:        now,
			},
		},
	}

	if err := store.Save(state); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if loaded.Token != state.Token {
		t.Fatalf("token mismatch: got %q want %q", loaded.Token, state.Token)
	}
	if string(loaded.KeySalt) != string(state.KeySalt) {
		t.Fatalf("salt mismatch: got %q want %q", loaded.KeySalt, state.KeySalt)
	}
	if loaded.LastSyncAt == nil {
		t.Fatal("expected LastSyncAt")
	}
	if !loaded.LastSyncAt.Equal(*state.LastSyncAt) {
		t.Fatalf("LastSyncAt mismatch: got %v want %v", *loaded.LastSyncAt, *state.LastSyncAt)
	}
	if len(loaded.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(loaded.Items))
	}
	if loaded.Items[0].ID != "item-1" {
		t.Fatalf("unexpected item ID: %q", loaded.Items[0].ID)
	}
}

func TestStore_Save_NilState(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	store := NewStore(filepath.Join(dir, "state.json"))

	err := store.Save(nil)
	if err == nil {
		t.Fatal("expected error for nil state")
	}
}

func TestStore_Load_InvalidJSON(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "state.json")
	if err := os.WriteFile(path, []byte("{bad json"), 0o600); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	store := NewStore(path)

	_, err := store.Load()
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}
