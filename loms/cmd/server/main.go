package main

import (
	"context"
	"errors"
	"fmt"
	"golang.org/x/sync/errgroup"
	"log"
	"net/http"
	"os/signal"
	"route256.ozon.ru/project/loms/config"
	"route256.ozon.ru/project/loms/internals/app"
	"route256.ozon.ru/project/loms/migrations"
	"syscall"
	"time"
)

func main() {
	parentCtx := context.Background()
	appConfig := config.NewConfig()

	runMigrations(appConfig)

	ctx, stop := signal.NotifyContext(parentCtx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	lomsServer, err := app.NewLomsGrpcServer(ctx, appConfig)

	if err != nil {
		log.Fatal("Failed to create grpc server:" + err.Error())
		return
	}

	g, cancelCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		if err = lomsServer.ListenAndServe(cancelCtx); err != nil {
			return err
		}
		return nil
	})

	gwServer, err := app.NewLomsGrpcGatewayServer(appConfig)

	if err != nil {
		log.Fatal("Failed to create http server:" + err.Error())
		return
	}

	g.Go(func() error {
		log.Printf("Serving gRPC-Gateway on %s\n", gwServer.Addr)

		if err = gwServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			return errors.New(fmt.Sprintf("failed to serve gateway: %s", err.Error()))
		}

		return nil
	})

	g.Go(func() error {
		<-cancelCtx.Done()

		log.Printf("Got interrupted signal")

		ctx, cancel := context.WithTimeout(parentCtx, 5*time.Second)
		defer cancel()

		if err := lomsServer.Shutdown(ctx); err != nil {
			return errors.New(fmt.Sprintf("failed to shutdown gRPC server: %v", err))
		}

		if err := gwServer.Shutdown(ctx); err != nil {
			return errors.New(fmt.Sprintf("failed to shutdown gRPC-Gateway server: %v", err))
		}

		return nil
	})

	if err := g.Wait(); err != nil {
		fmt.Printf("exit reason: %s \n", err)
	}

	log.Println("Server successfully closed")
}

func runMigrations(config config.Config) {
	for _, connection := range config.App.DbConnections {
		migrations.ApplyMigrations(connection.Primary, "common")
	}

	migrations.ApplyMigrations(config.App.StocksDbConnStr, "stocks")
}
