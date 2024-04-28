package db

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"log"
)

type Tx interface {
	Begin(ctx context.Context) (pgx.Tx, error)
	CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error)
	SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults
	Exec(ctx context.Context, sql string, arguments ...interface{}) (commandTag pgconn.CommandTag, err error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
}

type TxType int

const (
	WriteOrRead TxType = iota + 1
	ReadOnly
)

func WithTransaction(ctx context.Context, db Pool, trType TxType, fn func(ctx context.Context, tx Tx) error) error {
	pool := db.Get(trType)
	tx, err := pool.Begin(ctx)

	if err != nil {
		return err
	}

	defer func(tx pgx.Tx, ctx context.Context) {
		if err == nil {
			return
		}

		rollbackErr := tx.Rollback(ctx)

		if rollbackErr != nil {
			return
		}
	}(tx, ctx)

	err = fn(ctx, tx)

	if err != nil {
		return err
	}

	err = tx.Commit(ctx)

	return nil
}

func WithTransactions(ctx context.Context, pools []Pool, trType TxType, fn func(ctx context.Context, tx []Tx) error) error {
	var err error
	transactions := make([]pgx.Tx, 0)

	for _, pool := range pools {
		p := pool.Get(trType)
		tx, poolErr := p.Begin(ctx)

		if poolErr != nil {
			err = poolErr
			return err
		}

		transactions = append(transactions, tx)
	}

	defer func(transactions []pgx.Tx, ctx context.Context) {
		if err == nil {
			return
		}

		for _, tx := range transactions {
			if err = tx.Rollback(ctx); err != nil {
				log.Println("rollback failed:", err)
			}
		}
	}(transactions, ctx)

	transactions1 := make([]Tx, 0)

	for _, t := range transactions {
		transactions1 = append(transactions1, t)
	}

	err = fn(ctx, transactions1)

	if err != nil {
		return err
	}

	for _, tx := range transactions {
		err = tx.Commit(ctx)
	}

	return nil
}
