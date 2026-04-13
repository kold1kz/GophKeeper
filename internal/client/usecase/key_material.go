package usecase

import (
	"crypto/rand"
	"fmt"

	"gophkeeper/internal/client/local"
)

const saltSize = 16

func ensureKeySalt(state *local.ClientState) ([]byte, error) {
	if len(state.KeySalt) != 0 {
		return state.KeySalt, nil
	}

	salt := make([]byte, saltSize)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("generate key salt: %w", err)
	}

	state.KeySalt = salt
	return salt, nil
}
