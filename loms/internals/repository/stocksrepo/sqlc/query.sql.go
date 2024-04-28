// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0
// source: query.sql

package stocksrepo

import (
	"context"
)

const addStock = `-- name: AddStock :exec
insert into stocks (sku_id, available, reserved)
values ($1, $2, $3)
`

type AddStockParams struct {
	SkuID     int64
	Available int64
	Reserved  int64
}

func (q *Queries) AddStock(ctx context.Context, arg AddStockParams) error {
	_, err := q.db.Exec(ctx, addStock, arg.SkuID, arg.Available, arg.Reserved)
	return err
}

const getStock = `-- name: GetStock :one
select sku_id, available, reserved
from stocks
where sku_id = $1
`

func (q *Queries) GetStock(ctx context.Context, skuID int64) (Stock, error) {
	row := q.db.QueryRow(ctx, getStock, skuID)
	var i Stock
	err := row.Scan(&i.SkuID, &i.Available, &i.Reserved)
	return i, err
}

const removeStock = `-- name: RemoveStock :exec
update stocks
set reserved=$1
where sku_id = $2
`

type RemoveStockParams struct {
	Reserved int64
	SkuID    int64
}

func (q *Queries) RemoveStock(ctx context.Context, arg RemoveStockParams) error {
	_, err := q.db.Exec(ctx, removeStock, arg.Reserved, arg.SkuID)
	return err
}

const reserveStock = `-- name: ReserveStock :exec
update stocks
set reserved=$1, available=$2
where sku_id = $3
`

type ReserveStockParams struct {
	Reserved  int64
	Available int64
	SkuID     int64
}

func (q *Queries) ReserveStock(ctx context.Context, arg ReserveStockParams) error {
	_, err := q.db.Exec(ctx, reserveStock, arg.Reserved, arg.Available, arg.SkuID)
	return err
}
