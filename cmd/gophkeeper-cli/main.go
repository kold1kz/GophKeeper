// Команда gophkeeper-cli представляет собой CLI-клиент для работы с
// менеджером паролей GophKeeper.
//
// Клиент позволяет:
//   - регистрировать пользователя;
//   - выполнять вход;
//   - синхронизировать данные с сервером;
//   - создавать, просматривать, обновлять и удалять записи;
//   - работать с локальным зашифрованным хранилищем;
//   - отображать информацию о сборке.
package main

import (
	"context"
	"fmt"
	"os"

	clientconfig "gophkeeper/internal/client/config"
	grpcclient "gophkeeper/internal/client/grpc"
	"gophkeeper/internal/client/local"
	"gophkeeper/internal/client/usecase"
)

var (
	// buildVersion содержит версию сборки клиента.
	buildVersion string
	// buildDate содержит дату сборки клиента.
	buildDate string
	// buildCommit содержит идентификатор коммита клиента.
	buildCommit string
)

// na возвращает строку "N/A", если значение пустое.
//
// Используется для вывода метаданных сборки.
func na(s string) string {
	if s == "" {
		return "N/A"
	}
	return s
}

// main является точкой входа CLI-приложения.
//
// Выполняет запуск клиентской логики и завершает процесс с кодом 1
// при возникновении ошибки.
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

