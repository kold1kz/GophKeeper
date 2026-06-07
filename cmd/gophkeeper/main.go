// Команда gophkeeper запускает сервер менеджера паролей GophKeeper.
//
// Сервер предоставляет gRPC API для:
//   - регистрации пользователей;
//   - аутентификации;
//   - хранения приватных данных;
//   - синхронизации данных между клиентами.
//
// При включенном HTTPS сервер автоматически подготавливает TLS-сертификаты
// и запускается в защищенном режиме.
package main

import (
	"context"
	"fmt"
	"gophkeeper/internal/middleware"
	"gophkeeper/internal/repository"
	"gophkeeper/internal/service"
	pb "gophkeeper/proto"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"gophkeeper/cmd/certutil"
	"gophkeeper/internal/config"
	"gophkeeper/internal/grpcserver"
	//"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	// buildVersion содержит версию сборки приложения.
	buildVersion string
	// buildDate содержит дату сборки приложения.
	buildDate string
	// buildCommit содержит идентификатор коммита сборки.
	buildCommit string
)

// na возвращает строку "N/A", если переданное значение пустое.
//
// Используется для безопасного отображения метаданных сборки.
func na(s string) string {
	if s == "" {
		return "N/A"
	}
	return s
}

// loadConfig загружает и валидирует конфигурацию сервера.
//
// Возвращает готовую конфигурацию или ошибку, если параметры некорректны.
func loadConfig() (*config.Config, error) {
	cfg := config.Init()
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}
	return cfg, nil
}

// main является точкой входа серверного приложения.
//
// Загружает переменные окружения из .env и запускает сервер.
// В случае ошибки завершает процесс с ненулевым кодом выхода.
func main() {
	_ = godotenv.Load()

	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// run инициализирует зависимости и запускает gRPC-сервер.
//
// Функция выполняет:
//   - загрузку конфигурации;
//   - инициализацию логгера;
//   - создание репозиториев и сервисов;
//   - запуск gRPC-сервера;
//   - корректное завершение по системному сигналу.
//
// Возвращает ошибку, если сервер не удалось запустить или произошла ошибка
// во время его работы.
func run() error {
	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	defer cfg.Close()

	logger := middleware.InitLogger()
	defer logger.Sync()
	logger.Infow("build info",
		"version", na(buildVersion),
		"date", na(buildDate),
		"commit", na(buildCommit),
	)

	if cfg.GRPCServerAddress == "" {
		return fmt.Errorf("grpc server address cannot be empty")
	}

	if cfg.DB == nil {
		return fmt.Errorf("grpc login requires database connection")
	}

	userRepo := repository.NewPostgresUserRepository(cfg.DB.GetPool())
	vaultRepo := repository.NewPostgresVaultRepository(cfg.DB.GetPool())

	registerSvc := service.NewRegisterService(userRepo)
	loginSvc := service.NewLoginService(userRepo)
	vaultSvc := service.NewVaultService(vaultRepo)

	lis, err := net.Listen("tcp", cfg.GRPCServerAddress)
	if err != nil {
		return fmt.Errorf("grpc listen: %w", err)
	}

	serverOpts := make([]grpc.ServerOption, 0, 2)
	serverOpts = append(serverOpts,
		grpc.ChainUnaryInterceptor(
			grpcserver.AuthInterceptor,
			grpcserver.LoggingInterceptor(logger),
		),
	)

	if cfg.EnableHTTPS {
		certPath, keyPath, err := certutil.EnsureCertFiles(certutil.EnsureOptions{
			CertPath:    "./cert/cert.pem",
			KeyPath:     "./cert/key.pem",
			ValidFor:    30 * 24 * time.Hour,
			RenewBefore: 24 * time.Hour,
			Hosts:       []string{"localhost", "127.0.0.1", "::1"},
		})
		if err != nil {
			return fmt.Errorf("ensure tls certs: %w", err)
		}

		creds, err := credentials.NewServerTLSFromFile(certPath, keyPath)
		if err != nil {
			return fmt.Errorf("grpc tls creds: %w", err)
		}

		serverOpts = append(serverOpts, grpc.Creds(creds))
		logger.Infow("grpc server starting with TLS", "address", cfg.GRPCServerAddress)
	} else {
		logger.Infow("grpc server starting without TLS", "address", cfg.GRPCServerAddress)
	}

	grpcSrv := grpc.NewServer(serverOpts...)
	pb.RegisterGophKeeperServiceServer(
		grpcSrv,
		grpcserver.NewServer(registerSvc, loginSvc, vaultSvc))

	errCh := make(chan error, 1)
	go func() {
		errCh <- grpcSrv.Serve(lis)
	}()

	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGTERM,
		syscall.SIGINT,
		syscall.SIGQUIT,
	)
	defer stop()

	select {
	case <-ctx.Done():
		logger.Infow("shutdown signal received")

		stopped := make(chan struct{})
		go func() {
			grpcSrv.GracefulStop()
			close(stopped)
		}()

		select {
		case <-stopped:
			logger.Infow("grpc server stopped gracefully")
			return nil
		case <-time.After(10 * time.Second):
			logger.Warnw("grpc graceful stop timeout exceeded, forcing stop")
			grpcSrv.Stop()
			return nil
		}

	case err := <-errCh:
		if err == nil {
			return nil
		}
		return fmt.Errorf("grpc server run error: %w", err)
	}
}
