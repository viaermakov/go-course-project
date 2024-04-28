package ordersrepo

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"route256.ozon.ru/project/loms/internals/infra/db"
	"route256.ozon.ru/project/loms/internals/infra/shardmanager"
	ordersrepo "route256.ozon.ru/project/loms/internals/repository/ordersrepo/sqlc"
	"route256.ozon.ru/project/loms/model/itemmodel"
	"route256.ozon.ru/project/loms/model/ordermodel"
)

type OrdersRepo struct {
}

var (
	ErrOrderNotFound = errors.New("order is not found")
)

func NewRepo() *OrdersRepo {
	return &OrdersRepo{}
}

func (repo *OrdersRepo) Create(ctx context.Context, tx db.Tx, shardIndex shardmanager.ShardIndex, userId int64, items []itemmodel.Item) (int64, error) {
	q := ordersrepo.New(tx)
	orderId, err := repo.generateId(ctx, q, shardIndex)

	if err != nil {
		return 0, handleSqlError(err)
	}

	err = q.InsertOrderInfo(ctx, ordersrepo.InsertOrderInfoParams{
		OrderID: orderId,
		UserID:  userId,
		Status:  int32(ordermodel.StatusNew),
	})

	if err != nil {
		return 0, handleSqlError(err)
	}

	for _, item := range items {
		err := q.InsertOrderItem(ctx, ordersrepo.InsertOrderItemParams{
			OrderID: orderId,
			ItemID:  int64(item.SkuId),
			Count:   int32(item.Count),
		})

		if err != nil {
			return 0, handleSqlError(err)
		}
	}

	return orderId, err
}

func (repo *OrdersRepo) generateId(ctx context.Context, q *ordersrepo.Queries, shardIndex shardmanager.ShardIndex) (int64, error) {
	initialId := int64(shardmanager.MaxShards)
	lastOrder, err := q.GetLastOrderItem(ctx)

	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return 0, handleSqlError(err)
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		initialId = lastOrder.OrderID - (lastOrder.OrderID % shardmanager.MaxShards)
	}

	return shardmanager.GenerateUniqId(initialId, shardIndex), nil
}

func (repo *OrdersRepo) SetStatus(ctx context.Context, tx db.Tx, orderId int64, status ordermodel.Status) error {
	q := ordersrepo.New(tx)

	return q.UpdateOrderStatus(ctx, ordersrepo.UpdateOrderStatusParams{
		OrderID: orderId,
		Status:  int32(status),
	})
}

func (repo *OrdersRepo) GetOrder(ctx context.Context, tx db.Tx, orderId int64) (int64, ordermodel.Status, []itemmodel.Item, error) {
	q := ordersrepo.New(tx)
	orderInfo, err := q.GetOrderInfo(ctx, orderId)

	if err != nil {
		return 0, 0, []itemmodel.Item{}, handleSqlError(err)
	}

	orderItems, err := q.GetOrderItem(ctx, orderId)

	if err != nil {
		return 0, 0, []itemmodel.Item{}, handleSqlError(err)
	}

	var items []itemmodel.Item

	for _, item := range orderItems {
		items = append(items, itemmodel.Item{
			SkuId: uint32(item.ItemID),
			Count: uint16(item.Count),
		})
	}

	return orderInfo.UserID, ordermodel.Status(orderInfo.Status), items, nil
}

func handleSqlError(err error) error {
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return ErrOrderNotFound
	default:
		return err
	}
}
