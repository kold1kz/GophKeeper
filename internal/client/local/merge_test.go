package local

import (
	"testing"
	"time"
)

func TestApplySyncedItems_AppendsNewItems(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC().Round(0)
	state := &ClientState{
		Items: []LocalItem{
			{ID: "1", Title: "old"},
		},
	}

	ApplySyncedItems(state, []LocalItem{
		{ID: "2", Title: "new"},
	}, now)

	if len(state.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(state.Items))
	}
	if state.Items[1].ID != "2" {
		t.Fatalf("unexpected appended item ID: %q", state.Items[1].ID)
	}
	if state.LastSyncAt == nil || !state.LastSyncAt.Equal(now) {
		t.Fatalf("unexpected LastSyncAt: %v", state.LastSyncAt)
	}
}

func TestApplySyncedItems_ReplacesExistingItems(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC().Round(0)
	state := &ClientState{
		Items: []LocalItem{
			{ID: "1", Title: "old-title", Version: 1},
		},
	}

	ApplySyncedItems(state, []LocalItem{
		{ID: "1", Title: "new-title", Version: 2},
	}, now)

	if len(state.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(state.Items))
	}
	if state.Items[0].Title != "new-title" {
		t.Fatalf("expected replaced title, got %q", state.Items[0].Title)
	}
	if state.Items[0].Version != 2 {
		t.Fatalf("expected version 2, got %d", state.Items[0].Version)
	}
}
