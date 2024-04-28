package lomsservice

import (
	"context"
	"errors"
	"github.com/gojuno/minimock/v3"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/stretchr/testify/require"
	"log"
	"route256.ozon.ru/project/loms/internals/infra/db"
	"route256.ozon.ru/project/loms/internals/infra/shardmanager"
	"route256.ozon.ru/project/loms/model/itemmodel"
	"route256.ozon.ru/project/loms/model/ordermodel"
	"strconv"
	"testing"
)

type inputData struct {
	userId  int64
	orderId int64
	skuId   int64
	items   []itemmodel.Item
}

func getTxMock(ctx context.Context, client db.Pool, p pgxmock.PgxPoolIface) db.Tx {
	p.ExpectBegin()
	tx, _ := client.Get(db.WriteOrRead).Begin(ctx)
	return tx
}

func TestLomsService_GetOrder(t *testing.T) {
	tests := []struct {
		name       string
		inputData  inputData
		mock       func(ctx context.Context, shardManager *ShardManagerProviderMock, client db.Pool, pp pgxmock.PgxPoolIface, l *StocksProviderMock, p *OrdersProviderMock, n *NotifierProviderMock, i inputData, wantResult *ordermodel.Info, wantErr error)
		wantResult *ordermodel.Info
		wantErr    error
	}{
		{
			name: "should be error if failed to find order",
			inputData: inputData{
				orderId: 1,
				userId:  2,
			},
			mock: func(ctx context.Context, sh *ShardManagerProviderMock, pool db.Pool, conn pgxmock.PgxPoolIface, s *StocksProviderMock, o *OrdersProviderMock, n *NotifierProviderMock, i inputData, wantResult *ordermodel.Info, wantErr error) {
				tx := getTxMock(ctx, pool, conn)
				conn.ExpectBegin()
				conn.ExpectCommit()
				sh.GetByOrderIdMock.Expect(i.orderId).Return(pool, nil)
				o.GetOrderMock.Expect(ctx, tx, i.orderId).Return(0, 0, []itemmodel.Item{}, wantErr)
			},
			wantErr: errors.New("failed to find an order"),
		},
		{
			name: "should be successful",
			inputData: inputData{
				orderId: 0,
				userId:  2,
				items: []itemmodel.Item{
					{SkuId: 1, Count: 1},
				},
			},
			mock: func(ctx context.Context, sh *ShardManagerProviderMock, pool db.Pool, conn pgxmock.PgxPoolIface, s *StocksProviderMock, o *OrdersProviderMock, n *NotifierProviderMock, i inputData, wantResult *ordermodel.Info, wantErr error) {
				tx := getTxMock(ctx, pool, conn)
				conn.ExpectBegin()
				conn.ExpectCommit()
				sh.GetByOrderIdMock.Expect(i.orderId).Return(pool, nil)
				o.GetOrderMock.Expect(ctx, tx, i.orderId).Return(i.userId, wantResult.Status, wantResult.Items, nil)
			},
			wantResult: &ordermodel.Info{
				Status: ordermodel.StatusPaid,
				User:   2,
				Items: []itemmodel.Item{
					{SkuId: 1, Count: 1},
				},
			},
		},
	}

	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			mc := minimock.NewController(t)
			conn, err := pgxmock.NewPool()

			if err != nil {
				log.Fatalln(err.Error())
			}

			pool, err := db.NewDbClientFromConnection([]db.Tx{conn}, []db.Tx{})

			if err != nil {
				log.Fatalln(err.Error())
			}

			ctx := context.Background()
			shardManager := NewShardManagerProviderMock(mc)
			stocksProviderMock := NewStocksProviderMock(mc)
			ordersProviderMock := NewOrdersProviderMock(mc)
			notifierProviderMock := NewNotifierProviderMock(mc)
			lomsService := NewLomsService(
				shardManager,
				pool,
				stocksProviderMock,
				ordersProviderMock,
				notifierProviderMock,
			)

			test.mock(ctx, shardManager, pool, conn, stocksProviderMock, ordersProviderMock, notifierProviderMock, test.inputData, test.wantResult, test.wantErr)
			gotResult, gotErr := lomsService.GetOrder(ctx, test.inputData.orderId)

			require.ErrorIs(t, test.wantErr, gotErr)
			require.Equal(t, test.wantResult, gotResult)
		})
	}
}

