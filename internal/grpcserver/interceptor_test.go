package grpcserver

import (
	"context"
	"testing"

	"gophkeeper/internal/auth"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestAuthInterceptor_PublicMethod(t *testing.T) {
	t.Parallel()

	called := false

	_, err := AuthInterceptor(
		context.Background(),
		nil,
		&grpc.UnaryServerInfo{FullMethod: "/gophkeeper.GophKeeperService/Login"},
		func(ctx context.Context, req any) (any, error) {
			called = true
			return "ok", nil
		},
	)
	if err != nil {
		t.Fatalf("AuthInterceptor returned error: %v", err)
	}
	if !called {
		t.Fatal("expected handler to be called")
	}
}

func TestAuthInterceptor_MissingAuthorization(t *testing.T) {
	t.Parallel()

	_, err := AuthInterceptor(
		metadata.NewIncomingContext(context.Background(), metadata.MD{}),
		nil,
		&grpc.UnaryServerInfo{FullMethod: "/gophkeeper.GophKeeperService/ListItems"},
		func(ctx context.Context, req any) (any, error) {
			return "ok", nil
		},
	)
	if err == nil {
		t.Fatal("expected error")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("expected grpc status error")
	}
	if st.Code() != codes.Unauthenticated {
		t.Fatalf("expected Unauthenticated, got %v", st.Code())
	}
}

func TestAuthInterceptor_InvalidFormat(t *testing.T) {
	t.Parallel()

	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "bad-token"))

	_, err := AuthInterceptor(
		ctx,
		nil,
		&grpc.UnaryServerInfo{FullMethod: "/gophkeeper.GophKeeperService/ListItems"},
		func(ctx context.Context, req any) (any, error) {
			return "ok", nil
		},
	)
	if err == nil {
		t.Fatal("expected error")
	}

	st, _ := status.FromError(err)
	if st.Code() != codes.Unauthenticated {
		t.Fatalf("expected Unauthenticated, got %v", st.Code())
	}
}

func TestAuthInterceptor_Success(t *testing.T) {
	t.Parallel()

	token, err := auth.NewTokenForUserID("123")
	if err != nil {
		t.Fatalf("NewTokenForUserID returned error: %v", err)
	}

	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Bearer "+token))

	var capturedUserID string
	_, err = AuthInterceptor(
		ctx,
		nil,
		&grpc.UnaryServerInfo{FullMethod: "/gophkeeper.GophKeeperService/ListItems"},
		func(ctx context.Context, req any) (any, error) {
			userID, ok := userIDFromContext(ctx)
			if !ok {
				t.Fatal("expected user id in context")
			}
			capturedUserID = userID
			return "ok", nil
		},
	)
	if err != nil {
		t.Fatalf("AuthInterceptor returned error: %v", err)
	}
	if capturedUserID != "123" {
		t.Fatalf("expected user id 123, got %q", capturedUserID)
	}
}
