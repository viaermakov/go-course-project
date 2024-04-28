package app

import (
	"github.com/IBM/sarama"
	"route256.ozon.ru/project/loms/config"
	"route256.ozon.ru/project/loms/internals/infra/db"
	"route256.ozon.ru/project/loms/internals/infra/kafka"
	"route256.ozon.ru/project/loms/internals/repository/notifierrepo"
	"route256.ozon.ru/project/loms/internals/service/notifierservice"
	"time"
)

func NewNotifierProducer(config config.Config, pools []db.Pool) (*notifierservice.NotifierService, error) {
	syncProducer, err := kafka.NewSyncProducer(config.Kafka,
		kafka.WithIdempotent(),
		kafka.WithRequiredAcks(sarama.WaitForAll),
		kafka.WithMaxOpenRequests(1),
		kafka.WithMaxRetries(5),
		kafka.WithRetryBackoff(10*time.Millisecond),
	)

	if err != nil {
		return nil, err
	}

	repo := notifierrepo.NewRepo()
	service := notifierservice.NewService(pools, syncProducer, repo)

	return service, nil
}
