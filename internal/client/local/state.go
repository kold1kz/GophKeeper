// Package local содержит типы и функции для работы с локальным состоянием
// клиента GophKeeper.
//
// Пакет используется CLI-клиентом для хранения токена, ключевого материала,
// времени последней синхронизации и локальной копии пользовательских данных.
package local

import "time"

// ClientState описывает локальное состояние клиента.
//
// Состояние включает:
//   - токен авторизации;
//   - соль для вывода ключа шифрования;
//   - время последней синхронизации;
//   - список локально сохраненных элементов.
type ClientState struct {
	// Token содержит токен авторизации пользователя.
	Token string `json:"token"`
	// KeySalt содержит соль, используемую для вывода ключа из мастер-пароля.
	KeySalt []byte `json:"key_salt,omitempty"`
	// LastSyncAt содержит время последней успешной синхронизации с сервером.
	LastSyncAt *time.Time `json:"last_sync_at,omitempty"`
	// Items содержит локально сохраненные элементы пользователя.
	Items []LocalItem `json:"items"`
}

// LocalItem описывает элемент, сохраненный в локальном состоянии клиента.
//
// Полезная нагрузка хранится в зашифрованном виде.
type LocalItem struct {
	// ID содержит уникальный идентификатор записи.
	ID string `json:"id"`
	// Type содержит тип записи.
	Type string `json:"type"`
	// Title содержит заголовок записи.
	Title string `json:"title"`
	// Meta содержит дополнительную текстовую метаинформацию.
	Meta string `json:"meta"`
	// PayloadEncrypted содержит зашифрованную полезную нагрузку.
	PayloadEncrypted []byte `json:"payload_encrypted"`
	// PayloadNonce содержит nonce, использованный при шифровании полезной нагрузки.
	PayloadNonce []byte `json:"payload_nonce"`
	// PayloadHash содержит хэш полезной нагрузки.
	PayloadHash string `json:"payload_hash"`
	// Version содержит версию записи для разрешения конфликтов синхронизации.
	Version int64 `json:"version"`
	// ClientUpdatedAt содержит локальное время последнего изменения записи.
	ClientUpdatedAt *time.Time `json:"client_updated_at,omitempty"`
	// CreatedAt содержит время создания записи.
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt содержит время последнего обновления записи.
	UpdatedAt time.Time `json:"updated_at"`
	// DeletedAt содержит время мягкого удаления записи, если она была удалена.
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}
