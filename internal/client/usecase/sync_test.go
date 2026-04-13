package usecase

import (
	"context"
	"net"
	"path/filepath"
	"testing"
	"time"

	grpcclient "gophkeeper/internal/client/grpc"
	"gophkeeper/internal/client/local"
	pb "gophkeeper/proto"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type testSyncServer struct {
	pb.UnimplementedGophKeeperServiceServer
}

func (s *testSyncServer) SyncItems(ctx context.Context, req *pb.SyncItemsRequest) (*pb.SyncItemsResponse, error) {
	id := "item-1"
	title := "note1"
	meta := "meta1"
	hash := "hash1"
	version := int64(1)
	now := time.Now().UTC().Round(0)

	item := pb.VaultItem_builder{
		Id:               proto.String(id),
		Type:             pb.ItemType_ITEM_TYPE_TEXT.Enum(),
		Title:            proto.String(title),
		Meta:             proto.String(meta),
		PayloadEncrypted: []byte("cipher"),
		PayloadNonce:     []byte("nonce"),
		PayloadHash:      proto.String(hash),
		Version:          proto.Int64(version),
		CreatedAt:        timestamppb.New(now),
		UpdatedAt:        timestamppb.New(now),
	}.Build()

	return pb.SyncItemsResponse_builder{
		Items:      []*pb.VaultItem{item},
		ServerTime: timestamppb.New(now),
	}.Build(), nil
}

func startSyncTestServer(t *testing.T, srv pb.GophKeeperServiceServer) (string, func()) {
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

func TestSyncUseCase_Integration(t *testing.T) {
	addr, cleanup := startSyncTestServer(t, &testSyncServer{})
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

	uc := NewSyncUseCase(store, client)

	err = uc.Execute(context.Background())
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	state, err := store.Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if len(state.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(state.Items))
	}
	if state.Items[0].ID != "item-1" {
		t.Fatalf("expected item-1, got %q", state.Items[0].ID)
	}
	if state.LastSyncAt == nil {
		t.Fatal("expected LastSyncAt to be set")
	}
}
