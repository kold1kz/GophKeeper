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

func TestUserIDInt64FromContext_Success(t *testing.T) {
	t.Parallel()

	ctx := withUserID(context.Background(), "123")

	id, ok := userIDInt64FromContext(ctx)
	if !ok {
		t.Fatal("expected int64 user id")
	}
	if id != 123 {
		t.Fatalf("expected 123, got %d", id)
	}
}

func TestUserIDInt64FromContext_Invalid(t *testing.T) {
	t.Parallel()

	ctx := withUserID(context.Background(), "abc")

	_, ok := userIDInt64FromContext(ctx)
	if ok {
		t.Fatal("expected conversion failure")
	}
}
