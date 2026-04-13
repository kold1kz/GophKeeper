package usecase

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	grpcclient "gophkeeper/internal/client/grpc"
	"gophkeeper/internal/client/local"
)

func TestCreateLoginPasswordItemUseCase_NotAuthenticated(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	store := local.NewStore(filepath.Join(dir, "state.json"))

	client := &grpcclient.Client{}
	uc := NewCreateLoginPasswordItemUseCase(store, client)

	_, err := uc.Execute(
		context.Background(),
		"title",
		"login",
		"password",
		"meta",
		"master-password",
	)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "not authenticated") {
		t.Fatalf("unexpected error: %v", err)
	}
}
