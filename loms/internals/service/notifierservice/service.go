package notifierservice

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/IBM/sarama"
	"log"
	"route256.ozon.ru/project/loms/config"
	"route256.ozon.ru/project/loms/internals/infra/db"
	"route256.ozon.ru/project/loms/internals/repository/notifierrepo"
	"strconv"
	"time"
)

type NotifierRepoProvider interface {
	RetrieveEvents(ctx context.Context, tx db.Tx) ([]notifierrepo.Info, error)
	MarkAllAsSent(ctx context.Context, tx db.Tx, events []notifierrepo.Info) error
}

type ProducerProvider interface {
	SendMessage(msg *sarama.ProducerMessage) (partition int32, offset int64, err error)
	SendMessages(msgs []*sarama.ProducerMessage) error
	Close() error
}

type NotifierService struct {
	pools    []db.Pool
	repo     NotifierRepoProvider
	producer ProducerProvider
}

type MessageEvent struct {
	OrderId int64
	Time    time.Time
	Message string
}

func NewService(pools []db.Pool, producer ProducerProvider, repo NotifierRepoProvider) *NotifierService {
	return &NotifierService{
		pools:    pools,
		producer: producer,
		repo:     repo,
	}
}

func (o *NotifierService) Run(ctx context.Context, config config.ProducerConfig) error {
	log.Printf("Notifier: listening for events...")
	defer o.producer.Close()

	ticker := time.NewTicker(time.Duration(config.IntervalSec) * time.Second)

	for {
		select {
		case <-ctx.Done():
			ticker.Stop()
			log.Printf("Notifier: stopping service...")
			return nil
		case <-ticker.C:
			if err := o.sendEvents(ctx, config.Topic); err != nil {
				log.Printf("Notifier: failed to send events: %v", err)
			}
		}
	}
}

func (o *NotifierService) sendEvents(ctx context.Context, topic string) error {
	for _, pool := range o.pools {
		err := db.WithTransaction(ctx, pool, db.WriteOrRead, func(ctx context.Context, tx db.Tx) error {
			items, err := o.repo.RetrieveEvents(ctx, tx)

			if len(items) == 0 {
				return nil
			}

			messages := make([]*sarama.ProducerMessage, 0)

			for _, item := range items {
				bytes, itemErr := json.Marshal(createEvent(item))

				if itemErr != nil {
					err = itemErr
					break
				}

				msg := &sarama.ProducerMessage{
					Topic: topic,
					Key:   sarama.StringEncoder(strconv.FormatInt(item.OrderId, 10)),
					Value: sarama.ByteEncoder(bytes),
					Headers: []sarama.RecordHeader{
						{
							Key:   []byte("loms_service"),
							Value: []byte("order_status"),
						},
					},
					Timestamp: time.Now(),
				}

				messages = append(messages, msg)
			}

			if err != nil {
				return err
			}

			err = o.producer.SendMessages(messages)

			if err != nil {
				return err
			}

			return o.repo.MarkAllAsSent(ctx, tx, items)
		})

		if err != nil {
			return err
		}
	}

	return nil
}

func createEvent(info notifierrepo.Info) MessageEvent {
	return MessageEvent{
		OrderId: info.OrderId,
		Time:    info.Time,
		Message: fmt.Sprintf("[order_status] order %d changed status to %s\n", info.OrderId, info.Status.String()),
	}
}
