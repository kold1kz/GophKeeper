package usecase

import (
	"context"
	"net"
	"testing"

	grpcclient "gophkeeper/internal/client/grpc"
	pb "gophkeeper/proto"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

type testRegisterServer struct {
	pb.UnimplementedGophKeeperServiceServer
}

func (s *testRegisterServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	id := "42"

	return pb.RegisterResponse_builder{
		UserId: proto.String(id),
	}.Build(), nil
}

func startRegisterTestServer(t *testing.T, srv pb.GophKeeperServiceServer) (string, func()) {
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

func TestRegisterUseCase_Integration(t *testing.T) {
	addr, cleanup := startRegisterTestServer(t, &testRegisterServer{})
	defer cleanup()

	client, err := grpcclient.New(addr)
	if err != nil {
		t.Fatalf("grpcclient.New returned error: %v", err)
	}
	defer client.Close()

	uc := NewRegisterUseCase(client)

	id, err := uc.Execute(context.Background(), "u1", "p1")
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if id != "42" {
		t.Fatalf("expected id 42, got %q", id)
	}
}
