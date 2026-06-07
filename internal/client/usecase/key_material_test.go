package usecase

import (
	"bytes"
	"testing"

	"gophkeeper/internal/client/local"
)

func TestEnsureKeySalt_GeneratesSalt(t *testing.T) {
	t.Parallel()

	state := &local.ClientState{}

	salt, err := ensureKeySalt(state)
	if err != nil {
		t.Fatalf("ensureKeySalt returned error: %v", err)
	}
	if len(salt) == 0 {
		t.Fatal("expected generated salt")
	}
	if !bytes.Equal(state.KeySalt, salt) {
		t.Fatal("expected state.KeySalt to be set")
	}
}

func TestEnsureKeySalt_ReusesExistingSalt(t *testing.T) {
	t.Parallel()

	state := &local.ClientState{
		KeySalt: []byte("existing-salt"),
	}

	salt, err := ensureKeySalt(state)
	if err != nil {
		t.Fatalf("ensureKeySalt returned error: %v", err)
	}
	if !bytes.Equal(salt, []byte("existing-salt")) {
		t.Fatalf("expected existing salt, got %q", salt)
	}
}
