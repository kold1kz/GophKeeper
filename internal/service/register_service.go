package service

import (
	"context"
	"errors"
	"strings"

	"gophkeeper/internal/auth"
	"gophkeeper/internal/repository"
)

type RegisterService interface {
	Register(ctx context.Context, login, password string) (string, error)
}

// RegisterService предоставляет бизнес-логику регистрации пользователей.
//
// Сервис создает нового пользователя, предварительно проверяя,
// что пользователь с таким логином еще не существует.
type registerService struct {
	users repository.UserRepository
}

// NewRegisterService создает сервис регистрации пользователя.
func NewRegisterService(users repository.UserRepository) RegisterService {
	return &registerService{users: users}
}

// Register регистрирует нового пользователя.
//
// Пароль пользователя хэшируется перед сохранением.
// Возвращает ErrUserAlreadyExists, если пользователь с таким логином уже существует.

func (s *registerService) Register(ctx context.Context, login, password string) (string, error) {
	login = strings.TrimSpace(login)
	if login == "" || password == "" {
		return "", ErrInvalidRegisterData
	}

	passwordHash, err := auth.HashPassword(password)
	if err != nil {
		return "", err
	}

	id, err := s.users.Create(ctx, login, passwordHash)
	if err != nil {
		if errors.Is(err, repository.ErrUserAlreadyExists) {
			return "", repository.ErrUserAlreadyExists
		}
		return "", err
	}

	return id, nil
}
