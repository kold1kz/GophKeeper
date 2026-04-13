// Package grpcserver содержит реализацию gRPC-сервера GophKeeper.
//
// Пакет включает:
//   - обработчики gRPC-методов;
//   - интерсепторы аутентификации и логирования;
//   - преобразование protobuf-моделей в доменные структуры и обратно;
//   - работу с пользовательским идентификатором в context.Context.
package grpcserver

import (
	"context"
	"strconv"
)

type contextKey string

const userIDContextKey contextKey = "userID"

// withUserID сохраняет идентификатор пользователя в контексте запроса.
func withUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDContextKey, userID)
}

// userIDFromContext извлекает идентификатор пользователя из контекста.
//
// Возвращает строковое значение идентификатора и признак успешного извлечения.
func userIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userIDContextKey).(string)
	return userID, ok
}

// userIDInt64FromContext извлекает идентификатор пользователя из контекста
// и преобразует его в тип int64.
//
// Если идентификатор отсутствует, пустой или не является корректным числом,
// функция возвращает 0 и false.
func userIDInt64FromContext(ctx context.Context) (int64, bool) {
	userID, ok := userIDFromContext(ctx)
	if !ok || userID == "" {
		return 0, false
	}

	id, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return 0, false
	}

	return id, true
}
