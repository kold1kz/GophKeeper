package usecase

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"gophkeeper/internal/client/local"
)

func TestListLocalItemsUseCase_NoItems(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	store := local.NewStore(filepath.Join(dir, "state.json"))

	uc := NewListLocalItemsUseCase(store)

	output := captureStdout(t, func() {
		if err := uc.Execute(); err != nil {
			t.Fatalf("Execute returned error: %v", err)
		}
	})

	if !strings.Contains(output, "no items") {
		t.Fatalf("expected output to contain %q, got %q", "", output)
	}
}

func TestListLocalItemsUseCase_ShowsOnlyActiveItems(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	store := local.NewStore(filepath.Join(dir, "state.json"))

	now := time.Now().UTC().Round(0)
	deletedAt := now

	err := store.Save(&local.ClientState{
		Items: []local.LocalItem{
			{
				ID:        "1",
				Type:      "ITEM_TYPE_TEXT",
				Title:     "visible-item",
				Meta:      "meta1",
				UpdatedAt: now,
			},
			{
				ID:        "2",
				Type:      "ITEM_TYPE_TEXT",
				Title:     "deleted-item",
				Meta:      "meta2",
				UpdatedAt: now,
				DeletedAt: &deletedAt,
			},
		},
	})
	if err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	uc := NewListLocalItemsUseCase(store)

	output := captureStdout(t, func() {
		if err := uc.Execute(); err != nil {
			t.Fatalf("Execute returned error: %v", err)
		}
	})

	if !strings.Contains(output, "visible-item") {
		t.Fatalf("expected visible item in output, got %q", output)
	}
	if strings.Contains(output, "deleted-item") {
		t.Fatalf("did not expect deleted item in output, got %q", output)
	}
}
