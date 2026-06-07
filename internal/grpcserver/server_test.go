package grpcserver

import (
	"context"
	"testing"
	"time"

	"gophkeeper/internal/model"
	"gophkeeper/internal/repository"
	"gophkeeper/internal/service"
	pb "gophkeeper/proto"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type registerServiceMock struct {
	registerFn func(ctx context.Context, login, password string) (string, error)
}

func (m *registerServiceMock) Register(ctx context.Context, login, password string) (string, error) {
	return m.registerFn(ctx, login, password)
}

type loginServiceMock struct {
	loginFn func(ctx context.Context, login, password string) (string, error)
}

func (m *loginServiceMock) Login(ctx context.Context, login, password string) (string, error) {
	return m.loginFn(ctx, login, password)
}

type vaultServiceMock struct {
	createFn func(ctx context.Context, item *model.VaultItem) (*model.VaultItem, error)
	getFn    func(ctx context.Context, userID string, itemID string) (*model.VaultItem, error)
	listFn   func(ctx context.Context, userID string, includeDeleted bool, limit, offset int) ([]*model.VaultItem, error)
	updateFn func(ctx context.Context, item *model.VaultItem) (*model.VaultItem, error)
	deleteFn func(ctx context.Context, userID string, itemID string) (time.Time, error)
	syncFn   func(ctx context.Context, userID string, since time.Time, includeDeleted bool) ([]*model.VaultItem, error)
}

func (m *vaultServiceMock) CreateItem(ctx context.Context, item *model.VaultItem) (*model.VaultItem, error) {
	return m.createFn(ctx, item)
}
func (m *vaultServiceMock) GetItem(ctx context.Context, userID string, itemID string) (*model.VaultItem, error) {
	return m.getFn(ctx, userID, itemID)
}
func (m *vaultServiceMock) ListItems(ctx context.Context, userID string, includeDeleted bool, limit, offset int) ([]*model.VaultItem, error) {
	return m.listFn(ctx, userID, includeDeleted, limit, offset)
}
func (m *vaultServiceMock) UpdateItem(ctx context.Context, item *model.VaultItem) (*model.VaultItem, error) {
	return m.updateFn(ctx, item)
}
func (m *vaultServiceMock) DeleteItem(ctx context.Context, userID string, itemID string) (time.Time, error) {
	return m.deleteFn(ctx, userID, itemID)
}
func (m *vaultServiceMock) SyncItems(ctx context.Context, userID string, since time.Time, includeDeleted bool) ([]*model.VaultItem, error) {
	return m.syncFn(ctx, userID, since, includeDeleted)
}

func TestServer_Register_Success(t *testing.T) {
	t.Parallel()

	registerSvc := &registerServiceMock{
		registerFn: func(ctx context.Context, login, password string) (string, error) {
			return "550e8400-e29b-41d4-a716-446655440000", nil
		},
	}
	loginSvc := &loginServiceMock{}
	vaultSvc := &vaultServiceMock{}

	srv := NewServer(registerSvc, loginSvc, vaultSvc)

	login := "u1"
	password := "p1"

	resp, err := srv.Register(context.Background(), pb.RegisterRequest_builder{
		Login:    &login,
		Password: &password,
	}.Build())
	if err != nil {
		t.Fatalf("Register returned error: %v", err)
	}
	if resp.GetUserId() != "550e8400-e29b-41d4-a716-446655440000" {
		t.Fatalf("unexpected user id: %q", resp.GetUserId())
	}
}

func TestServer_Register_AlreadyExists(t *testing.T) {
	t.Parallel()

	registerSvc := &registerServiceMock{
		registerFn: func(ctx context.Context, login, password string) (string, error) {
			return "", repository.ErrUserAlreadyExists
		},
	}
	srv := NewServer(registerSvc, &loginServiceMock{}, &vaultServiceMock{})

	login := "u1"
	password := "p1"

	_, err := srv.Register(context.Background(), pb.RegisterRequest_builder{
		Login:    &login,
		Password: &password,
	}.Build())
	if err == nil {
		t.Fatal("expected error")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("expected grpc status error")
	}
	if st.Code() != codes.AlreadyExists {
		t.Fatalf("expected AlreadyExists, got %v", st.Code())
	}
}

func TestServer_Login_InvalidCredentials(t *testing.T) {
	t.Parallel()

	loginSvc := &loginServiceMock{
		loginFn: func(ctx context.Context, login, password string) (string, error) {
			return "", service.ErrInvalidCredentials
		},
	}
	srv := NewServer(&registerServiceMock{}, loginSvc, &vaultServiceMock{})

	login := "u1"
	password := "bad"

	_, err := srv.Login(context.Background(), pb.LoginRequest_builder{
		Login:    &login,
		Password: &password,
	}.Build())
	if err == nil {
		t.Fatal("expected error")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("expected grpc status error")
	}
	if st.Code() != codes.Unauthenticated {
		t.Fatalf("expected Unauthenticated, got %v", st.Code())
	}
}

func TestServer_CreateItem_Unauthenticated(t *testing.T) {
	t.Parallel()

	srv := NewServer(&registerServiceMock{}, &loginServiceMock{}, &vaultServiceMock{})

	title := "note1"
	meta := "meta"
	payload := []byte("cipher")

	_, err := srv.CreateItem(context.Background(), pb.CreateItemRequest_builder{
		Type:             pb.ItemType_ITEM_TYPE_TEXT.Enum(),
		Title:            &title,
		Meta:             &meta,
		PayloadEncrypted: payload,
	}.Build())
	if err == nil {
		t.Fatal("expected error")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("expected grpc status error")
	}
	if st.Code() != codes.Unauthenticated {
		t.Fatalf("expected Unauthenticated, got %v", st.Code())
	}
}

func TestServer_CreateItem_Success(t *testing.T) {
	t.Parallel()

	vaultSvc := &vaultServiceMock{
		createFn: func(ctx context.Context, item *model.VaultItem) (*model.VaultItem, error) {
			item.ID = uuid.MustParse("7d444840-9dc0-11d1-b245-5ffdce74fad2")
			return item, nil
		},
	}
	srv := NewServer(&registerServiceMock{}, &loginServiceMock{}, vaultSvc)

	ctx := withUserID(context.Background(), "1")

	title := "note1"
	meta := "meta"
	payload := []byte("cipher")

	resp, err := srv.CreateItem(ctx, pb.CreateItemRequest_builder{
		Type:             pb.ItemType_ITEM_TYPE_TEXT.Enum(),
		Title:            &title,
		Meta:             &meta,
		PayloadEncrypted: payload,
	}.Build())
	if err != nil {
		t.Fatalf("CreateItem returned error: %v", err)
	}
	if resp.GetItem() == nil {
		t.Fatal("expected item in response")
	}
	if resp.GetItem().GetId() != "7d444840-9dc0-11d1-b245-5ffdce74fad2" {
		t.Fatalf("unexpected item id: %q", resp.GetItem().GetId())
	}
}
