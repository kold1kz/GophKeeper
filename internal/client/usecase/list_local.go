package usecase

import (
	"fmt"

	"gophkeeper/internal/client/local"
)

type ListLocalItemsUseCase struct {
	store *local.Store
}

func NewListLocalItemsUseCase(store *local.Store) *ListLocalItemsUseCase {
	return &ListLocalItemsUseCase{
		store: store,
	}
}

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
