package usecase

import (
	"context"
	"net"
	"path/filepath"
	"testing"
	"time"

	clientcrypto "gophkeeper/internal/client/crypto"
	grpcclient "gophkeeper/internal/client/grpc"
	"gophkeeper/internal/client/local"
	pb "gophkeeper/proto"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type testUpdateTextServer struct {
	pb.UnimplementedGophKeeperServiceServer
}

func (s *testUpdateTextServer) UpdateItem(ctx context.Context, req *pb.UpdateItemRequest) (*pb.UpdateItemResponse, error) {
	id := req.GetId()
	title := req.GetTitle()
	meta := req.GetMeta()
	hash := req.GetPayloadHash()
	version := req.GetVersion() + 1
	now := time.Now().UTC().Round(0)

	item := pb.VaultItem_builder{
		Id:               proto.String(id),
		Type:             pb.ItemType_ITEM_TYPE_TEXT.Enum(),
		Title:            proto.String(title),
		Meta:             proto.String(meta),
		PayloadEncrypted: req.GetPayloadEncrypted(),
		PayloadNonce:     req.GetPayloadNonce(),
		PayloadHash:      proto.String(hash),
		Version:          proto.Int64(version),
		CreatedAt:        timestamppb.New(now.Add(-time.Hour)),
		UpdatedAt:        timestamppb.New(now),
	}.Build()

	return pb.UpdateItemResponse_builder{
		Item: item,
	}.Build(), nil
}

func startUpdateTextTestServer(t *testing.T, srv pb.GophKeeperServiceServer) (string, func()) {
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

func TestUpdateTextItemUseCase_Integration(t *testing.T) {
	addr, cleanup := startUpdateTextTestServer(t, &testUpdateTextServer{})
	defer cleanup()

	dir := t.TempDir()
	store := local.NewStore(filepath.Join(dir, "state.json"))

	salt := []byte("0123456789abcdef")
	key, err := clientcrypto.DeriveKey("master-password", salt)
	if err != nil {
		t.Fatalf("DeriveKey returned error: %v", err)
	}

	ciphertext, nonce, err := clientcrypto.Encrypt(key, []byte("old text"))
	if err != nil {
		t.Fatalf("Encrypt returned error: %v", err)
	}

	now := time.Now().UTC().Round(0)
	err = store.Save(&local.ClientState{
		Token:   "token-123",
		KeySalt: salt,
		Items: []local.LocalItem{
			{
				ID:               "item-1",
				Type:             "ITEM_TYPE_TEXT",
				Title:            "note1",
				Meta:             "meta1",
				PayloadEncrypted: ciphertext,
				PayloadNonce:     nonce,
				PayloadHash:      "hash1",
				Version:          5,
				CreatedAt:        now.Add(-time.Hour),
				UpdatedAt:        now,
			},
		},
	})
	if err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	client, err := grpcclient.New(addr)
	if err != nil {
		t.Fatalf("grpcclient.New returned error: %v", err)
	}
	defer client.Close()

	uc := NewUpdateTextItemUseCase(store, client)

	err = uc.Execute(context.Background(), "item-1", "note1-updated", "new text", "meta-updated", "master-password")
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
}
