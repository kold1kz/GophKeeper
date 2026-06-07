// Package usecase содержит прикладные сценарии работы CLI-клиента GophKeeper.
//
// Пакет реализует пользовательские сценарии работы с локальным состоянием,
// шифрованием, синхронизацией и удалённым gRPC-сервером.
package usecase

import (
	"crypto/rand"
	"fmt"

	"gophkeeper/internal/client/local"
)

const saltSize = 16

// ensureKeySalt возвращает соль для вывода клиентского ключа.
//
// Если соль уже сохранена в локальном состоянии, функция возвращает её.
// Если соли ещё нет, функция генерирует новую случайную соль,
// сохраняет её в состоянии клиента и возвращает результат.
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
