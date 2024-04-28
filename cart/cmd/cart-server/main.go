package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"route256.ozon.ru/project/cart/config"
	"route256.ozon.ru/project/cart/internals/app"
	"route256.ozon.ru/project/cart/internals/infra/errgrp"
	"syscall"
	"time"
)

func main() {
	parentCtx := context.Background()

	appConfig := config.NewConfig()
	server := app.NewCartServer(parentCtx, appConfig)

	ctx, stop := signal.NotifyContext(parentCtx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	g, cancelCtx := errgrp.WithContext(ctx)

	log.Printf("Starting server")

	g.Go(func() error {
		err := server.ListenAndServe()

		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}

		return nil
	})

	g.Go(func() error {
		<-cancelCtx.Done()

		log.Printf("Got interrupted signal")

		ctx, cancel := context.WithTimeout(parentCtx, 5*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			return errors.Join(errors.New("Server shutdown returned an err: %v\n"), err)
		}

		return nil
	})

	if err := g.Wait(); err != nil {
		fmt.Printf("exit reason: %s \n", err)
	}

	log.Println("Server successfully closed")
}
