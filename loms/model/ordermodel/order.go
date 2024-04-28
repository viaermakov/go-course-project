package ordermodel

import "route256.ozon.ru/project/loms/model/itemmodel"

type Status int

const (
	StatusNew Status = iota + 1
	StatusAwaiting
	StatusFailed
	StatusPaid
	StatusCanceled
)

type Info struct {
	OrderId int64
	Status  Status
	User    int64
	Items   []itemmodel.Item
}

func (s Status) String() string {
	switch s {
	case StatusNew:
		return "new"
	case StatusPaid:
		return "paid"
	case StatusCanceled:
		return "canceled"
	case StatusAwaiting:
		return "awaiting"
	case StatusFailed:
		return "failed"
	}

	return "unknown"
}
