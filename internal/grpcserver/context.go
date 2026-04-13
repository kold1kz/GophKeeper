package grpcserver

import (
	"context"
	"strconv"
)

type contextKey string

const userIDContextKey contextKey = "userID"

func withUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDContextKey, userID)
}

func userIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userIDContextKey).(string)
	return userID, ok
}

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
