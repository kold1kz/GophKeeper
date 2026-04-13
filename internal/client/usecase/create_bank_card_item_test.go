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

type testCreateBankCardServer struct {
	pb.UnimplementedGophKeeperServiceServer
}

func (s *testCreateBankCardServer) CreateItem(ctx context.Context, req *pb.CreateItemRequest) (*pb.CreateItemResponse, error) {
	id := "card-item-1"
	title := req.GetTitle()
	meta := req.GetMeta()
	hash := req.GetPayloadHash()
	version := int64(1)

	item := pb.VaultItem_builder{
		Id:               proto.String(id),
		Type:             req.GetType().Enum(),
		Title:            proto.String(title),
		Meta:             proto.String(meta),
		PayloadEncrypted: req.GetPayloadEncrypted(),
		PayloadNonce:     req.GetPayloadNonce(),
		PayloadHash:      proto.String(hash),
		Version:          proto.Int64(version),
	}.Build()

	return pb.CreateItemResponse_builder{
		Item: item,
	}.Build(), nil
}

func startCreateBankCardTestServer(t *testing.T, srv pb.GophKeeperServiceServer) (string, func()) {
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

func TestCreateBankCardItemUseCase_Integration(t *testing.T) {
	addr, cleanup := startCreateBankCardTestServer(t, &testCreateBankCardServer{})
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

	uc := NewCreateBankCardItemUseCase(store, client)

	itemID, err := uc.Execute(
		context.Background(),
		"visa",
		"1234 5678 9012 3456",
		"John Doe",
		"12/28",
		"123",
		"primary card",
		"master-password",
	)
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if itemID != "card-item-1" {
		t.Fatalf("expected card-item-1, got %q", itemID)
	}

	state, err := store.Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if len(state.KeySalt) == 0 {
		t.Fatal("expected KeySalt to be saved")
	}
}