func TestLomsService_CreateOrder(t *testing.T) {
	tests := []struct {
		name       string
		inputData  inputData
		mock       func(ctx context.Context, sh *ShardManagerProviderMock, client db.Pool, pp pgxmock.PgxPoolIface, l *StocksProviderMock, p *OrdersProviderMock, n *NotifierProviderMock, i inputData, wantResult int64, wantErr error)
		wantResult int64
		wantErr    error
	}{
		{
			name: "should be error if failed to create order",
			inputData: inputData{
				userId: 1,
				items: []itemmodel.Item{
					{SkuId: 1, Count: 1},
				},
			},
			mock: func(ctx context.Context, sh *ShardManagerProviderMock, pool db.Pool, conn pgxmock.PgxPoolIface, s *StocksProviderMock, o *OrdersProviderMock, n *NotifierProviderMock, i inputData, wantResult int64, wantErr error) {
				tx := getTxMock(ctx, pool, conn)
				conn.ExpectBegin()
				conn.ExpectBegin()
				conn.ExpectCommit()
				conn.ExpectCommit()
				sh.GetMock.Expect(shardmanager.ShardKey(strconv.FormatInt(i.userId, 10))).Return(pool, 0, nil)
				o.CreateMock.Expect(ctx, tx, 0, i.userId, i.items).Return(0, wantErr)
			},
			wantErr: errors.New("failed to create an order"),
		},
		{
			name: "should be error if failed update status to failed",
			inputData: inputData{
				userId: 1,
				items: []itemmodel.Item{
					{SkuId: 1, Count: 1},
				},
			},
			mock: func(ctx context.Context, sh *ShardManagerProviderMock, pool db.Pool, conn pgxmock.PgxPoolIface, s *StocksProviderMock, o *OrdersProviderMock, n *NotifierProviderMock, i inputData, wantResult int64, wantErr error) {
				tx := getTxMock(ctx, pool, conn)
				conn.ExpectBegin()
				conn.ExpectBegin()
				conn.ExpectCommit()
				conn.ExpectCommit()
				n.PublishMock.When(ctx, tx, i.orderId, ordermodel.StatusNew).Then(nil)
				sh.GetMock.Expect(shardmanager.ShardKey(strconv.FormatInt(i.userId, 10))).Return(pool, 0, nil)
				o.CreateMock.Expect(ctx, tx, 0, i.userId, i.items).Return(wantResult, nil)
				s.ReserveMock.Expect(ctx, tx, i.items).Return(errors.New("failed to reserve"))
				o.SetStatusMock.Expect(ctx, tx, wantResult, ordermodel.StatusFailed).Return(wantErr)
			},
			wantErr: errors.New("failed to update status to failed"),
		},
		{
			name: "should be error if failed to reserve product",
			inputData: inputData{
				userId: 1,
				items: []itemmodel.Item{
					{SkuId: 1, Count: 1},
				},
			},
			mock: func(ctx context.Context, sh *ShardManagerProviderMock, pool db.Pool, conn pgxmock.PgxPoolIface, s *StocksProviderMock, o *OrdersProviderMock, n *NotifierProviderMock, i inputData, wantResult int64, wantErr error) {
				tx := getTxMock(ctx, pool, conn)
				conn.ExpectBegin()
				conn.ExpectBegin()
				conn.ExpectCommit()
				conn.ExpectCommit()
				n.PublishMock.When(minimock.AnyContext, tx, i.orderId, ordermodel.StatusNew).Then(nil)
				sh.GetMock.Expect(shardmanager.ShardKey(strconv.FormatInt(i.userId, 10))).Return(pool, 0, nil)
				o.CreateMock.Expect(minimock.AnyContext, tx, 0, i.userId, i.items).Return(wantResult, nil)
				s.ReserveMock.Expect(minimock.AnyContext, tx, i.items).Return(wantErr)
				n.PublishMock.When(minimock.AnyContext, tx, i.orderId, ordermodel.StatusFailed).Then(nil)
				o.SetStatusMock.Expect(minimock.AnyContext, tx, wantResult, ordermodel.StatusFailed).Return(nil)
			},
			wantErr: errors.New("failed to reserve"),
		},
		{
			name: "should be error if failed update status to awaiting",
			inputData: inputData{
				userId: 1,
				items: []itemmodel.Item{
					{SkuId: 1, Count: 1},
				},
			},
			mock: func(ctx context.Context, sh *ShardManagerProviderMock, pool db.Pool, conn pgxmock.PgxPoolIface, s *StocksProviderMock, o *OrdersProviderMock, n *NotifierProviderMock, i inputData, wantResult int64, wantErr error) {
				tx := getTxMock(ctx, pool, conn)
				conn.ExpectBegin()
				conn.ExpectBegin()
				conn.ExpectCommit()
				conn.ExpectCommit()
				n.PublishMock.When(minimock.AnyContext, tx, i.orderId, ordermodel.StatusNew).Then(nil)
				sh.GetMock.Expect(shardmanager.ShardKey(strconv.FormatInt(i.userId, 10))).Return(pool, 0, nil)
				o.CreateMock.Expect(minimock.AnyContext, tx, 0, i.userId, i.items).Return(wantResult, nil)
				s.ReserveMock.Expect(minimock.AnyContext, tx, i.items).Return(nil)
				n.PublishMock.When(minimock.AnyContext, tx, i.orderId, ordermodel.StatusAwaiting).Then(nil)
				o.SetStatusMock.Expect(minimock.AnyContext, tx, wantResult, ordermodel.StatusAwaiting).Return(wantErr)
			},
			wantErr: errors.New("failed to update status to awaiting"),
		},
		{
			name: "should be successful",
			inputData: inputData{
				userId: 1,
				items: []itemmodel.Item{
					{SkuId: 1, Count: 1},
				},
			},
			mock: func(ctx context.Context, sh *ShardManagerProviderMock, pool db.Pool, conn pgxmock.PgxPoolIface, s *StocksProviderMock, o *OrdersProviderMock, n *NotifierProviderMock, i inputData, wantResult int64, wantErr error) {
				tx := getTxMock(ctx, pool, conn)
				conn.ExpectBegin()
				conn.ExpectBegin()
				conn.ExpectCommit()
				conn.ExpectCommit()
				o.CreateMock.When(minimock.AnyContext, tx, 0, i.userId, i.items).Then(wantResult, nil)
				n.PublishMock.When(minimock.AnyContext, tx, wantResult, ordermodel.StatusNew).Then(nil)
				sh.GetMock.Expect(shardmanager.ShardKey(strconv.FormatInt(i.userId, 10))).Return(pool, 0, nil)
				s.ReserveMock.When(minimock.AnyContext, tx, i.items).Then(nil)
				n.PublishMock.When(minimock.AnyContext, tx, wantResult, ordermodel.StatusAwaiting).Then(nil)
				o.SetStatusMock.When(minimock.AnyContext, tx, wantResult, ordermodel.StatusAwaiting).Then(nil)
			},
			wantErr:    nil,
			wantResult: 999,
		},
	}

	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			mc := minimock.NewController(t)
			conn, err := pgxmock.NewPool()

			if err != nil {
				log.Fatalln(err.Error())
			}

			pool, err := db.NewDbClientFromConnection([]db.Tx{conn}, []db.Tx{})

			if err != nil {
				log.Fatalln(err.Error())
			}

			ctx := context.Background()
			shardManager := NewShardManagerProviderMock(mc)
			stocksProviderMock := NewStocksProviderMock(mc)
			ordersProviderMock := NewOrdersProviderMock(mc)
			notifierProviderMock := NewNotifierProviderMock(mc)
			lomsService := NewLomsService(
				shardManager,
				pool,
				stocksProviderMock,
				ordersProviderMock,
				notifierProviderMock,
			)

			test.mock(ctx, shardManager, pool, conn, stocksProviderMock, ordersProviderMock, notifierProviderMock, test.inputData, test.wantResult, test.wantErr)
			gotResult, gotErr := lomsService.CreateOrder(ctx, test.inputData.userId, test.inputData.items)

			require.ErrorIs(t, test.wantErr, gotErr)
			require.Equal(t, test.wantResult, gotResult)
		})
	}
}

