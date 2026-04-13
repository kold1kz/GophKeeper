// Package grpcserver содержит реализацию gRPC-сервера GophKeeper.
package grpcserver

import (
	"context"
	"errors"
	"strconv"

	"gophkeeper/internal/repository"
	"gophkeeper/internal/service"
	pb "gophkeeper/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

// Server реализует gRPC-сервис GophKeeper.
//
// Сервер делегирует бизнес-логику сервисному слою:
//   - регистрацию;
//   - вход пользователя;
//   - работу с приватными данными.
type Server struct {
	pb.UnimplementedGophKeeperServiceServer
	loginService    service.LoginService
	registerService service.RegisterService
	vaultService    service.VaultService
}

// NewServer создаёт новый экземпляр gRPC-сервера GophKeeper.
func NewServer(
	registerSvc service.RegisterService,
	loginSvc service.LoginService,
	vaultSvc service.VaultService,
) *Server {
	return &Server{
		registerService: registerSvc,
		loginService:    loginSvc,
		vaultService:    vaultSvc,
	}
}

// Register обрабатывает запрос регистрации нового пользователя.
//
// При успешной регистрации возвращает идентификатор созданного пользователя.
// Ошибки сервисного слоя преобразуются в соответствующие gRPC-статусы.
func (s *Server) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	if req.GetLogin() == "" || req.GetPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "empty register data")
	}

	userID, err := s.registerService.Register(ctx, req.GetLogin(), req.GetPassword())
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidRegisterData):
			return nil, status.Error(codes.InvalidArgument, "invalid register data")
		case errors.Is(err, repository.ErrUserAlreadyExists):
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		default:
			return nil, status.Error(codes.Internal, "register failed")
		}
	}

	return pb.RegisterResponse_builder{
		UserId: proto.String(strconv.Itoa(userID)),
	}.Build(), nil
}

// Login обрабатывает запрос аутентификации пользователя.
//
// При успешной проверке логина и пароля возвращает токен доступа.
// Ошибки сервисного слоя преобразуются в соответствующие gRPC-статусы.
func (s *Server) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	if req.GetLogin() == "" || req.GetPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "empty credentials")
	}

	token, err := s.loginService.Login(ctx, req.GetLogin(), req.GetPassword())
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidCredentials):
			return nil, status.Error(codes.Unauthenticated, "invalid credentials")
		default:
			return nil, status.Error(codes.Internal, "login failed")
		}
	}

	return pb.LoginResponse_builder{
		Token: proto.String(token),
	}.Build(), nil
}
