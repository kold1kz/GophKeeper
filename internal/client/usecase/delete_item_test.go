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
	"google.golang.org/protobuf/types/known/timestamppb"
)

type testDeleteItemServer struct {
	pb.UnimplementedGophKeeperServiceServer
}

func (s *testDeleteItemServer) DeleteItem(ctx context.Context, req *pb.DeleteItemRequest) (*pb.DeleteItemResponse, error) {
	id := req.GetId()

	return pb.DeleteItemResponse_builder{
		Id:        proto.String(id),
		DeletedAt: timestamppb.Now(),
	}.Build(), nil
}

func startDeleteItemTestServer(t *testing.T, srv pb.GophKeeperServiceServer) (string, func()) {
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

func TestDeleteItemUseCase_Integration(t *testing.T) {
	addr, cleanup := startDeleteItemTestServer(t, &testDeleteItemServer{})
	defer cleanup()

	dir := t.TempDir()
	store := local.NewStore(filepath.Join(dir, "state.json"))

	err := store.Save(&local.ClientState{
		Token: "token-123",
	})
	if err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	client, err := grpcclient.New(addr)
	if err != nil {
		t.Fatalf("grpcclient.New returned error: %v", err)
	}
	defer client.Close()

	uc := NewDeleteItemUseCase(store, client)

	err = uc.Execute(context.Background(), "item-123")
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
}
