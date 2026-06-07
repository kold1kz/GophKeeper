// Package model содержит доменные модели приложения GophKeeper.
//
// В пакете описаны основные сущности предметной области,
// используемые сервером, репозиториями и сервисами.
package model

import (
	"time"

	"github.com/google/uuid"
)

// ItemType описывает тип хранимой записи.
type ItemType string

const (
	// ItemTypeLoginPassword обозначает запись типа логин/пароль.
	ItemTypeLoginPassword ItemType = "login_password"
	// ItemTypeText обозначает запись с произвольным текстом.
	ItemTypeText ItemType = "text"
	// ItemTypeBinary обозначает запись с бинарными данными.
	ItemTypeBinary ItemType = "binary"
	// ItemTypeBankCard обозначает запись с данными банковской карты.
	ItemTypeBankCard ItemType = "bank_card"
)

// VaultItem представляет собой единицу хранимых приватных данных пользователя.
//
// Запись хранит зашифрованный полезный payload, служебные поля,
// а также временные метки создания, обновления и удаления.
type VaultItem struct {
	// ID — уникальный идентификатор записи.
	ID uuid.UUID
	// UserID — идентификатор владельца записи.
	UserID string
	// Type — тип записи.
	Type ItemType
	// Title — заголовок записи.
	Title string
	// Meta — произвольная текстовая метаинформация.
	Meta string
	// PayloadEncrypted — зашифрованное содержимое записи.
	PayloadEncrypted []byte
	// PayloadNonce — nonce, использованный при шифровании payload.
	PayloadNonce []byte
	// PayloadHash — хэш payload для контроля целостности и сравнения.
	PayloadHash string
	// Version — версия записи для разрешения конфликтов обновления.
	Version int64
	// ClientUpdatedAt — время изменения записи на стороне клиента.
	ClientUpdatedAt *time.Time
	// CreatedAt — время создания записи.
	CreatedAt time.Time
	// UpdatedAt — время последнего обновления записи.
	UpdatedAt time.Time
	// DeletedAt — время мягкого удаления записи.
	DeletedAt *time.Time
}
