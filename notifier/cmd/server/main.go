package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"route256.ozon.ru/project/notifier/config"
	"route256.ozon.ru/project/notifier/internal/app"
	"sync"
	"syscall"
)

func main() {
	var (
		wg        = &sync.WaitGroup{}
		appConfig = config.NewConfig()
		ctx       = runSignalHandler(context.Background(), wg)
	)

	log.Println("Running notifier")

	notifier := app.NewNotifier(appConfig)
	notifier.Run(ctx, wg)
}

func runSignalHandler(ctx context.Context, wg *sync.WaitGroup) context.Context {
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)

	sigCtx, cancel := context.WithCancel(ctx)

	wg.Add(1)

	go func() {
		defer fmt.Println("[signal] terminate")
		defer signal.Stop(sigterm)
		defer wg.Done()
		defer cancel()

		for {
			select {
			case sig, ok := <-sigterm:
				if !ok {
					fmt.Printf("[signal] signal chan closed: %s\n", sig.String())
					return
				}
				fmt.Printf("[signal] signal recv: %s\n", sig.String())
				return

			case _, ok := <-sigCtx.Done():
				if !ok {
					fmt.Println("[signal] context closed")
					return
				}
				fmt.Printf("[signal] ctx done: %s\n", ctx.Err().Error())
				return
			}
		}
	}()

	return sigCtx
}
