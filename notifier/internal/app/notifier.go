package app

import (
	"context"
	"fmt"
	"github.com/IBM/sarama"
	"log"
	"route256.ozon.ru/project/notifier/config"
	"route256.ozon.ru/project/notifier/internal/infra/kafka"
	"route256.ozon.ru/project/notifier/internal/transport"
	"sync"
)

type Notifier struct {
	consumer *kafka.ConsumerGroup
}

func NewNotifier(appConfig config.Config) *Notifier {
	consumerGroup, err := kafka.NewConsumerGroup(
		appConfig,
		transport.NewNotifierConsumerHandler(),
		kafka.WithOffsetsInitial(sarama.OffsetOldest),
	)

	if err != nil {
		log.Fatal(err)
	}

	return &Notifier{
		consumer: consumerGroup,
	}
}

func (n Notifier) Run(ctx context.Context, wg *sync.WaitGroup) {
	defer n.consumer.Close()

	runCGErrorHandler(ctx, n.consumer, wg)
	n.consumer.Run(ctx, wg)

	wg.Wait()
}

func runCGErrorHandler(ctx context.Context, cg sarama.ConsumerGroup, wg *sync.WaitGroup) {
	wg.Add(1)

	go func() {
		defer wg.Done()

		for {
			select {
			case chErr, ok := <-cg.Errors():
				if !ok {
					fmt.Println("Consumer: chan closed")
					return
				}

				fmt.Printf("Consumer: error: %s\n", chErr)
			case <-ctx.Done():
				fmt.Printf("Consumer: ctx closed: %s\n", ctx.Err().Error())
				return
			}
		}
	}()
}
