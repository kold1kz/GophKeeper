// Package grpcclient содержит клиент для взаимодействия с gRPC-сервером
// GophKeeper.
//
// Пакет отвечает за:
//   - установку gRPC-соединения;
//   - доступ к сгенерированному protobuf-клиенту;
//   - корректное закрытие соединения;
//   - добавление токена авторизации в контекст запроса.
package grpcclient

import (
	"context"
	"fmt"

	pb "gophkeeper/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// Client представляет обертку над gRPC-соединением и protobuf-клиентом.
//
// Используется CLI-приложением для выполнения запросов к серверу.
type Client struct {
	conn   *grpc.ClientConn
	client pb.GophKeeperServiceClient
}

// New создает новый gRPC-клиент для подключения к серверу по указанному адресу.
//
// Возвращает инициализированный клиент или ошибку, если соединение не удалось
// создать.
func New(address string) (*Client, error) {
	conn, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("create grpc client: %w", err)
	}

	return &Client{
		conn:   conn,
		client: pb.NewGophKeeperServiceClient(conn),
	}, nil
}

// Close закрывает gRPC-соединение клиента.
//
// Если клиент или соединение не инициализированы, функция завершится без ошибки.
func (c *Client) Close() error {
	if c == nil || c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

// Raw возвращает сгенерированный protobuf-клиент.
//
// Используется use case-слоем для выполнения RPC-вызовов.
func (c *Client) Raw() pb.GophKeeperServiceClient {
	return c.client
}

// ContextWithBearerToken добавляет Bearer-токен в исходящий gRPC-контекст.
//
// Функция используется для передачи токена авторизации в metadata
// при выполнении защищенных RPC-запросов.
func ContextWithBearerToken(ctx context.Context, token string) context.Context {
	md := metadata.Pairs("authorization", "Bearer "+token)
	return metadata.NewOutgoingContext(ctx, md)
}