func TestLomsService_CancelOrder(t *testing.T) {
	tests := []struct {
		name       string
		inputData  inputData
		mock       func(ctx context.Context, sh *ShardManagerProviderMock, client db.Pool, pp pgxmock.PgxPoolIface, l *StocksProviderMock, p *OrdersProviderMock, n *NotifierProviderMock, i inputData, wantResult int64, wantErr error)
		wantResult int64
		wantErr    error
	}{
		{
			name: "should be error if failed to get order",
			inputData: inputData{
				orderId: 1,
				userId:  2,
				items: []itemmodel.Item{
					{SkuId: 1, Count: 1},
				},
			},
			mock: func(ctx context.Context, sh *ShardManagerProviderMock, pool db.Pool, conn pgxmock.PgxPoolIface, s *StocksProviderMock, o *OrdersProviderMock, n *NotifierProviderMock, i inputData, wantResult int64, wantErr error) {
				tx := getTxMock(ctx, pool, conn)
				sh.GetByOrderIdMock.Expect(i.orderId).Return(pool, nil)
				conn.ExpectBegin()
				conn.ExpectBegin()
				conn.ExpectCommit()
				conn.ExpectCommit()
				o.GetOrderMock.Expect(ctx, tx, i.orderId).Return(0, 0, []itemmodel.Item{}, wantErr)

			},
			wantErr: errors.New("failed to get an order"),
		},
		{
			name: "should be error if order is not in awaiting status",
			inputData: inputData{
				orderId: 1,
				userId:  2,
				items: []itemmodel.Item{
					{SkuId: 1, Count: 1},
				},
			},
			mock: func(ctx context.Context, sh *ShardManagerProviderMock, pool db.Pool, conn pgxmock.PgxPoolIface, s *StocksProviderMock, o *OrdersProviderMock, n *NotifierProviderMock, i inputData, wantResult int64, wantErr error) {
				tx := getTxMock(ctx, pool, conn)
				sh.GetByOrderIdMock.Expect(i.orderId).Return(pool, nil)
				conn.ExpectBegin()
				conn.ExpectBegin()
				conn.ExpectCommit()
				conn.ExpectCommit()
				o.GetOrderMock.When(ctx, tx, i.orderId).Then(i.userId, ordermodel.StatusPaid, i.items, nil)
			},
			wantErr: ErrIncorrectStatus,
		},
		{
			name: "should be error if failed to cancel stocks",
			inputData: inputData{
				orderId: 1,
				userId:  2,
				items: []itemmodel.Item{
					{SkuId: 1, Count: 1},
				},
			},
			mock: func(ctx context.Context, sh *ShardManagerProviderMock, pool db.Pool, conn pgxmock.PgxPoolIface, s *StocksProviderMock, o *OrdersProviderMock, n *NotifierProviderMock, i inputData, wantResult int64, wantErr error) {
				tx := getTxMock(ctx, pool, conn)
				sh.GetByOrderIdMock.Expect(i.orderId).Return(pool, nil)
				conn.ExpectBegin()
				conn.ExpectBegin()
				conn.ExpectCommit()
				conn.ExpectCommit()
				o.GetOrderMock.Expect(ctx, tx, i.orderId).Return(i.userId, ordermodel.StatusAwaiting, i.items, nil)
				s.CancelMock.Expect(ctx, tx, i.items).Return(wantErr)
			},
			wantErr: errors.New("failed to cancel stocks"),
		},
		{
			name: "should be error if failed to update status",
			inputData: inputData{
				orderId: 1,
				userId:  2,
				items: []itemmodel.Item{
					{SkuId: 1, Count: 1},
				},
			},
			mock: func(ctx context.Context, sh *ShardManagerProviderMock, pool db.Pool, conn pgxmock.PgxPoolIface, s *StocksProviderMock, o *OrdersProviderMock, n *NotifierProviderMock, i inputData, wantResult int64, wantErr error) {
				tx := getTxMock(ctx, pool, conn)
				sh.GetByOrderIdMock.Expect(i.orderId).Return(pool, nil)
				conn.ExpectBegin()
				conn.ExpectBegin()
				conn.ExpectCommit()
				conn.ExpectCommit()
				o.GetOrderMock.Expect(ctx, tx, i.orderId).Return(i.userId, ordermodel.StatusAwaiting, i.items, nil)
				s.CancelMock.Expect(ctx, tx, i.items).Return(nil)
				o.SetStatusMock.Expect(ctx, tx, i.orderId, ordermodel.StatusCanceled).Return(wantErr)
			},
			wantErr: errors.New("failed to cancel stocks"),
		},
		{
			name: "should be successful",
			inputData: inputData{
				orderId: 1,
				userId:  2,
				items: []itemmodel.Item{
					{SkuId: 1, Count: 1},
				},
			},
			mock: func(ctx context.Context, sh *ShardManagerProviderMock, pool db.Pool, conn pgxmock.PgxPoolIface, s *StocksProviderMock, o *OrdersProviderMock, n *NotifierProviderMock, i inputData, wantResult int64, wantErr error) {
				tx := getTxMock(ctx, pool, conn)
				sh.GetByOrderIdMock.Expect(i.orderId).Return(pool, nil)
				conn.ExpectBegin()
				conn.ExpectBegin()
				conn.ExpectCommit()
				conn.ExpectCommit()
				o.GetOrderMock.When(ctx, tx, i.orderId).Then(i.userId, ordermodel.StatusAwaiting, i.items, nil)
				n.PublishMock.When(ctx, tx, i.orderId, ordermodel.StatusCanceled).Then(nil)
				s.CancelMock.When(ctx, tx, i.items).Then(nil)
				o.SetStatusMock.When(ctx, tx, i.orderId, ordermodel.StatusCanceled).Then(nil)
			},
		},
	}

	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			mc := minimock.NewController(t)
			conn, err := pgxmock.NewPool()

			if err != nil {
				log.Fatalln(err.Error())
			}

			pool, err := db.NewDbClientFromConnection([]db.Tx{conn}, []db.Tx{})

			if err != nil {
				log.Fatalln(err.Error())
			}

			ctx := context.Background()
			shardManager := NewShardManagerProviderMock(mc)
			stocksProviderMock := NewStocksProviderMock(mc)
			ordersProviderMock := NewOrdersProviderMock(mc)
			notifierProviderMock := NewNotifierProviderMock(mc)
			lomsService := NewLomsService(
				shardManager,
				pool,
				stocksProviderMock,
				ordersProviderMock,
				notifierProviderMock,
			)

			test.mock(ctx, shardManager, pool, conn, stocksProviderMock, ordersProviderMock, notifierProviderMock, test.inputData, test.wantResult, test.wantErr)
			gotErr := lomsService.CancelOrder(ctx, test.inputData.orderId)

			if test.wantErr == nil {
				require.NoError(t, gotErr)
			} else {
				require.ErrorAs(t, gotErr, &test.wantErr)
			}
		})
	}
}

