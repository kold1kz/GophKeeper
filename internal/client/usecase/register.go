// Package usecase содержит прикладные сценарии работы CLI-клиента GophKeeper.
package usecase

import (
	"context"
	"fmt"

	grpcclient "gophkeeper/internal/client/grpc"
	pb "gophkeeper/proto"
)

// RegisterUseCase описывает сценарий регистрации нового пользователя.
type RegisterUseCase struct {
	grpc *grpcclient.Client
}

// NewRegisterUseCase создаёт новый use case для регистрации пользователя.
func NewRegisterUseCase(grpc *grpcclient.Client) *RegisterUseCase {
	return &RegisterUseCase{grpc: grpc}
}

// Execute отправляет на сервер запрос на регистрацию нового пользователя.
//
// В случае успеха функция возвращает идентификатор созданного пользователя.
func (u *RegisterUseCase) Execute(ctx context.Context, login, password string) (string, error) {
	resp, err := u.grpc.Raw().Register(ctx, pb.RegisterRequest_builder{
		Login:    &login,
		Password: &password,
	}.Build())
	if err != nil {
		return "", fmt.Errorf("register: %w", err)
	}

	return resp.GetUserId(), nil
}
