package lomsservice

import (
	"context"
	"errors"
	"fmt"
	"route256.ozon.ru/project/loms/internals/infra/db"
	"route256.ozon.ru/project/loms/internals/infra/shardmanager"
	"route256.ozon.ru/project/loms/model/itemmodel"
	"route256.ozon.ru/project/loms/model/ordermodel"
	"strconv"
)

type StocksProvider interface {
	Reserve(ctx context.Context, trx db.Tx, items []itemmodel.Item) error
	Remove(ctx context.Context, trx db.Tx, items []itemmodel.Item) error
	Cancel(ctx context.Context, trx db.Tx, items []itemmodel.Item) error
	GetById(ctx context.Context, trx db.Tx, skuId int64) (uint64, error)
}

type ShardManagerProvider interface {
	Get(key shardmanager.ShardKey) (db.Pool, shardmanager.ShardIndex, error)
	GetByOrderId(id int64) (db.Pool, error)
}

type OrdersProvider interface {
	Create(ctx context.Context, trx db.Tx, shardIndex shardmanager.ShardIndex, userId int64, items []itemmodel.Item) (int64, error)
	GetOrder(ctx context.Context, trx db.Tx, orderId int64) (int64, ordermodel.Status, []itemmodel.Item, error)
	SetStatus(ctx context.Context, trx db.Tx, orderId int64, status ordermodel.Status) error
}

type NotifierProvider interface {
	Publish(ctx context.Context, tx db.Tx, orderId int64, orderStatus ordermodel.Status) error
}

type LomsService struct {
	shardManager ShardManagerProvider
	stocksPool   db.Pool
	stocks       StocksProvider
	orders       OrdersProvider
	notifier     NotifierProvider
}

var (
	ErrIncorrectStatus = errors.New("couldn't process the order due to the incorrect status")
	ErrGetOrders       = errors.New("couldn't get all orders")
)

func NewLomsService(
	shardManager ShardManagerProvider,
	stocksPool db.Pool,
	stocks StocksProvider,
	orders OrdersProvider,
	notifier NotifierProvider,
) *LomsService {
	return &LomsService{
		shardManager: shardManager,
		stocksPool:   stocksPool,
		notifier:     notifier,
		stocks:       stocks,
		orders:       orders,
	}
}

func (service LomsService) CreateOrder(ctx context.Context, userId int64, items []itemmodel.Item) (int64, error) {
	var orderId int64

	shard, shardIndex, err := service.shardManager.Get(
		shardmanager.ShardKey(strconv.FormatInt(userId, 10)),
	)

	if err != nil {
		return orderId, err
	}

	err = db.WithTransactions(ctx, []db.Pool{shard, service.stocksPool}, db.WriteOrRead, func(ctx context.Context, tx []db.Tx) error {
		shardTx := tx[0]
		stocksTx := tx[1]

		id, err := service.orders.Create(ctx, shardTx, shardIndex, userId, items)
		orderId = id

		if err != nil {
			return err
		}

		err = service.notifier.Publish(ctx, shardTx, orderId, ordermodel.StatusNew)

		if err != nil {
			return err
		}

		err = service.stocks.Reserve(ctx, stocksTx, items)

		if err != nil {
			statusErr := service.orders.SetStatus(ctx, shardTx, orderId, ordermodel.StatusFailed)

			if statusErr != nil {
				return statusErr
			}

			notifierErr := service.notifier.Publish(ctx, shardTx, orderId, ordermodel.StatusFailed)

			if notifierErr != nil {
				return notifierErr
			}

			return err
		}

		err = service.notifier.Publish(ctx, shardTx, orderId, ordermodel.StatusAwaiting)

		if err != nil {
			return err
		}

		statusErr := service.orders.SetStatus(ctx, shardTx, orderId, ordermodel.StatusAwaiting)

		if statusErr != nil {
			return statusErr
		}

		return nil
	})

	return orderId, err
}

