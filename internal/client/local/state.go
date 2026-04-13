package local

import "time"

type ClientState struct {
	Token      string      `json:"token"`
	KeySalt    []byte      `json:"key_salt,omitempty"`
	LastSyncAt *time.Time  `json:"last_sync_at,omitempty"`
	Items      []LocalItem `json:"items"`
}

type LocalItem struct {
	ID               string     `json:"id"`
	Type             string     `json:"type"`
	Title            string     `json:"title"`
	Meta             string     `json:"meta"`
	PayloadEncrypted []byte     `json:"payload_encrypted"`
	PayloadNonce     []byte     `json:"payload_nonce"`
	PayloadHash      string     `json:"payload_hash"`
	Version          int64      `json:"version"`
	ClientUpdatedAt  *time.Time `json:"client_updated_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	DeletedAt        *time.Time `json:"deleted_at,omitempty"`
}
