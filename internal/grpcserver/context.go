// Package grpcserver содержит реализацию gRPC-сервера GophKeeper.
//
// Пакет включает:
//   - обработчики gRPC-методов;
//   - интерсепторы аутентификации и логирования;
//   - преобразование protobuf-моделей в доменные структуры и обратно;
//   - работу с пользовательским идентификатором в context.Context.
package grpcserver

import "context"

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

// validUserIDFromContext извлекает непустой идентификатор пользователя из контекста.
func validUserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := userIDFromContext(ctx)
	if !ok || userID == "" {
		return "", false
	}

	return userID, true
}
