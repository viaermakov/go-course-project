package migrations

import (
	"context"
	"embed"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"log"
)

//go:embed *
var embedMigrations embed.FS

func ApplyMigrations(dbConnStr string, dir string) {
	pool, err := connectDb(context.Background(), dbConnStr)

	if err != nil {
		log.Fatalln("failed to run migrations" + err.Error())
		return
	}

	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalln("failed to run migrations" + err.Error())
	}

	db := stdlib.OpenDBFromPool(pool)

	if err := goose.Up(db, dir); err != nil {
		log.Fatalln("failed to run migrations" + err.Error())
	}

	if err := db.Close(); err != nil {
		log.Fatalln("failed to run migrations" + err.Error())
	}
}

func RollbackMigrations(dbConnStr string, dir string) {
	pool, err := connectDb(context.Background(), dbConnStr)

	if err != nil {
		log.Fatalln("failed to run migrations" + err.Error())
		return
	}

	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalln("failed to run migrations" + err.Error())
	}

	db := stdlib.OpenDBFromPool(pool)

	if err := goose.Down(db, dir); err != nil {
		log.Fatalln("failed to rollback migrations" + err.Error())
	}

	if err := db.Close(); err != nil {
		log.Fatalln("failed to rollback migrations" + err.Error())
	}
}

func connectDb(ctx context.Context, dbConnStr string) (*pgxpool.Pool, error) {
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
