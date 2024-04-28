package kafka

import (
	"fmt"
	"github.com/IBM/sarama"
)

func NewSyncProducer(conf Config, opts ...Option) (sarama.SyncProducer, error) {
	config := PrepareConfig(opts...)
	syncProducer, err := sarama.NewSyncProducer(conf.Brokers, config)

	if err != nil {
		return nil, fmt.Errorf("NewSyncProducer failed: %w", err)
	}

	return syncProducer, nil
}
