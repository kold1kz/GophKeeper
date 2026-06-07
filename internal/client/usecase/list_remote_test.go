package usecase

import (
	"context"
	"net"
	"path/filepath"
	"strings"
	"testing"

	grpcclient "gophkeeper/internal/client/grpc"
	"gophkeeper/internal/client/local"
	pb "gophkeeper/proto"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

type testListRemoteServer struct {
	pb.UnimplementedGophKeeperServiceServer
}

func (s *testListRemoteServer) ListItems(ctx context.Context, req *pb.ListItemsRequest) (*pb.ListItemsResponse, error) {
	id := "remote-item-1"
	title := "remote-note"
	meta := "remote-meta"
	hash := "hash1"
	version := int64(1)

	item := pb.VaultItem_builder{
		Id:               proto.String(id),
		Type:             pb.ItemType_ITEM_TYPE_TEXT.Enum(),
		Title:            proto.String(title),
		Meta:             proto.String(meta),
		PayloadEncrypted: []byte("cipher"),
		PayloadNonce:     []byte("nonce"),
		PayloadHash:      proto.String(hash),
		Version:          proto.Int64(version),
	}.Build()

	return pb.ListItemsResponse_builder{
		Items: []*pb.VaultItem{item},
	}.Build(), nil
}

func startListRemoteTestServer(t *testing.T, srv pb.GophKeeperServiceServer) (string, func()) {
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

func TestListRemoteItemsUseCase_Integration(t *testing.T) {
	addr, cleanup := startListRemoteTestServer(t, &testListRemoteServer{})
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

	uc := NewListRemoteItemsUseCase(store, client)

	output := captureStdout(t, func() {
		if err := uc.Execute(context.Background()); err != nil {
			t.Fatalf("Execute returned error: %v", err)
		}
	})

	if !strings.Contains(output, "remote-item-1") {
		t.Fatalf("expected item id in output, got %q", output)
	}
	if !strings.Contains(output, "remote-note") {
		t.Fatalf("expected title in output, got %q", output)
	}
}
