package testcontainer

import (
	"context"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"route256.ozon.ru/project/loms/config"
	"time"
)

func CreatePgContainer(ctx context.Context, config config.Config) (*postgres.PostgresContainer, error) {
	return postgres.RunContainer(
		ctx,
		testcontainers.WithImage("docker.io/postgres:16.2-alpine"),
		postgres.WithDatabase(config.App.TestDbDatabase),
		postgres.WithUsername(config.App.TestDbUser),
		postgres.WithPassword(config.App.TestDbPassword),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(5*time.Second),
		),
	)
}
