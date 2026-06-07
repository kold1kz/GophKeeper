// Package usecase содержит прикладные сценарии работы CLI-клиента GophKeeper.
package usecase

import (
	"context"
	"fmt"

	grpcclient "gophkeeper/internal/client/grpc"
	"gophkeeper/internal/client/local"
	pb "gophkeeper/proto"
)

// LoginUseCase описывает сценарий аутентификации пользователя.
type LoginUseCase struct {
	store *local.Store
	grpc  *grpcclient.Client
}

// NewLoginUseCase создаёт новый use case для входа пользователя в систему.
func NewLoginUseCase(store *local.Store, grpc *grpcclient.Client) *LoginUseCase {
	return &LoginUseCase{
		store: store,
		grpc:  grpc,
	}
}

// Execute выполняет аутентификацию пользователя на сервере.
//
// После успешного входа функция сохраняет полученный токен
// в локальном состоянии клиента.
func (u *LoginUseCase) Execute(ctx context.Context, login, password string) error {
	resp, err := u.grpc.Raw().Login(ctx, pb.LoginRequest_builder{
		Login:    &login,
		Password: &password,
	}.Build())
	if err != nil {
		return fmt.Errorf("login: %w", err)
	}

	state, err := u.store.Load()
	if err != nil {
		return err
	}

	state.Token = resp.GetToken()

	if err := u.store.Save(state); err != nil {
		return err
	}

	return nil
}
