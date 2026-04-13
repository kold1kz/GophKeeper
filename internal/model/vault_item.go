package model

import "time"

type ItemType string

const (
	ItemTypeLoginPassword ItemType = "login_password"
	ItemTypeText          ItemType = "text"
	ItemTypeBinary        ItemType = "binary"
	ItemTypeBankCard      ItemType = "bank_card"
)

type VaultItem struct {
	ID               string
	UserID           int64
	Type             ItemType
	Title            string
	Meta             string
	PayloadEncrypted []byte
	PayloadNonce     []byte
	PayloadHash      string
	Version          int64
	ClientUpdatedAt  *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        *time.Time
}
