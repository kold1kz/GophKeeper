package grpcclient

import (
	"context"
	"testing"

	pb "gophkeeper/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestContextWithBearerToken(t *testing.T) {
	t.Parallel()

	ctx := ContextWithBearerToken(context.Background(), "token-123")

	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		t.Fatal("expected outgoing metadata")
	}

	values := md.Get("authorization")
	if len(values) != 1 {
		t.Fatalf("expected one authorization value, got %d", len(values))
	}
	if values[0] != "Bearer token-123" {
		t.Fatalf("unexpected authorization header: %q", values[0])
	}
}

func TestRaw(t *testing.T) {
	t.Parallel()

	client := &Client{
		client: pb.NewGophKeeperServiceClient(&grpc.ClientConn{}),
	}

	if client.Raw() == nil {
		t.Fatal("expected non-nil raw client")
	}
}

func TestClose_NilConn(t *testing.T) {
	t.Parallel()

	client := &Client{}
	if err := client.Close(); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}
