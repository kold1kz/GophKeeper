// Package local содержит типы и функции для работы с локальным состоянием
// клиента GophKeeper.
//
// Пакет предоставляет файловое хранилище для токена, параметров шифрования
// и локально сохраненных элементов.
package local

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Store представляет файловое хранилище локального состояния клиента.
//
// Состояние хранится в JSON-файле по заданному пути.
type Store struct {
	path string
}

// NewStore создает новое файловое хранилище локального состояния.
//
// path определяет путь к JSON-файлу состояния клиента.
func NewStore(path string) *Store {
	return &Store{path: path}
}

// Load загружает локальное состояние клиента из файла.
//
// Если файл состояния отсутствует, возвращается пустое состояние без ошибки.
// Возвращает ошибку, если файл не удалось прочитать или разобрать.
func (s *Store) Load() (*ClientState, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return &ClientState{}, nil
		}
		return nil, fmt.Errorf("read local state: %w", err)
	}

	var state ClientState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("unmarshal local state: %w", err)
	}

	return &state, nil
}

// Save сохраняет локальное состояние клиента в файл.
//
// При необходимости функция создает директорию для файла состояния.
// Возвращает ошибку, если состояние равно nil, сериализация не удалась
// или запись файла завершилась ошибкой.
func (s *Store) Save(state *ClientState) error {
	if state == nil {
		return fmt.Errorf("state is nil")
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal local state: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
		return fmt.Errorf("create local state dir: %w", err)
	}

	if err := os.WriteFile(s.path, data, 0o600); err != nil {
		return fmt.Errorf("write local state: %w", err)
	}

	return nil
}
