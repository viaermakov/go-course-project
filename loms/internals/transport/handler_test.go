package transport

import (
	"errors"
	"github.com/gojuno/minimock/v3"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"route256.ozon.ru/project/loms/model/itemmodel"
	"route256.ozon.ru/project/loms/model/ordermodel"
	"route256.ozon.ru/project/loms/pkg/api/order/v1"
	"testing"
)

func TestLomsHandler_OrderCreate(t *testing.T) {
	tests := []struct {
		name       string
		inputData  *order.OrderCreateRequest
		mock       func(l *LomsProviderMock, i *order.OrderCreateRequest, wantResult *order.OrderCreateResponse, wantErr codes.Code)
		wantResult *order.OrderCreateResponse
		wantErr    codes.Code
	}{
		{
			name: "should be error if failed to create order",
			inputData: &order.OrderCreateRequest{
				User:  1,
				Items: []*order.OrderItem{{Sku: 1, Count: 1}},
			},
			mock: func(l *LomsProviderMock, i *order.OrderCreateRequest, wantResult *order.OrderCreateResponse, wantErr codes.Code) {
				l.CreateOrderMock.Expect(minimock.AnyContext, i.User, []itemmodel.Item{{SkuId: 1, Count: 1}}).Return(0, errors.New("failed to create order"))
			},
			wantErr: codes.Internal,
		},
		{
			name: "should be successful",
			inputData: &order.OrderCreateRequest{
				User:  1,
				Items: []*order.OrderItem{},
			},
			mock: func(l *LomsProviderMock, i *order.OrderCreateRequest, wantResult *order.OrderCreateResponse, wantErr codes.Code) {
				l.CreateOrderMock.Expect(minimock.AnyContext, i.User, []itemmodel.Item{}).Return(wantResult.OrderID, nil)
			},
			wantResult: &order.OrderCreateResponse{
				OrderID: 2,
			},
		},
	}

	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			mc := minimock.NewController(t)
			lomsProviderMock := NewLomsProviderMock(mc)
			lomsService := NewLomsHandler(lomsProviderMock)

			test.mock(lomsProviderMock, test.inputData, test.wantResult, test.wantErr)
			gotResult, gotErr := lomsService.OrderCreate(minimock.AnyContext, test.inputData)

			require.Equal(t, test.wantErr, status.Code(gotErr))
			require.Equal(t, test.wantResult, gotResult)
		})
	}
}

func TestLomsHandler_OrderCancel(t *testing.T) {
	tests := []struct {
		name       string
		inputData  *order.OrderCancelRequest
		mock       func(l *LomsProviderMock, i *order.OrderCancelRequest, wantResult *order.OrderCancelResponse, wantErr codes.Code)
		wantResult *order.OrderCancelResponse
		wantErr    codes.Code
	}{
		{
			name: "should be error if failed to cancel order",
			inputData: &order.OrderCancelRequest{
				OrderID: 1,
			},
			mock: func(l *LomsProviderMock, i *order.OrderCancelRequest, wantResult *order.OrderCancelResponse, wantErr codes.Code) {
				l.CancelOrderMock.Expect(minimock.AnyContext, i.OrderID).Return(errors.New("failed to cancel order"))
			},
			wantErr: codes.Internal,
		},
		{
			name: "should be successful",
			inputData: &order.OrderCancelRequest{
				OrderID: 1,
			},
			mock: func(l *LomsProviderMock, i *order.OrderCancelRequest, wantResult *order.OrderCancelResponse, wantErr codes.Code) {
				l.CancelOrderMock.Expect(minimock.AnyContext, i.OrderID).Return(nil)
			},
			wantResult: &order.OrderCancelResponse{},
		},
	}

	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			mc := minimock.NewController(t)
			lomsProviderMock := NewLomsProviderMock(mc)
			lomsService := NewLomsHandler(lomsProviderMock)

			test.mock(lomsProviderMock, test.inputData, test.wantResult, test.wantErr)
			gotResult, gotErr := lomsService.OrderCancel(minimock.AnyContext, test.inputData)

			require.Equal(t, test.wantErr, status.Code(gotErr))
			require.Equal(t, test.wantResult, gotResult)
		})
	}
}

