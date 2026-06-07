// Package usecase содержит прикладные сценарии работы CLI-клиента GophKeeper.
package usecase

import (
	"fmt"

	"gophkeeper/internal/client/local"
)

// ListLocalItemsUseCase описывает сценарий вывода локальных записей.
type ListLocalItemsUseCase struct {
	store *local.Store
}

// NewListLocalItemsUseCase создаёт новый use case для вывода локальных записей.
func NewListLocalItemsUseCase(store *local.Store) *ListLocalItemsUseCase {
	return &ListLocalItemsUseCase{
		store: store,
	}
}

// Execute загружает локальное состояние и выводит в консоль все активные записи.
//
// Удалённые записи не выводятся. Если активных записей нет,
// функция печатает сообщение "no items".
func (u *ListLocalItemsUseCase) Execute() error {
	state, err := u.store.Load()
	if err != nil {
		return err
	}

	hasActiveItems := false
	for _, item := range state.Items {
		if item.DeletedAt != nil {
			continue
		}

		hasActiveItems = true

		fmt.Println("ID:", item.ID)
		fmt.Println("Type:", item.Type)
		fmt.Println("Title:", item.Title)
		fmt.Println("Meta:", item.Meta)
		fmt.Println("Updated:", item.UpdatedAt)
		fmt.Println("---")
	}

	if !hasActiveItems {
		fmt.Println("no items")
	}

	return nil
}
