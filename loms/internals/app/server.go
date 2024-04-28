package app

import (
	"context"
	"errors"
	"fmt"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"net/http"
	"route256.ozon.ru/project/loms/config"
	"route256.ozon.ru/project/loms/internals/infra/db"
	"route256.ozon.ru/project/loms/internals/infra/shardmanager"
	"route256.ozon.ru/project/loms/internals/repository/notifierrepo"
	"route256.ozon.ru/project/loms/internals/repository/ordersrepo"
	"route256.ozon.ru/project/loms/internals/repository/stocksrepo"
	"route256.ozon.ru/project/loms/internals/service/lomsservice"
	"route256.ozon.ru/project/loms/internals/service/notifierservice"
	"route256.ozon.ru/project/loms/internals/transport"
	"route256.ozon.ru/project/loms/internals/transport/middleware"
	desc "route256.ozon.ru/project/loms/pkg/api/order/v1"
)

type LomsGrpcServer struct {
	config   config.Config
	server   *grpc.Server
	notifier *notifierservice.NotifierService
}

type LomsHttpServer struct {
	config config.Config
	server *http.Server
}

func NewLomsGrpcServer(ctx context.Context, config config.Config) (*LomsGrpcServer, error) {
	dbPools, err := db.NewPools(ctx, config.App.DbConnections)

	if err != nil {
		return nil, err
	}

	stocksPool, err := db.NewPool(ctx, []string{config.App.StocksDbConnStr}, []string{})

	if err != nil {
		return nil, err
	}

	stocksStorage := stocksrepo.NewRepo()
	ordersStorage := ordersrepo.NewRepo()
	notifierStorage := notifierrepo.NewRepo()

	shardManager := shardmanager.New(
		shardmanager.GetShardFn(len(dbPools)),
		dbPools,
	)

	lomsHandler := transport.NewLomsHandler(
		lomsservice.NewLomsService(
			shardManager,
			stocksPool,
			stocksStorage,
			ordersStorage,
			notifierStorage,
		),
	)

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			middleware.Logger,
			middleware.Validate,
		),
	)

	producer, err := NewNotifierProducer(config, dbPools)

	if err != nil {
		return nil, err
	}

	reflection.Register(grpcServer)
	desc.RegisterOrderServer(grpcServer, lomsHandler)

	return &LomsGrpcServer{
		config:   config,
		server:   grpcServer,
		notifier: producer,
	}, nil
}

func (s *LomsGrpcServer) ListenAndServe(ctx context.Context) error {
	listen, err := net.Listen("tcp", fmt.Sprintf(":%d", s.config.App.GrpcPort))

	if err != nil {
		return errors.New("Failed to start s:" + err.Error())
	}

	go func() {
		if err = s.notifier.Run(ctx, s.config.Producer); err != nil {
			log.Fatal("Failed to start notifier:" + err.Error())
		}
	}()

	log.Printf("Serving gRPC-s on %v\n", s.config.App.GrpcPort)

	if err = s.server.Serve(listen); err != nil {
		return errors.New("Failed to start s:" + err.Error())
	}

	return nil
}

func (s *LomsGrpcServer) Shutdown(ctx context.Context) error {
	ok := make(chan struct{})

	go func() {
		s.server.GracefulStop()
		close(ok)
	}()

	select {
	case <-ok:
		return nil
	case <-ctx.Done():
		s.server.Stop()
		return ctx.Err()
	}
}

func NewLomsGrpcGatewayServer(config config.Config) (*http.Server, error) {
	conn, err := grpc.Dial(
		fmt.Sprintf(":%d", config.App.GrpcPort),
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		return nil, errors.New("Failed to dial:" + err.Error())
	}

	gatewayMux := runtime.NewServeMux()

	if err := desc.RegisterOrderHandler(context.Background(), gatewayMux, conn); err != nil {
		return nil, err
	}

	gwServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.App.HttpPort),
		Handler: middleware.WithHTTPLoggingMiddleware(gatewayMux),
	}

	return gwServer, nil
}
