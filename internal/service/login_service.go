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

type loginService struct {
	users repository.UserRepository
}

func NewLoginService(users repository.UserRepository) LoginService {
	return &loginService{users: users}
}

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
