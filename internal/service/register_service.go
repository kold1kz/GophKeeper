package service

import (
	"context"
	"errors"
	"strings"

	"gophkeeper/internal/auth"
	"gophkeeper/internal/repository"
)

type RegisterService interface {
	Register(ctx context.Context, login, password string) (int, error)
}

type registerService struct {
	users repository.UserRepository
}

func NewRegisterService(users repository.UserRepository) RegisterService {
	return &registerService{users: users}
}

func (s *registerService) Register(ctx context.Context, login, password string) (int, error) {
	login = strings.TrimSpace(login)
	if login == "" || password == "" {
		return 0, ErrInvalidRegisterData
	}

	passwordHash, err := auth.HashPassword(password)
	if err != nil {
		return 0, err
	}

	id, err := s.users.Create(ctx, login, passwordHash)
	if err != nil {
		if errors.Is(err, repository.ErrUserAlreadyExists) {
			return 0, repository.ErrUserAlreadyExists
		}
		return 0, err
	}

	return id, nil
}
