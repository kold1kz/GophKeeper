package usecase

import (
	"context"
	"fmt"

	grpcclient "gophkeeper/internal/client/grpc"
	pb "gophkeeper/proto"
)

type RegisterUseCase struct {
	grpc *grpcclient.Client
}

func NewRegisterUseCase(grpc *grpcclient.Client) *RegisterUseCase {
	return &RegisterUseCase{grpc: grpc}
}

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
