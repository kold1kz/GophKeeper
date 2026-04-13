package usecase

import (
	"context"
	"net"
	"path/filepath"
	"testing"

	grpcclient "gophkeeper/internal/client/grpc"
	"gophkeeper/internal/client/local"
	pb "gophkeeper/proto"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

type testGophKeeperServer struct {
	pb.UnimplementedGophKeeperServiceServer
}

func (s *testGophKeeperServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	token := "token-123"

	return pb.LoginResponse_builder{
		Token: proto.String(token),
	}.Build(), nil
}

func startTestGRPCServer(t *testing.T, srv pb.GophKeeperServiceServer) (addr string, cleanup func()) {
	t.Helper()

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen returned error: %v", err)
	}

	grpcSrv := grpc.NewServer()
	pb.RegisterGophKeeperServiceServer(grpcSrv, srv)

	go func() {
		_ = grpcSrv.Serve(lis)
	}()

	return lis.Addr().String(), func() {
		grpcSrv.Stop()
		_ = lis.Close()
	}
}

func TestLoginUseCase_Integration(t *testing.T) {
	addr, cleanup := startTestGRPCServer(t, &testGophKeeperServer{})
	defer cleanup()

	dir := t.TempDir()
	store := local.NewStore(filepath.Join(dir, "state.json"))

	client, err := grpcclient.New(addr)
	if err != nil {
		t.Fatalf("grpcclient.New returned error: %v", err)
	}
	defer client.Close()

	uc := NewLoginUseCase(store, client)

	err = uc.Execute(context.Background(), "u1", "p1")
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	state, err := store.Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if state.Token != "token-123" {
		t.Fatalf("expected token token-123, got %q", state.Token)
	}
}
