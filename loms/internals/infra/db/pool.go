package db

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"route256.ozon.ru/project/loms/config"
)

type Pool interface {
	Get(trxType TxType) Tx
}

type PoolClient struct {
	primaryConnections   []Tx
	secondaryConnections []Tx
	currentWriterIdx     int
	currentReaderIdx     int
}

func NewPools(ctx context.Context, connections []config.DbConnection) ([]Pool, error) {
	pools := make([]Pool, 0)

	for _, connection := range connections {
		pool, err := NewPool(ctx, []string{connection.Primary}, connection.Secondary)

		if err != nil {
			return nil, err
		}

		pools = append(pools, pool)
	}

	return pools, nil
}

func NewPool(ctx context.Context, primaryConnStrings []string, secondaryConnStrings []string) (*PoolClient, error) {
	primaryConnections := make([]Tx, 0)

	for _, primaryConnStr := range primaryConnStrings {
		conn, err := connectDb(ctx, primaryConnStr)

		if err != nil {
			return nil, err
		}

		primaryConnections = append(primaryConnections, conn)
	}

	secondaryConnections := make([]Tx, 0)

	for _, secondaryConnStr := range secondaryConnStrings {
		conn, err := connectDb(ctx, secondaryConnStr)

		if err != nil {
			return nil, err
		}

		secondaryConnections = append(secondaryConnections, conn)
	}

	return &PoolClient{
		primaryConnections:   primaryConnections,
		secondaryConnections: secondaryConnections,
		currentReaderIdx:     0,
	}, nil
}

func NewDbClientFromConnection(primaryConnStrings []Tx, secondaryConnStrings []Tx) (*PoolClient, error) {
	return &PoolClient{
		primaryConnections:   primaryConnStrings,
		secondaryConnections: secondaryConnStrings,
		currentReaderIdx:     0,
	}, nil
}

func (d *PoolClient) Get(trxType TxType) Tx {
	if trxType == WriteOrRead {
		return d.getWriter()
	}
	return d.getReader()
}

func (d *PoolClient) getWriter() Tx {
	if d.currentWriterIdx == len(d.primaryConnections) {
		d.currentWriterIdx = 0
	}

	conn := d.primaryConnections[d.currentWriterIdx]
	d.currentWriterIdx += 1

	return conn
}

func (d *PoolClient) getReader() Tx {
	if d.currentReaderIdx == len(d.secondaryConnections) {
		d.currentReaderIdx = 0
		return d.getWriter()
	}

	conn := d.secondaryConnections[d.currentReaderIdx]
	d.currentReaderIdx += 1

	return conn
}

func connectDb(ctx context.Context, dbConnStr string) (Tx, error) {
	conn, err := pgxpool.New(ctx, dbConnStr)

	if err != nil {
		return nil, err
	}

	err = conn.Ping(ctx)

	if err != nil {
		return nil, err
	}

	return conn, nil
}
