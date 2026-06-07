package grpcserver

import (
	"context"
	"testing"
)

func TestWithUserID_AndUserIDFromContext(t *testing.T) {
	t.Parallel()

	ctx := withUserID(context.Background(), "42")

	userID, ok := userIDFromContext(ctx)
	if !ok {
		t.Fatal("expected userID in context")
	}
	if userID != "42" {
		t.Fatalf("expected userID 42, got %q", userID)
	}
}

func TestValidUserIDFromContext_Success(t *testing.T) {
	t.Parallel()

	ctx := withUserID(context.Background(), "550e8400-e29b-41d4-a716-446655440000")

	id, ok := validUserIDFromContext(ctx)
	if !ok {
		t.Fatal("expected user id")
	}
	if id != "550e8400-e29b-41d4-a716-446655440000" {
		t.Fatalf("unexpected user id: %s", id)
	}
}

func TestValidUserIDFromContext_Empty(t *testing.T) {
	t.Parallel()

	ctx := withUserID(context.Background(), "")

	_, ok := validUserIDFromContext(ctx)
	if ok {
		t.Fatal("expected empty user id failure")
	}
}
