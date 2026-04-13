package grpcclient

import (
	"net"
	"testing"

	pb "gophkeeper/proto"

	"google.golang.org/grpc"
)

type testServer struct {
	pb.UnimplementedGophKeeperServiceServer
}

func TestNew(t *testing.T) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen returned error: %v", err)
	}
	defer lis.Close()

	srv := grpc.NewServer()
	pb.RegisterGophKeeperServiceServer(srv, &testServer{})
	defer srv.Stop()

	go func() {
		_ = srv.Serve(lis)
	}()

	client, err := New(lis.Addr().String())
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	if client == nil {
		t.Fatal("expected client")
	}
	defer client.Close()

	if client.Raw() == nil {
		t.Fatal("expected raw client")
	}
}
