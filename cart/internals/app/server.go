package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"route256.ozon.ru/project/cart/config"
	"route256.ozon.ru/project/cart/internals/clients"
	"route256.ozon.ru/project/cart/internals/infra/cacher"
	"route256.ozon.ru/project/cart/internals/service"
	"route256.ozon.ru/project/cart/internals/storage"
	"route256.ozon.ru/project/cart/internals/transport"
	"time"
)

type CartServer struct {
	server  *http.Server
	Handler http.Handler
	Config  config.Config
}

func NewCartServer(ctx context.Context, config config.Config) *CartServer {
	var server CartServer

	lomsClient, err := clients.NewLomsClient(config.LomsApi)

	if err != nil {
		log.Fatalf("failed to create LomsClient: %v", err)
		return nil
	}

	cartService := service.NewCartService(
		storage.NewInMemoryCartStorage(),
		clients.NewProductClient(
			cacher.New(
				ctx,
				config.RedisAddr,
				time.Duration(config.RedisTTL)*time.Second,
			),
		),
		lomsClient,
	)

	server.Handler = transport.NewHandler(cartService)
	server.Config = config

	return &server
}

func (c *CartServer) ListenAndServe() error {
	c.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", c.Config.HttpPort),
		Handler: c.Handler,
	}
	log.Printf("Serving cart server on %v\n", c.Config.HttpPort)
	return c.server.ListenAndServe()
}

func (c *CartServer) Shutdown(ctx context.Context) error {
	return c.server.Shutdown(ctx)
}
