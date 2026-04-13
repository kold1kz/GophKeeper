// Package grpcserver содержит реализацию gRPC-сервера GophKeeper.
package grpcserver

import (
	"fmt"
	"time"

	"gophkeeper/internal/model"
	pb "gophkeeper/proto"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// protoItemTypeToModel преобразует тип записи из protobuf-представления
// в доменный тип model.ItemType.
//
// Возвращает ошибку, если тип не поддерживается.
func protoItemTypeToModel(t pb.ItemType) (model.ItemType, error) {
	switch t {
	case pb.ItemType_ITEM_TYPE_LOGIN_PASSWORD:
		return model.ItemTypeLoginPassword, nil
	case pb.ItemType_ITEM_TYPE_TEXT:
		return model.ItemTypeText, nil
	case pb.ItemType_ITEM_TYPE_BINARY:
		return model.ItemTypeBinary, nil
	case pb.ItemType_ITEM_TYPE_BANK_CARD:
		return model.ItemTypeBankCard, nil
	default:
		return "", fmt.Errorf("unsupported item type: %v", t)
	}
}

// modelItemTypeToProto преобразует доменный тип записи
// в protobuf-представление.
func modelItemTypeToProto(t model.ItemType) pb.ItemType {
	switch t {
	case model.ItemTypeLoginPassword:
		return pb.ItemType_ITEM_TYPE_LOGIN_PASSWORD
	case model.ItemTypeText:
		return pb.ItemType_ITEM_TYPE_TEXT
	case model.ItemTypeBinary:
		return pb.ItemType_ITEM_TYPE_BINARY
	case model.ItemTypeBankCard:
		return pb.ItemType_ITEM_TYPE_BANK_CARD
	default:
		return pb.ItemType_ITEM_TYPE_UNSPECIFIED
	}
}

// createItemRequestToModel преобразует protobuf-запрос создания записи
// в доменную модель VaultItem.
//
// В запись подставляется идентификатор текущего пользователя.
func createItemRequestToModel(userID int64, req *pb.CreateItemRequest) (*model.VaultItem, error) {
	itemType, err := protoItemTypeToModel(req.GetType())
	if err != nil {
		return nil, err
	}

	var clientUpdatedAt *time.Time
	if ts := req.GetClientUpdatedAt(); ts != nil {
		t := ts.AsTime()
		clientUpdatedAt = &t
	}

	return &model.VaultItem{
		UserID:           userID,
		Type:             itemType,
		Title:            req.GetTitle(),
		Meta:             req.GetMeta(),
		PayloadEncrypted: req.GetPayloadEncrypted(),
		PayloadNonce:     req.GetPayloadNonce(),
		PayloadHash:      req.GetPayloadHash(),
		ClientUpdatedAt:  clientUpdatedAt,
	}, nil
}

// updateItemRequestToModel преобразует protobuf-запрос обновления записи
// в доменную модель VaultItem.
//
// В запись подставляется идентификатор текущего пользователя.
func updateItemRequestToModel(userID int64, req *pb.UpdateItemRequest) *model.VaultItem {
	var clientUpdatedAt *time.Time
	if ts := req.GetClientUpdatedAt(); ts != nil {
		t := ts.AsTime()
		clientUpdatedAt = &t
	}

	return &model.VaultItem{
		ID:               req.GetId(),
		UserID:           userID,
		Title:            req.GetTitle(),
		Meta:             req.GetMeta(),
		PayloadEncrypted: req.GetPayloadEncrypted(),
		PayloadNonce:     req.GetPayloadNonce(),
		PayloadHash:      req.GetPayloadHash(),
		ClientUpdatedAt:  clientUpdatedAt,
		Version:          req.GetVersion(),
	}
}

// modelItemToProto преобразует доменную запись VaultItem
// в protobuf-представление VaultItem.
func modelItemToProto(item *model.VaultItem) *pb.VaultItem {
	if item == nil {
		return nil
	}

	var clientUpdatedAt *timestamppb.Timestamp
	if item.ClientUpdatedAt != nil {
		clientUpdatedAt = timestamppb.New(*item.ClientUpdatedAt)
	}

	var deletedAt *timestamppb.Timestamp
	if item.DeletedAt != nil {
		deletedAt = timestamppb.New(*item.DeletedAt)
	}

	return pb.VaultItem_builder{
		Id:               &item.ID,
		Type:             modelItemTypeToProto(item.Type).Enum(),
		Title:            &item.Title,
		Meta:             &item.Meta,
		PayloadEncrypted: item.PayloadEncrypted,
		PayloadNonce:     item.PayloadNonce,
		PayloadHash:      &item.PayloadHash,
		Version:          &item.Version,
		ClientUpdatedAt:  clientUpdatedAt,
		CreatedAt:        timestamppb.New(item.CreatedAt),
		UpdatedAt:        timestamppb.New(item.UpdatedAt),
		DeletedAt:        deletedAt,
	}.Build()
}

// modelItemsToProto преобразует срез доменных записей
// в срез protobuf-объектов.
func modelItemsToProto(items []*model.VaultItem) []*pb.VaultItem {
	result := make([]*pb.VaultItem, 0, len(items))
	for _, item := range items {
		result = append(result, modelItemToProto(item))
	}
	return result
}
