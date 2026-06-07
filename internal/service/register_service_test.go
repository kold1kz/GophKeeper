package service

import (
	"context"
	"errors"
	"testing"

	"gophkeeper/internal/repository"
)

func TestRegisterService_Register_Success(t *testing.T) {
	t.Parallel()

	repo := &userRepoMock{
		findByUsernameFn: func(ctx context.Context, login string) (*repository.User, error) {
			return nil, nil
		},
		createFn: func(ctx context.Context, login, passwordHash string) (string, error) {
			if login != "u1" {
				t.Fatalf("unexpected login: %s", login)
			}
			if passwordHash == "" {
				t.Fatal("expected password hash")
			}
			return "550e8400-e29b-41d4-a716-446655440000", nil
		},
	}

	svc := NewRegisterService(repo)

	id, err := svc.Register(context.Background(), "u1", "p1")
	if err != nil {
		t.Fatalf("Register returned error: %v", err)
	}
	if id != "550e8400-e29b-41d4-a716-446655440000" {
		t.Fatalf("unexpected id: %s", id)
	}
}

func TestRegisterService_Register_UserAlreadyExists(t *testing.T) {
	t.Parallel()

	repo := &userRepoMock{
		findByUsernameFn: func(ctx context.Context, login string) (*repository.User, error) {
			return nil, nil
		},
		createFn: func(ctx context.Context, login, passwordHash string) (string, error) {
			return "", repository.ErrUserAlreadyExists
		},
	}

	svc := NewRegisterService(repo)

	_, err := svc.Register(context.Background(), "u1", "p1")
	if !errors.Is(err, repository.ErrUserAlreadyExists) {
		t.Fatalf("expected ErrUserAlreadyExists, got %v", err)
	}
}
