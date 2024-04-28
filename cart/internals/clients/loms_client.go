package clients

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"route256.ozon.ru/project/cart/models"
	desc "route256.ozon.ru/project/cart/pkg/api/order/v1"
	"time"
)

type LomsClient struct {
	client desc.OrderClient
}

func NewLomsClient(address string) (*LomsClient, error) {
	connection, err := grpc.Dial(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		return nil, err
	}

	return &LomsClient{
		client: desc.NewOrderClient(connection),
	}, nil
}

func (lomsClient *LomsClient) CreateOrder(ctx context.Context, userId int64, items []models.Product) (int64, error) {
	orderItems := make([]*desc.OrderItem, len(items))

	for i, item := range items {
		orderItems[i] = &desc.OrderItem{
			Sku:   uint32(item.SkuId),
			Count: uint32(item.Count),
		}
	}

	request := &desc.OrderCreateRequest{
		User:  userId,
		Items: orderItems,
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	response, err := lomsClient.client.OrderCreate(ctx, request)

	if err != nil {
		return 0, err
	}

	return response.OrderID, nil
}

func (lomsClient *LomsClient) GetStockInfo(ctx context.Context, skuId int64) (uint64, error) {
	request := &desc.StocksInfoRequest{Sku: uint32(skuId)}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	response, err := lomsClient.client.StocksInfo(ctx, request)

	if err != nil {
		return 0, err
	}

	return response.Count, err
}
