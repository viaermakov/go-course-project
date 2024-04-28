package kafka

import (
	"context"
	"github.com/IBM/sarama"
	"log"
	"route256.ozon.ru/project/notifier/config"
	"sync"
	"time"
)

type Config struct {
	GroupName string
	Topics    []string
}

type ConsumerGroup struct {
	sarama.ConsumerGroup
	handler sarama.ConsumerGroupHandler
	topics  []string
}

func NewConsumerGroup(appConfig config.Config, consumerGroupHandler sarama.ConsumerGroupHandler, opts ...Option) (*ConsumerGroup, error) {
	saramaConfig := sarama.NewConfig()
	saramaConfig.Version = sarama.MaxVersion
	saramaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest
	saramaConfig.Consumer.Group.ResetInvalidOffsets = true
	saramaConfig.Consumer.Group.Heartbeat.Interval = 3 * time.Second
	saramaConfig.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRoundRobin()}
	saramaConfig.Consumer.Group.Session.Timeout = 60 * time.Second
	saramaConfig.Consumer.Group.Rebalance.Timeout = 60 * time.Second
	saramaConfig.Consumer.Return.Errors = true
	saramaConfig.Consumer.Offsets.AutoCommit.Enable = true
	saramaConfig.Consumer.Offsets.AutoCommit.Interval = 5 * time.Second

	for _, opt := range opts {
		opt.Apply(saramaConfig)
	}

	consumerGroup, err := sarama.NewConsumerGroup(
		appConfig.Kafka.Brokers,
		appConfig.Consumer.GroupName,
		saramaConfig,
	)

	if err != nil {
		return nil, err
	}

	return &ConsumerGroup{
		ConsumerGroup: consumerGroup,
		handler:       consumerGroupHandler,
		topics:        []string{appConfig.Consumer.Topic},
	}, nil
}

func (c *ConsumerGroup) Run(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)

	go func() {
		defer wg.Done()

		for {
			if err := c.ConsumerGroup.Consume(ctx, c.topics, c.handler); err != nil {
				log.Printf("Error from consumeer: %v\n", err)
			}

			if ctx.Err() != nil {
				log.Printf("[consumer-group]: ctx closed: %s\n", ctx.Err().Error())
				return
			}
		}
	}()
}
