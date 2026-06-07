package grpcserver

import (
	"testing"
	"time"

	"gophkeeper/internal/model"
	pb "gophkeeper/proto"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestProtoItemTypeToModel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		in      pb.ItemType
		want    model.ItemType
		wantErr bool
	}{
		{
			name: "text",
			in:   pb.ItemType_ITEM_TYPE_TEXT,
			want: model.ItemTypeText,
		},
		{
			name: "login_password",
			in:   pb.ItemType_ITEM_TYPE_LOGIN_PASSWORD,
			want: model.ItemTypeLoginPassword,
		},
		{
			name: "binary",
			in:   pb.ItemType_ITEM_TYPE_BINARY,
			want: model.ItemTypeBinary,
		},
		{
			name: "bank_card",
			in:   pb.ItemType_ITEM_TYPE_BANK_CARD,
			want: model.ItemTypeBankCard,
		},
		{
			name:    "unsupported",
			in:      pb.ItemType_ITEM_TYPE_UNSPECIFIED,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := protoItemTypeToModel(tt.in)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}

			if err != nil {
				t.Fatalf("protoItemTypeToModel returned error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestModelItemTypeToProto(t *testing.T) {
	t.Parallel()

	if got := modelItemTypeToProto(model.ItemTypeText); got != pb.ItemType_ITEM_TYPE_TEXT {
		t.Fatalf("expected ITEM_TYPE_TEXT, got %v", got)
	}

	if got := modelItemTypeToProto(model.ItemType("unknown")); got != pb.ItemType_ITEM_TYPE_UNSPECIFIED {
		t.Fatalf("expected ITEM_TYPE_UNSPECIFIED, got %v", got)
	}
}

func TestCreateItemRequestToModel(t *testing.T) {
	t.Parallel()

	title := "note1"
	meta := "meta1"
	payload := []byte("cipher")
	nonce := []byte("nonce")
	hash := "hash"
	clientUpdatedAt := timestamppb.New(time.Now().UTC())

	req := pb.CreateItemRequest_builder{
		Type:             pb.ItemType_ITEM_TYPE_TEXT.Enum(),
		Title:            &title,
		Meta:             &meta,
		PayloadEncrypted: payload,
		PayloadNonce:     nonce,
		PayloadHash:      &hash,
		ClientUpdatedAt:  clientUpdatedAt,
	}.Build()

	item, err := createItemRequestToModel("10", req)
	if err != nil {
		t.Fatalf("createItemRequestToModel returned error: %v", err)
	}

	if item.UserID != "10" {
		t.Fatalf("expected userID 10, got %s", item.UserID)
	}
	if item.Type != model.ItemTypeText {
		t.Fatalf("expected type %q, got %q", model.ItemTypeText, item.Type)
	}
	if item.Title != title {
		t.Fatalf("expected title %q, got %q", title, item.Title)
	}
	if item.Meta != meta {
		t.Fatalf("expected meta %q, got %q", meta, item.Meta)
	}
	if string(item.PayloadEncrypted) != string(payload) {
		t.Fatal("payload mismatch")
	}
	if string(item.PayloadNonce) != string(nonce) {
		t.Fatal("nonce mismatch")
	}
	if item.PayloadHash != hash {
		t.Fatalf("expected hash %q, got %q", hash, item.PayloadHash)
	}
	if item.ClientUpdatedAt == nil {
		t.Fatal("expected ClientUpdatedAt")
	}
}

func TestUpdateItemRequestToModel(t *testing.T) {
	t.Parallel()

	id := "7d444840-9dc0-11d1-b245-5ffdce74fad2"
	title := "title"
	meta := "meta"
	payload := []byte("cipher")
	nonce := []byte("nonce")
	hash := "hash"
	version := int64(7)

	req := pb.UpdateItemRequest_builder{
		Id:               &id,
		Title:            &title,
		Meta:             &meta,
		PayloadEncrypted: payload,
		PayloadNonce:     nonce,
		PayloadHash:      &hash,
		Version:          &version,
	}.Build()

	item := updateItemRequestToModel("15", req)

	if item.ID != uuid.MustParse(id) {
		t.Fatalf("expected id %q, got %q", id, item.ID)
	}
	if item.UserID != "15" {
		t.Fatalf("expected userID 15, got %s", item.UserID)
	}
	if item.Version != version {
		t.Fatalf("expected version %d, got %d", version, item.Version)
	}
}

func TestModelItemToProto(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC().Round(0)
	item := &model.VaultItem{
		ID:               uuid.MustParse("7d444840-9dc0-11d1-b245-5ffdce74fad2"),
		Type:             model.ItemTypeText,
		Title:            "note1",
		Meta:             "meta1",
		PayloadEncrypted: []byte("cipher"),
		PayloadNonce:     []byte("nonce"),
		PayloadHash:      "hash",
		Version:          3,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	got := modelItemToProto(item)
	if got == nil {
		t.Fatal("expected non-nil proto item")
	}
	if got.GetId() != item.ID.String() {
		t.Fatalf("expected id %q, got %q", item.ID, got.GetId())
	}
	if got.GetTitle() != item.Title {
		t.Fatalf("expected title %q, got %q", item.Title, got.GetTitle())
	}
	if got.GetVersion() != item.Version {
		t.Fatalf("expected version %d, got %d", item.Version, got.GetVersion())
	}
}

func TestModelItemsToProto(t *testing.T) {
	t.Parallel()

	items := []*model.VaultItem{
		{ID: uuid.MustParse("7d444840-9dc0-11d1-b245-5ffdce74fad2"), Type: model.ItemTypeText, Title: "a"},
		{ID: uuid.MustParse("8d444840-9dc0-11d1-b245-5ffdce74fad2"), Type: model.ItemTypeText, Title: "b"},
	}

	got := modelItemsToProto(items)
	if len(got) != 2 {
		t.Fatalf("expected 2 items, got %d", len(got))
	}
	if got[0].GetId() != items[0].ID.String() || got[1].GetId() != items[1].ID.String() {
		t.Fatal("unexpected items order/content")
	}
}
