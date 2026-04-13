// Package service содержит бизнес-логику приложения.
//
// Пакет реализует сценарии регистрации, аутентификации и работы
// с пользовательским хранилищем данных.
package service

import (
	"context"
	"strconv"
	"strings"

	"gophkeeper/internal/auth"
	"gophkeeper/internal/repository"
)

type LoginService interface {
	Login(ctx context.Context, login, password string) (string, error)
}

// LoginService предоставляет бизнес-логику аутентификации пользователя.
//
// Сервис проверяет учетные данные пользователя и выдает токен доступа
// при успешной аутентификации.
type loginService struct {
	users repository.UserRepository
}

// NewLoginService создает сервис аутентификации пользователя.
func NewLoginService(users repository.UserRepository) LoginService {
	return &loginService{users: users}
}

// Login выполняет аутентификацию пользователя по логину и паролю.
//
// При успешной проверке возвращает токен доступа.
// Возвращает ErrInvalidCredentials, если пользователь не найден
// или пароль не совпадает.
func (s *loginService) Login(ctx context.Context, login, password string) (string, error) {
	login = strings.TrimSpace(login)
	if login == "" || password == "" {
		return "", ErrInvalidCredentials
	}

	user, err := s.users.FindByUsername(ctx, login)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", ErrInvalidCredentials
	}

	if err := auth.CheckPasswordHash(password, user.Password); err != nil {
		return "", ErrInvalidCredentials
	}

	token, err := auth.NewTokenForUserID(strconv.Itoa(user.ID))
	if err != nil {
		return "", err
	}

	return token, nil
}