func (service LomsService) GetOrder(ctx context.Context, orderId int64) (*ordermodel.Info, error) {
	shard, err := service.shardManager.GetByOrderId(orderId)

	if err != nil {
		return nil, err
	}

	var orderInfo *ordermodel.Info

	err = db.WithTransaction(ctx, shard, db.ReadOnly, func(ctx context.Context, tx db.Tx) error {
		info, err := service.getOrder(ctx, tx, orderId)
		orderInfo = info
		return err
	})

	return orderInfo, err
}

func (service LomsService) PayOrder(ctx context.Context, orderId int64) error {
	shard, err := service.shardManager.GetByOrderId(orderId)

	if err != nil {
		return err
	}

	return db.WithTransactions(ctx, []db.Pool{shard, service.stocksPool}, db.WriteOrRead, func(ctx context.Context, tx []db.Tx) error {
		shardTx := tx[0]
		stocksTx := tx[1]
		order, err := service.getOrder(ctx, shardTx, orderId)

		if err != nil {
			return err
		}

		if order.Status != ordermodel.StatusAwaiting {
			return fmt.Errorf("%w - order is not in status awaiting", ErrIncorrectStatus)
		}

		err = service.stocks.Remove(ctx, stocksTx, order.Items)

		if err != nil {
			return err
		}

		err = service.orders.SetStatus(ctx, shardTx, orderId, ordermodel.StatusPaid)

		if err != nil {
			return err
		}

		err = service.notifier.Publish(ctx, shardTx, orderId, ordermodel.StatusPaid)

		if err != nil {
			return err
		}

		return nil
	})
}

func (service LomsService) CancelOrder(ctx context.Context, orderId int64) error {
	shard, err := service.shardManager.GetByOrderId(orderId)

	if err != nil {
		return err
	}

	return db.WithTransactions(ctx, []db.Pool{shard, service.stocksPool}, db.WriteOrRead, func(ctx context.Context, tx []db.Tx) error {
		shardTx := tx[0]
		stocksTx := tx[1]
		order, err := service.getOrder(ctx, shardTx, orderId)

		if err != nil {
			return err
		}

		if order.Status != ordermodel.StatusAwaiting {
			return fmt.Errorf("%w - order is not in status awaiting", ErrIncorrectStatus)
		}

		err = service.stocks.Cancel(ctx, stocksTx, order.Items)

		if err != nil {
			return err
		}

		err = service.orders.SetStatus(ctx, shardTx, orderId, ordermodel.StatusCanceled)

		if err != nil {
			return err
		}

		err = service.notifier.Publish(ctx, shardTx, orderId, ordermodel.StatusCanceled)

		if err != nil {
			return err
		}

		return nil
	})
}

func (service LomsService) GetOrders(ctx context.Context, orderIds []int64) ([]*ordermodel.Info, error) {
	orders := make([]*ordermodel.Info, len(orderIds))

	for i, orderId := range orderIds {
		shard, err := service.shardManager.GetByOrderId(orderId)

		if err != nil {
			return nil, ErrGetOrders
		}

		err = db.WithTransaction(ctx, shard, db.ReadOnly, func(ctx context.Context, tx db.Tx) error {
			info, err := service.getOrder(ctx, tx, orderId)

			if err != nil {
				return err
			}

			orders[i] = info
			return nil
		})

		if err != nil {
			return nil, ErrGetOrders
		}
	}

	return orders, nil
}

func (service LomsService) GetAvailableStocks(ctx context.Context, skuId int64) (uint64, error) {
	var availableStocks uint64

	err := db.WithTransaction(ctx, service.stocksPool, db.ReadOnly, func(ctx context.Context, tx db.Tx) error {
		stocks, err := service.stocks.GetById(ctx, tx, skuId)
		availableStocks = stocks
		return err
	})

	return availableStocks, err
}

func (service LomsService) getOrder(ctx context.Context, tx db.Tx, orderId int64) (*ordermodel.Info, error) {
	userId, status, items, err := service.orders.GetOrder(ctx, tx, orderId)

	if err != nil {
		return nil, err
	}

	return &ordermodel.Info{
		OrderId: orderId,
		Status:  status,
		User:    userId,
		Items:   items,
	}, nil
}
