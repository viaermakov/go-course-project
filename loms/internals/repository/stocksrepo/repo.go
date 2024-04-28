package stocksrepo

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"route256.ozon.ru/project/loms/internals/infra/db"
	stocksrepo "route256.ozon.ru/project/loms/internals/repository/stocksrepo/sqlc"
	"route256.ozon.ru/project/loms/model/itemmodel"
)

type StocksRepo struct {
}

var (
	ErrUnknownProductId       = errors.New("product id is unknown")
	ErrProductsOutOfStock     = errors.New("product is out of stock")
	ErrExceededReservedAmount = errors.New("product amount exceeds the reserved quantity")
)

func NewRepo() *StocksRepo {
	return &StocksRepo{}
}

func (repo *StocksRepo) Reserve(ctx context.Context, tx db.Tx, items []itemmodel.Item) error {
	q := stocksrepo.New(tx)
	err := repo.available(ctx, q, items)

	if err != nil {
		return err
	}

	for _, item := range items {
		stock, err := q.GetStock(ctx, int64(item.SkuId))

		if err != nil {
			return handleSqlError(err)
		}

		err = q.ReserveStock(ctx, stocksrepo.ReserveStockParams{
			SkuID:     int64(item.SkuId),
			Reserved:  stock.Reserved + int64(item.Count),
			Available: stock.Available - int64(item.Count),
		})

		if err != nil {
			return handleSqlError(err)
		}
	}

	return nil
}

func (repo *StocksRepo) Remove(ctx context.Context, tx db.Tx, items []itemmodel.Item) error {
	q := stocksrepo.New(tx)
	err := repo.canBeReserved(ctx, q, items)

	if err != nil {
		return err
	}

	for _, item := range items {
		stock, err := q.GetStock(ctx, int64(item.SkuId))

		if err != nil {
			return handleSqlError(err)
		}

		err = q.RemoveStock(ctx, stocksrepo.RemoveStockParams{
			SkuID:    int64(item.SkuId),
			Reserved: stock.Reserved - int64(item.Count),
		})

		if err != nil {
			return handleSqlError(err)
		}
	}

	return nil
}

func (repo *StocksRepo) Cancel(ctx context.Context, tx db.Tx, items []itemmodel.Item) error {
	q := stocksrepo.New(tx)

	err := repo.available(ctx, q, items)

	if err != nil {
		return err
	}

	for _, item := range items {
		stock, err := q.GetStock(ctx, int64(item.SkuId))

		if err != nil {
			return handleSqlError(err)
		}

		err = q.ReserveStock(ctx, stocksrepo.ReserveStockParams{
			SkuID:     int64(item.SkuId),
			Reserved:  stock.Reserved - int64(item.Count),
			Available: stock.Available + int64(item.Count),
		})

		if err != nil {
			return handleSqlError(err)
		}
	}

	return nil
}

func (repo *StocksRepo) GetById(ctx context.Context, tx db.Tx, skuId int64) (uint64, error) {
	q := stocksrepo.New(tx)
	stockData, err := q.GetStock(ctx, skuId)

	if err != nil {
		return 0, handleSqlError(err)
	}

	return uint64(stockData.Available), nil
}

func (repo *StocksRepo) available(ctx context.Context, q *stocksrepo.Queries, items []itemmodel.Item) error {
	for _, item := range items {
		stock, err := q.GetStock(ctx, int64(item.SkuId))

		if err != nil {
			return handleSqlError(err)
		}

		if stock.Available < int64(item.Count) {
			return ErrExceededReservedAmount
		}
	}

	return nil
}

func (repo *StocksRepo) canBeReserved(ctx context.Context, q *stocksrepo.Queries, items []itemmodel.Item) error {
	for _, item := range items {
		stock, err := q.GetStock(ctx, int64(item.SkuId))

		if err != nil {
			return handleSqlError(err)
		}

		if stock.Reserved < int64(item.Count) {
			return ErrExceededReservedAmount
		}
	}

	return nil
}

func handleSqlError(err error) error {
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return ErrUnknownProductId
	default:
		return err
	}
}
