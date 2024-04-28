package notifierrepo

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"route256.ozon.ru/project/loms/internals/infra/db"
	notifierrepo "route256.ozon.ru/project/loms/internals/repository/notifierrepo/sqlc"
	"route256.ozon.ru/project/loms/model/ordermodel"
	"time"
)

type OrderNotifierRepo struct {
}

type Info struct {
	Status  ordermodel.Status
	Time    time.Time
	OrderId int64
}

type Status int

const (
	StatusAwaiting Status = iota + 1
	StatusCompleted
)

func NewRepo() *OrderNotifierRepo {
	return &OrderNotifierRepo{}
}

func (o *OrderNotifierRepo) Publish(ctx context.Context, tx db.Tx, orderId int64, orderStatus ordermodel.Status) error {
	q := notifierrepo.New(tx)

	err := q.AddEvent(ctx, notifierrepo.AddEventParams{
		OrderID:     orderId,
		SendStatus:  int32(StatusAwaiting),
		OrderStatus: int32(orderStatus),
		CreatedAt:   pgtype.Timestamp{Time: time.Now().UTC(), Valid: true},
	})

	return err
}

func (o *OrderNotifierRepo) MarkAsSent(ctx context.Context, tx db.Tx, orderId int64) error {
	q := notifierrepo.New(tx)

	err := q.UpdateEvent(ctx, notifierrepo.UpdateEventParams{
		OrderID:    orderId,
		SendStatus: int32(StatusCompleted),
		SendAt:     pgtype.Timestamp{Time: time.Now().UTC(), Valid: true},
	})

	return err
}

func (o *OrderNotifierRepo) RetrieveEvents(ctx context.Context, tx db.Tx) ([]Info, error) {
	q := notifierrepo.New(tx)

	awaitingEvents, err := q.GetScheduledEvents(ctx, int32(StatusAwaiting))

	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return []Info{}, err
	}

	events := make([]Info, 0)

	for _, event := range awaitingEvents {
		events = append(events, Info{
			Status:  ordermodel.Status(event.OrderStatus),
			OrderId: event.OrderID,
			Time:    event.CreatedAt.Time.UTC(),
		})
	}

	return events, nil
}

func (o *OrderNotifierRepo) MarkAllAsSent(ctx context.Context, tx db.Tx, events []Info) error {
	for _, event := range events {
		err := o.MarkAsSent(ctx, tx, event.OrderId)

		if err != nil {
			return err
		}
	}

	return nil
}
