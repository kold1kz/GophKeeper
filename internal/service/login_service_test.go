package service

import (
	"context"
	"errors"
	"testing"

	"gophkeeper/internal/auth"
	"gophkeeper/internal/repository"
)

type userRepoMock struct {
	findByUsernameFn func(ctx context.Context, login string) (*repository.User, error)
	createFn         func(ctx context.Context, login, passwordHash string) (int, error)
}

func (m *userRepoMock) FindByUsername(ctx context.Context, login string) (*repository.User, error) {
	return m.findByUsernameFn(ctx, login)
}

func (m *userRepoMock) Create(ctx context.Context, login, passwordHash string) (int, error) {
	return m.createFn(ctx, login, passwordHash)
}

func TestLoginService_Login_Success(t *testing.T) {
	t.Parallel()

	hash, err := auth.HashPassword("p1")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}

	repo := &userRepoMock{
		findByUsernameFn: func(ctx context.Context, login string) (*repository.User, error) {
			return &repository.User{
				ID:       1,
				Login:    "u1",
				Password: hash,
			}, nil
		},
		createFn: func(ctx context.Context, login, passwordHash string) (int, error) {
			return 0, nil
		},
	}

	svc := NewLoginService(repo)

	token, err := svc.Login(context.Background(), "u1", "p1")
	if err != nil {
		t.Fatalf("Login returned error: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}
}

func TestLoginService_Login_InvalidCredentials(t *testing.T) {
	t.Parallel()

	repo := &userRepoMock{
		findByUsernameFn: func(ctx context.Context, login string) (*repository.User, error) {
			return nil, nil
		},
		createFn: func(ctx context.Context, login, passwordHash string) (int, error) {
			return 0, nil
		},
	}

	svc := NewLoginService(repo)

	_, err := svc.Login(context.Background(), "u1", "p1")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}