func TestLomsService_PayOrder(t *testing.T) {
	tests := []struct {
		name       string
		inputData  inputData
		mock       func(ctx context.Context, sh *ShardManagerProviderMock, client db.Pool, pp pgxmock.PgxPoolIface, l *StocksProviderMock, p *OrdersProviderMock, n *NotifierProviderMock, i inputData, wantResult int64, wantErr error)
		wantResult int64
		wantErr    error
	}{
		{
			name: "should be error if failed to get order",
			inputData: inputData{
				orderId: 1,
				userId:  2,
				items: []itemmodel.Item{
					{SkuId: 1, Count: 1},
				},
			},
			mock: func(ctx context.Context, sh *ShardManagerProviderMock, pool db.Pool, conn pgxmock.PgxPoolIface, s *StocksProviderMock, o *OrdersProviderMock, n *NotifierProviderMock, i inputData, wantResult int64, wantErr error) {
				tx := getTxMock(ctx, pool, conn)
				sh.GetByOrderIdMock.Expect(i.orderId).Return(pool, nil)
				conn.ExpectBegin()
				conn.ExpectBegin()
				conn.ExpectCommit()
				conn.ExpectCommit()
				o.GetOrderMock.Expect(ctx, tx, i.orderId).Return(0, 0, []itemmodel.Item{}, wantErr)

			},
			wantErr: errors.New("failed to get an order"),
		},
		{
			name: "should be error if failed to remove stocks",
			inputData: inputData{
				orderId: 1,
				userId:  2,
				items: []itemmodel.Item{
					{SkuId: 1, Count: 1},
				},
			},
			mock: func(ctx context.Context, sh *ShardManagerProviderMock, pool db.Pool, conn pgxmock.PgxPoolIface, s *StocksProviderMock, o *OrdersProviderMock, n *NotifierProviderMock, i inputData, wantResult int64, wantErr error) {
				tx := getTxMock(ctx, pool, conn)
				sh.GetByOrderIdMock.Expect(i.orderId).Return(pool, nil)
				conn.ExpectBegin()
				conn.ExpectBegin()
				conn.ExpectCommit()
				conn.ExpectCommit()
				o.GetOrderMock.Expect(ctx, tx, i.orderId).Return(i.userId, ordermodel.StatusAwaiting, i.items, nil)
				s.RemoveMock.Expect(ctx, tx, i.items).Return(wantErr)
			},
			wantErr: errors.New("failed to remove stocks"),
		},
		{
			name: "should be error if order is not in awaiting status",
			inputData: inputData{
				orderId: 1,
				userId:  2,
				items: []itemmodel.Item{
					{SkuId: 1, Count: 1},
				},
			},
			mock: func(ctx context.Context, sh *ShardManagerProviderMock, pool db.Pool, conn pgxmock.PgxPoolIface, s *StocksProviderMock, o *OrdersProviderMock, n *NotifierProviderMock, i inputData, wantResult int64, wantErr error) {
				tx := getTxMock(ctx, pool, conn)
				sh.GetByOrderIdMock.Expect(i.orderId).Return(pool, nil)
				conn.ExpectBegin()
				conn.ExpectBegin()
				conn.ExpectCommit()
				conn.ExpectCommit()
				o.GetOrderMock.Expect(ctx, tx, i.orderId).Return(i.userId, ordermodel.StatusNew, i.items, nil)
			},
			wantErr: ErrIncorrectStatus,
		},
		{
			name: "should be error if failed to update order status",
			inputData: inputData{
				orderId: 1,
				userId:  2,
				items: []itemmodel.Item{
					{SkuId: 1, Count: 1},
				},
			},
			mock: func(ctx context.Context, sh *ShardManagerProviderMock, pool db.Pool, conn pgxmock.PgxPoolIface, s *StocksProviderMock, o *OrdersProviderMock, n *NotifierProviderMock, i inputData, wantResult int64, wantErr error) {
				tx := getTxMock(ctx, pool, conn)
				sh.GetByOrderIdMock.Expect(i.orderId).Return(pool, nil)
				conn.ExpectBegin()
				conn.ExpectBegin()
				conn.ExpectCommit()
				conn.ExpectCommit()
				o.GetOrderMock.Expect(ctx, tx, i.orderId).Return(i.userId, ordermodel.StatusAwaiting, i.items, nil)
				s.RemoveMock.Expect(ctx, tx, i.items).Return(nil)
				o.SetStatusMock.Expect(ctx, tx, i.orderId, ordermodel.StatusPaid).Return(wantErr)
			},
			wantErr: errors.New("failed to update order status"),
		},
		{
			name: "should be successful",
			inputData: inputData{
				orderId: 1,
				userId:  2,
				items: []itemmodel.Item{
					{SkuId: 1, Count: 1},
				},
			},
			mock: func(ctx context.Context, sh *ShardManagerProviderMock, pool db.Pool, conn pgxmock.PgxPoolIface, s *StocksProviderMock, o *OrdersProviderMock, n *NotifierProviderMock, i inputData, wantResult int64, wantErr error) {
				tx := getTxMock(ctx, pool, conn)
				sh.GetByOrderIdMock.Expect(i.orderId).Return(pool, nil)
				conn.ExpectBegin()
				conn.ExpectBegin()
				conn.ExpectCommit()
				conn.ExpectCommit()
				n.PublishMock.When(ctx, tx, i.orderId, ordermodel.StatusPaid).Then(nil)
				o.GetOrderMock.When(ctx, tx, i.orderId).Then(i.userId, ordermodel.StatusAwaiting, i.items, nil)
				s.RemoveMock.When(ctx, tx, i.items).Then(nil)
				o.SetStatusMock.When(ctx, tx, i.orderId, ordermodel.StatusPaid).Then(nil)
			},
		},
	}

	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			mc := minimock.NewController(t)
			conn, err := pgxmock.NewPool()

			if err != nil {
				log.Fatalln(err.Error())
			}

			pool, err := db.NewDbClientFromConnection([]db.Tx{conn}, []db.Tx{})

			if err != nil {
				log.Fatalln(err.Error())
			}

			ctx := context.Background()
			shardManager := NewShardManagerProviderMock(mc)
			stocksProviderMock := NewStocksProviderMock(mc)
			ordersProviderMock := NewOrdersProviderMock(mc)
			notifierProviderMock := NewNotifierProviderMock(mc)
			lomsService := NewLomsService(
				shardManager,
				pool,
				stocksProviderMock,
				ordersProviderMock,
				notifierProviderMock,
			)

			test.mock(ctx, shardManager, pool, conn, stocksProviderMock, ordersProviderMock, notifierProviderMock, test.inputData, test.wantResult, test.wantErr)
			gotErr := lomsService.PayOrder(ctx, test.inputData.orderId)

			if test.wantErr == nil {
				require.NoError(t, gotErr)
			} else {
				require.ErrorAs(t, gotErr, &test.wantErr)
			}
		})
	}
}

