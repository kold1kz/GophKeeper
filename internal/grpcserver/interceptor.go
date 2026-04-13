package grpcserver

import (
	"context"
	"strings"
	"time"

	"gophkeeper/internal/auth"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func AuthInterceptor(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (any, error) {
	switch info.FullMethod {
	case "/gophkeeper.GophKeeperService/Login",
		"/gophkeeper.GophKeeperService/Register":
		return handler(ctx, req)
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing authorization metadata")
	}

	values := md.Get("authorization")
	if len(values) == 0 || values[0] == "" {
		return nil, status.Error(codes.Unauthenticated, "missing authorization token")
	}

	authHeader := values[0]
	const prefix = "Bearer "
	if !strings.HasPrefix(authHeader, prefix) {
		return nil, status.Error(codes.Unauthenticated, "invalid authorization format")
	}

	token := strings.TrimPrefix(authHeader, prefix)

	userID, err := auth.ParseToken(token)
	if err != nil || userID == "" {
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}

	return handler(withUserID(ctx, userID), req)
}

func LoggingInterceptor(logger *zap.SugaredLogger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		start := time.Now()

		resp, err := handler(ctx, req)
		duration := time.Since(start)

		if err != nil {
			if st, ok := status.FromError(err); ok {
				logger.Errorw(
					"grpc request failed",
					"method", info.FullMethod,
					"code", st.Code().String(),
					"message", st.Message(),
					"duration", duration,
				)
			} else {
				logger.Errorw(
					"grpc request failed",
					"method", info.FullMethod,
					"error", err.Error(),
					"duration", duration,
				)
			}
			return resp, err
		}

		logger.Infow(
			"grpc request completed",
			"method", info.FullMethod,
			"duration", duration,
		)

		return resp, nil
	}
}
