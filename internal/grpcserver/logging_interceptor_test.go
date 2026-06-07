package grpcserver

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
	"google.golang.org/grpc"
)

func TestLoggingInterceptor_Success(t *testing.T) {
	core, recorded := observer.New(zap.InfoLevel)
	logger := zap.New(core).Sugar()

	interceptor := LoggingInterceptor(logger)

	resp, err := interceptor(
		context.Background(),
		"req",
		&grpc.UnaryServerInfo{FullMethod: "/gophkeeper.GophKeeperService/Login"},
		func(ctx context.Context, req any) (any, error) {
			return "ok", nil
		},
	)
	if err != nil {
		t.Fatalf("interceptor returned error: %v", err)
	}
	if resp != "ok" {
		t.Fatalf("expected response ok, got %v", resp)
	}

	if recorded.Len() == 0 {
		t.Fatal("expected logs to be written")
	}
}

func TestLoggingInterceptor_Error(t *testing.T) {
	core, recorded := observer.New(zap.ErrorLevel)
	logger := zap.New(core).Sugar()

	interceptor := LoggingInterceptor(logger)
	wantErr := errors.New("boom")

	_, err := interceptor(
		context.Background(),
		"req",
		&grpc.UnaryServerInfo{FullMethod: "/gophkeeper.GophKeeperService/Login"},
		func(ctx context.Context, req any) (any, error) {
			return nil, wantErr
		},
	)
	if err == nil {
		t.Fatal("expected error")
	}

	if recorded.Len() == 0 {
		t.Fatal("expected error logs to be written")
	}
}
