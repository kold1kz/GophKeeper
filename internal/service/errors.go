package service

import "errors"

var (

	// ErrInvalidRegisterData возвращается при ошибки регистрации.
	ErrInvalidRegisterData = errors.New("invalid register data")

	// ErrInvalidCredentials возвращается при неверном логине или пароле.
	ErrInvalidCredentials = errors.New("invalid credentials")

	// ErrInvalidItemData возвращается при некорректных данных элемента хранилища.
	ErrInvalidItemData = errors.New("invalid item data")

	// ErrItemNotFound возвращается, если элемент хранилища не найден.
	ErrItemNotFound = errors.New("item not found")

	// ErrItemDeleted возвращается, если элемент был удален и недоступен для операции.
	ErrItemDeleted = errors.New("item deleted")

	// ErrVersionConflict возвращается при конфликте версий во время обновления элемента.
	ErrVersionConflict = errors.New("version conflict")
)
