package kafka

import (
	"time"

	"github.com/IBM/sarama"
)

type Config struct {
	Brokers []string
}

func PrepareConfig(opts ...Option) *sarama.Config {
	c := sarama.NewConfig()

	// алгоритм выбора партиции
	{
		c.Producer.Partitioner = sarama.NewRoundRobinPartitioner
	}

	// acks параметр
	{
		c.Producer.RequiredAcks = sarama.WaitForAll
	}

	// семантика exactly once
	{
		c.Producer.Idempotent = true
	}

	// повторы ошибочных отправлений
	{
		c.Producer.Retry.Max = 2
		c.Producer.Retry.Backoff = 10 * time.Second
	}

	{
		c.Net.MaxOpenRequests = 1
	}

	// сжатие на клиенте
	{
		c.Producer.CompressionLevel = sarama.CompressionLevelDefault
		c.Producer.Compression = sarama.CompressionGZIP
	}

	{
		/*
			Если эта конфигурация используется для создания `SyncProducer`, оба параметра должны быть установлены
			в значение true, и вы не не должны читать данные из каналов, поскольку это уже делает продьюсер под капотом.
		*/
		c.Producer.Return.Successes = true
		c.Producer.Return.Errors = true
	}

	for _, opt := range opts {
		_ = opt.Apply(c)
	}

	return c
}
