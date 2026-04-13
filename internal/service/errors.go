package service

import "errors"

var (
	ErrInvalidRegisterData = errors.New("invalid register data")
	ErrInvalidCredentials  = errors.New("invalid credentials")

	ErrInvalidItemData = errors.New("invalid item data")
	ErrItemNotFound    = errors.New("item not found")
	ErrItemDeleted     = errors.New("item deleted")
	ErrVersionConflict = errors.New("version conflict")
)
