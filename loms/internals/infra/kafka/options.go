package kafka

import (
	"github.com/IBM/sarama"
	"time"
)

type Option interface {
	Apply(*sarama.Config) error
}

type optionFn func(*sarama.Config) error

func (fn optionFn) Apply(c *sarama.Config) error {
	return fn(c)
}

func WithProducerPartitioner(pfn sarama.PartitionerConstructor) Option {
	return optionFn(func(c *sarama.Config) error {
		c.Producer.Partitioner = pfn
		return nil
	})
}

func WithRequiredAcks(acks sarama.RequiredAcks) Option {
	return optionFn(func(c *sarama.Config) error {
		c.Producer.RequiredAcks = acks
		return nil
	})
}

func WithIdempotent() Option {
	return optionFn(func(c *sarama.Config) error {
		c.Producer.Idempotent = true
		return nil
	})
}

func WithMaxRetries(n int) Option {
	return optionFn(func(c *sarama.Config) error {
		c.Producer.Retry.Max = n
		return nil
	})
}

func WithRetryBackoff(d time.Duration) Option {
	return optionFn(func(c *sarama.Config) error {
		c.Producer.Retry.Backoff = d
		return nil
	})
}

func WithMaxOpenRequests(n int) Option {
	return optionFn(func(c *sarama.Config) error {
		c.Net.MaxOpenRequests = n
		return nil
	})
}

func WithProducerFlushMessages(n int) Option {
	return optionFn(func(c *sarama.Config) error {
		c.Producer.Flush.Messages = n
		return nil
	})
}

func WithProducerFlushFrequency(d time.Duration) Option {
	return optionFn(func(c *sarama.Config) error {
		c.Producer.Flush.Frequency = d
		return nil
	})
}