func TestLomsHandler_OrderInfo(t *testing.T) {
	tests := []struct {
		name       string
		inputData  *order.OrderInfoRequest
		mock       func(l *LomsProviderMock, i *order.OrderInfoRequest, wantResult *order.OrderInfoResponse, wantErr codes.Code)
		wantResult *order.OrderInfoResponse
		wantErr    codes.Code
	}{
		{
			name: "should be successful",
			inputData: &order.OrderInfoRequest{
				OrderID: 1,
			},
			mock: func(l *LomsProviderMock, i *order.OrderInfoRequest, wantResult *order.OrderInfoResponse, wantErr codes.Code) {
				orderInfo := ordermodel.Info{
					Status: ordermodel.StatusAwaiting,
					User:   2,
					Items: []itemmodel.Item{
						{SkuId: 1, Count: 2},
					},
				}

				l.GetOrderMock.Expect(minimock.AnyContext, i.OrderID).Return(&orderInfo, nil)
			},
			wantResult: &order.OrderInfoResponse{
				Status: order.OrderStatus_AWAITING,
				User:   2,
				Items:  []*order.OrderItem{{Sku: 1, Count: 2}},
			},
		},
		{
			name: "should be error if failed to get order",
			inputData: &order.OrderInfoRequest{
				OrderID: 1,
			},
			mock: func(l *LomsProviderMock, i *order.OrderInfoRequest, wantResult *order.OrderInfoResponse, wantErr codes.Code) {
				l.GetOrderMock.Expect(minimock.AnyContext, i.OrderID).Return(&ordermodel.Info{}, errors.New("failed to get order"))
			},
			wantErr: codes.Internal,
		},
	}

	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			mc := minimock.NewController(t)
			lomsProviderMock := NewLomsProviderMock(mc)
			lomsService := NewLomsHandler(lomsProviderMock)

			test.mock(lomsProviderMock, test.inputData, test.wantResult, test.wantErr)
			gotResult, gotErr := lomsService.OrderInfo(minimock.AnyContext, test.inputData)

			require.Equal(t, test.wantErr, status.Code(gotErr))
			require.Equal(t, test.wantResult, gotResult)
		})
	}
}

func TestLomsHandler_OrderPay(t *testing.T) {
	tests := []struct {
		name       string
		inputData  *order.OrderPayRequest
		mock       func(l *LomsProviderMock, i *order.OrderPayRequest, wantResult *order.OrderPayResponse, wantErr codes.Code)
		wantResult *order.OrderPayResponse
		wantErr    codes.Code
	}{
		{
			name: "should be successful",
			inputData: &order.OrderPayRequest{
				OrderID: 1,
			},
			mock: func(l *LomsProviderMock, i *order.OrderPayRequest, wantResult *order.OrderPayResponse, wantErr codes.Code) {
				l.PayOrderMock.Expect(minimock.AnyContext, i.OrderID).Return(nil)
			},
			wantResult: &order.OrderPayResponse{},
		},
		{
			name: "should be error if failed to pay order",
			inputData: &order.OrderPayRequest{
				OrderID: 1,
			},
			mock: func(l *LomsProviderMock, i *order.OrderPayRequest, wantResult *order.OrderPayResponse, wantErr codes.Code) {
				l.PayOrderMock.Expect(minimock.AnyContext, i.OrderID).Return(errors.New("failed to get order"))
			},
			wantErr: codes.Internal,
		},
	}

	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			mc := minimock.NewController(t)
			lomsProviderMock := NewLomsProviderMock(mc)
			lomsService := NewLomsHandler(lomsProviderMock)

			test.mock(lomsProviderMock, test.inputData, test.wantResult, test.wantErr)
			gotResult, gotErr := lomsService.OrderPay(minimock.AnyContext, test.inputData)

			require.Equal(t, test.wantErr, status.Code(gotErr))
			require.Equal(t, test.wantResult, gotResult)
		})
	}
}

func TestLomsHandler_StocksInfo(t *testing.T) {
	tests := []struct {
		name       string
		inputData  *order.StocksInfoRequest
		mock       func(l *LomsProviderMock, i *order.StocksInfoRequest, wantResult *order.StocksInfoResponse, wantErr codes.Code)
		wantResult *order.StocksInfoResponse
		wantErr    codes.Code
	}{
		{
			name: "should be successful",
			inputData: &order.StocksInfoRequest{
				Sku: 1,
			},
			mock: func(l *LomsProviderMock, i *order.StocksInfoRequest, wantResult *order.StocksInfoResponse, wantErr codes.Code) {
				l.GetAvailableStocksMock.Expect(minimock.AnyContext, int64(i.Sku)).Return(wantResult.Count, nil)
			},
			wantResult: &order.StocksInfoResponse{
				Count: 10,
			},
		},
		{
			name: "should be error if failed to get stock info",
			inputData: &order.StocksInfoRequest{
				Sku: 1,
			},
			mock: func(l *LomsProviderMock, i *order.StocksInfoRequest, wantResult *order.StocksInfoResponse, wantErr codes.Code) {
				l.GetAvailableStocksMock.Expect(minimock.AnyContext, int64(i.Sku)).Return(0, errors.New("failed to get order"))
			},
			wantErr: codes.Internal,
		},
	}

	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			mc := minimock.NewController(t)
			lomsProviderMock := NewLomsProviderMock(mc)
			lomsService := NewLomsHandler(lomsProviderMock)

			test.mock(lomsProviderMock, test.inputData, test.wantResult, test.wantErr)
			gotResult, gotErr := lomsService.StocksInfo(minimock.AnyContext, test.inputData)

			require.Equal(t, test.wantErr, status.Code(gotErr))
			require.Equal(t, test.wantResult, gotResult)
		})
	}
}