// run загружает конфигурацию клиента, создает зависимости и обрабатывает
// переданную CLI-команду.
//
// Возвращает ошибку, если команда выполнена некорректно или произошел сбой
// во время выполнения.
func run() error {
	cfg, err := clientconfig.Load()
	if err != nil {
		return err
	}

	if len(os.Args) < 2 {
		printUsage()
		return nil
	}

	if os.Args[1] == "version" {
		fmt.Printf("Build version: %s\n", na(buildVersion))
		fmt.Printf("Build date: %s\n", na(buildDate))
		fmt.Printf("Build commit: %s\n", na(buildCommit))
		return nil
	}

	grpcCli, err := grpcclient.New(cfg.ServerAddress)
	if err != nil {
		return err
	}
	defer grpcCli.Close()

	store := local.NewStore(cfg.StatePath)
	ctx := context.Background()

	switch os.Args[1] {
	case "register":
		if len(os.Args) != 4 {
			return fmt.Errorf("usage: gophkeeper-cli register <login> <password>")
		}

		uc := usecase.NewRegisterUseCase(grpcCli)
		userID, err := uc.Execute(ctx, os.Args[2], os.Args[3])
		if err != nil {
			return err
		}

		fmt.Println("registered, user_id:", userID)
		return nil

	case "login":
		if len(os.Args) != 4 {
			return fmt.Errorf("usage: gophkeeper-cli login <login> <password>")
		}

		uc := usecase.NewLoginUseCase(store, grpcCli)
		if err := uc.Execute(ctx, os.Args[2], os.Args[3]); err != nil {
			return err
		}

		fmt.Println("login successful")
		return nil

	case "sync":
		uc := usecase.NewSyncUseCase(store, grpcCli)
		if err := uc.Execute(ctx); err != nil {
			return err
		}

		fmt.Println("sync successful")
		return nil

	case "create-text":
		if len(os.Args) < 5 || len(os.Args) > 6 {
			return fmt.Errorf("usage: gophkeeper-cli create-text <title> <text> <master-password> [meta]")
		}

		meta := ""
		if len(os.Args) == 6 {
			meta = os.Args[5]
		}

		uc := usecase.NewCreateTextItemUseCase(store, grpcCli)
		itemID, err := uc.Execute(ctx, os.Args[2], os.Args[3], meta, os.Args[4])
		if err != nil {
			return err
		}

		fmt.Println("text item created, id:", itemID)
		return nil

	case "list-local":
		uc := usecase.NewListLocalItemsUseCase(store)
		if err := uc.Execute(); err != nil {
			return err
		}
		return nil

	case "get-local":
		if len(os.Args) != 4 {
			return fmt.Errorf("usage: gophkeeper-cli get-local <id> <master-password>")
		}

		uc := usecase.NewGetLocalItemUseCase(store)
		if err := uc.Execute(os.Args[2], os.Args[3]); err != nil {
			return err
		}

		return nil

	case "list-remote":
		uc := usecase.NewListRemoteItemsUseCase(store, grpcCli)
		if err := uc.Execute(ctx); err != nil {
			return err
		}
		return nil

	case "delete-item":
		if len(os.Args) != 3 {
			return fmt.Errorf("usage: gophkeeper-cli delete-item <id>")
		}

		uc := usecase.NewDeleteItemUseCase(store, grpcCli)
		if err := uc.Execute(ctx, os.Args[2]); err != nil {
			return err
		}

		return nil

	case "update-text":
		if len(os.Args) < 6 || len(os.Args) > 7 {
			return fmt.Errorf("usage: gophkeeper-cli update-text <id> <title> <text> <master-password> [meta]")
		}

		meta := ""
		if len(os.Args) == 7 {
			meta = os.Args[6]
		}

		uc := usecase.NewUpdateTextItemUseCase(store, grpcCli)
		if err := uc.Execute(ctx, os.Args[2], os.Args[3], os.Args[4], meta, os.Args[5]); err != nil {
			return err
		}

		return nil

	case "create-login":
		if len(os.Args) < 6 || len(os.Args) > 7 {
			return fmt.Errorf("usage: gophkeeper-cli create-login <title> <login> <password> <master-password> [meta]")
		}

		meta := ""
		if len(os.Args) == 7 {
			meta = os.Args[6]
		}

		uc := usecase.NewCreateLoginPasswordItemUseCase(store, grpcCli)
		itemID, err := uc.Execute(
			ctx,
			os.Args[2],
			os.Args[3],
			os.Args[4],
			meta,
			os.Args[5],
		)
		if err != nil {
			return err
		}

		fmt.Println("login/password item created, id:", itemID)
		return nil

	case "create-card":
		if len(os.Args) < 8 || len(os.Args) > 9 {
			return fmt.Errorf("usage: gophkeeper-cli create-card <title> <number> <holder> <expiry> <cvv> <master-password> [meta]")
		}

		meta := ""
		if len(os.Args) == 9 {
			meta = os.Args[8]
		}

		uc := usecase.NewCreateBankCardItemUseCase(store, grpcCli)

		id, err := uc.Execute(
			ctx,
			os.Args[2], // title
			os.Args[3], // number
			os.Args[4], // holder
			os.Args[5], // expiry
			os.Args[6], // cvv
			meta,
			os.Args[7], // master password
		)
		if err != nil {
			return err
		}

		fmt.Println("bank card created, id:", id)
		return nil

	case "create-file":
		if len(os.Args) < 5 || len(os.Args) > 6 {
			return fmt.Errorf("usage: gophkeeper-cli create-file <title> <file-path> <master-password> [meta]")
		}

		meta := ""
		if len(os.Args) == 6 {
			meta = os.Args[5]
		}

		uc := usecase.NewCreateBinaryItemUseCase(store, grpcCli)
		itemID, err := uc.Execute(ctx, os.Args[2], os.Args[3], meta, os.Args[4])
		if err != nil {
			return err
		}

		fmt.Println("binary item created, id:", itemID)
		return nil

	case "get-file":
		if len(os.Args) != 5 {
			return fmt.Errorf("usage: gophkeeper-cli get-file <id> <master-password> <output-dir>")
		}

		uc := usecase.NewGetBinaryItemUseCase(store)
		if err := uc.Execute(os.Args[2], os.Args[3], os.Args[4]); err != nil {
			return err
		}

		return nil

	default:
		printUsage()
		return nil
	}
}

// printUsage выводит справку по поддерживаемым CLI-командам.
func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  gophkeeper-cli register <login> <password>")
	fmt.Println("  gophkeeper-cli login <login> <password>")
	fmt.Println("  gophkeeper-cli sync")
	fmt.Println("  gophkeeper-cli create-text <title> <text> <master-password> [meta]")
	fmt.Println("  gophkeeper-cli get-local <id> <master-password>")
	fmt.Println("  gophkeeper-cli update-text <id> <title> <text> <master-password> [meta]")
	fmt.Println("  gophkeeper-cli list-local")
	fmt.Println("  gophkeeper-cli list-remote")
	fmt.Println("  gophkeeper-cli delete-item <id>")
	fmt.Println("  gophkeeper-cli create-login <title> <login> <password> <master-password> [meta]")
	fmt.Println("  gophkeeper-cli create-card <title> <number> <holder> <expiry> <cvv> <master-password> [meta]")
	fmt.Println("  gophkeeper-cli create-file <title> <file-path> <master-password> [meta]")
	fmt.Println("  gophkeeper-cli get-file <id> <master-password> <output-dir>")

	fmt.Println("  gophkeeper-cli version")
}
