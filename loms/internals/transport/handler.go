package transport

import (
	"context"
	"errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"route256.ozon.ru/project/loms/internals/repository/stocksrepo"
	"route256.ozon.ru/project/loms/model/itemmodel"
	"route256.ozon.ru/project/loms/model/ordermodel"
	servicepb "route256.ozon.ru/project/loms/pkg/api/order/v1"
)

var _ servicepb.OrderServer = (*LomsHandler)(nil)

type LomsHandler struct {
	service LomsProvider
	servicepb.UnimplementedOrderServer
}

type LomsProvider interface {
	CreateOrder(ctx context.Context, userId int64, items []itemmodel.Item) (int64, error)
	GetOrder(ctx context.Context, orderId int64) (*ordermodel.Info, error)
	PayOrder(ctx context.Context, orderId int64) error
	CancelOrder(ctx context.Context, orderId int64) error
	GetOrders(ctx context.Context, orderIds []int64) ([]*ordermodel.Info, error)
	GetAvailableStocks(ctx context.Context, skuId int64) (uint64, error)
}

func NewLomsHandler(service LomsProvider) *LomsHandler {
	return &LomsHandler{
		service: service,
	}
}

func (h LomsHandler) OrderCreate(context context.Context, req *servicepb.OrderCreateRequest) (*servicepb.OrderCreateResponse, error) {
	orderId, err := h.service.CreateOrder(
		context,
		req.User,
		prepareModelItems(req.Items),
	)

	if err = handleError(err); err != nil {
		return nil, err
	}

	return &servicepb.OrderCreateResponse{OrderID: orderId}, nil
}

func (h LomsHandler) OrderInfo(context context.Context, req *servicepb.OrderInfoRequest) (*servicepb.OrderInfoResponse, error) {
	orderInfo, err := h.service.GetOrder(context, req.OrderID)

	if err = handleError(err); err != nil {
		return nil, err
	}

	return &servicepb.OrderInfoResponse{
		Status: preparePbProductStatus(orderInfo.Status),
		Items:  preparePbItems(orderInfo.Items),
		User:   orderInfo.User,
	}, nil
}

func (h LomsHandler) OrderPay(context context.Context, req *servicepb.OrderPayRequest) (*servicepb.OrderPayResponse, error) {
	err := h.service.PayOrder(context, req.OrderID)

	if err = handleError(err); err != nil {
		return nil, err
	}

	return &servicepb.OrderPayResponse{}, nil
}

func (h LomsHandler) OrderCancel(context context.Context, req *servicepb.OrderCancelRequest) (*servicepb.OrderCancelResponse, error) {
	err := h.service.CancelOrder(context, req.OrderID)

	if err = handleError(err); err != nil {
		return nil, err
	}

	return &servicepb.OrderCancelResponse{}, nil
}

func (h LomsHandler) StocksInfo(context context.Context, req *servicepb.StocksInfoRequest) (*servicepb.StocksInfoResponse, error) {
	availableStocks, err := h.service.GetAvailableStocks(context, int64(req.Sku))

	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return &servicepb.StocksInfoResponse{Count: availableStocks}, nil
}

func (h LomsHandler) OrdersList(context context.Context, req *servicepb.OrdersListRequest) (*servicepb.OrdersListResponse, error) {
	orders, err := h.service.GetOrders(context, req.OrderIds)

	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return &servicepb.OrdersListResponse{Orders: preparePbOrders(orders)}, nil
}

func handleError(err error) error {
	switch {
	case errors.Is(err, stocksrepo.ErrProductsOutOfStock):
	case errors.Is(err, stocksrepo.ErrExceededReservedAmount):
		return status.Errorf(codes.FailedPrecondition, err.Error())
	case err != nil:
		return status.Errorf(codes.Internal, err.Error())
	}

	return nil
}

func prepareModelItems(items []*servicepb.OrderItem) []itemmodel.Item {
	modelItems := make([]itemmodel.Item, 0)

	for _, item := range items {
		modelItems = append(modelItems, itemmodel.Item{
			Count: uint16(item.Count),
			SkuId: item.Sku,
		})
	}

	return modelItems
}

func preparePbItems(items []itemmodel.Item) []*servicepb.OrderItem {
	pbItems := make([]*servicepb.OrderItem, 0)

	for _, item := range items {
		pbItems = append(pbItems, &servicepb.OrderItem{
			Count: uint32(item.Count),
			Sku:   item.SkuId,
		})
	}

	return pbItems
}

func preparePbOrders(items []*ordermodel.Info) []*servicepb.OrderInfo {
	pbItems := make([]*servicepb.OrderInfo, 0)

	for _, item := range items {
		pbItems = append(pbItems, &servicepb.OrderInfo{
			OrderID: item.OrderId,
			User:    item.User,
			Items:   preparePbItems(item.Items),
		})
	}

	return pbItems
}

func preparePbProductStatus(status ordermodel.Status) servicepb.OrderStatus {
	newStatus := servicepb.OrderStatus_UNSPECIFIED

	switch status {
	case ordermodel.StatusNew:
		newStatus = servicepb.OrderStatus_NEW
	case ordermodel.StatusAwaiting:
		newStatus = servicepb.OrderStatus_AWAITING
	case ordermodel.StatusFailed:
		newStatus = servicepb.OrderStatus_FAILED
	case ordermodel.StatusPaid:
		newStatus = servicepb.OrderStatus_PAYED
	case ordermodel.StatusCanceled:
		newStatus = servicepb.OrderStatus_CANCELLED
	}

	return newStatus
}