func TestLomsService_GetAvailableStocks(t *testing.T) {
	tests := []struct {
		name       string
		inputData  inputData
		mock       func(ctx context.Context, sh *ShardManagerProviderMock, client db.Pool, pp pgxmock.PgxPoolIface, l *StocksProviderMock, p *OrdersProviderMock, n *NotifierProviderMock, i inputData, wantResult uint64, wantErr error)
		wantResult uint64
		wantErr    error
	}{
		{
			name: "should be error if failed to find stocks",
			inputData: inputData{
				skuId: 3,
			},
			mock: func(ctx context.Context, sh *ShardManagerProviderMock, pool db.Pool, conn pgxmock.PgxPoolIface, s *StocksProviderMock, o *OrdersProviderMock, n *NotifierProviderMock, i inputData, wantResult uint64, wantErr error) {
				tx := getTxMock(ctx, pool, conn)
				conn.ExpectBegin()
				conn.ExpectCommit()
				s.GetByIdMock.Expect(ctx, tx, i.skuId).Return(0, wantErr)
			},
			wantErr: errors.New("failed to find stocks"),
		},
		{
			name: "should be successful",
			inputData: inputData{
				skuId: 3,
			},
			mock: func(ctx context.Context, sh *ShardManagerProviderMock, pool db.Pool, conn pgxmock.PgxPoolIface, s *StocksProviderMock, o *OrdersProviderMock, n *NotifierProviderMock, i inputData, wantResult uint64, wantErr error) {
				tx := getTxMock(ctx, pool, conn)
				conn.ExpectBegin()
				conn.ExpectCommit()
				s.GetByIdMock.Expect(ctx, tx, i.skuId).Return(wantResult, nil)
			},
			wantResult: 100,
		},
	}

	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			mc := minimock.NewController(t)
			conn, err := pgxmock.NewPool()

			if err != nil {
				log.Fatalln(err.Error())
			}

			pool, err := db.NewDbClientFromConnection([]db.Tx{conn}, []db.Tx{})

			if err != nil {
				log.Fatalln(err.Error())
			}

			ctx := context.Background()
			shardManager := NewShardManagerProviderMock(mc)
			stocksProviderMock := NewStocksProviderMock(mc)
			ordersProviderMock := NewOrdersProviderMock(mc)
			notifierProviderMock := NewNotifierProviderMock(mc)
			lomsService := NewLomsService(
				shardManager,
				pool,
				stocksProviderMock,
				ordersProviderMock,
				notifierProviderMock,
			)
			test.mock(ctx, shardManager, pool, conn, stocksProviderMock, ordersProviderMock, notifierProviderMock, test.inputData, test.wantResult, test.wantErr)
			gotResult, gotErr := lomsService.GetAvailableStocks(ctx, test.inputData.skuId)

			if test.wantErr == nil {
				require.NoError(t, gotErr)
			} else {
				require.ErrorAs(t, gotErr, &test.wantErr)
			}

			require.Equal(t, test.wantResult, gotResult)
		})
	}
}
