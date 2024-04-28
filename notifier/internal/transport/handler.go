package transport

import (
	"encoding/json"
	"fmt"
	"github.com/IBM/sarama"
	"log"
	"time"
)

type MessageEvent struct {
	Topic     string
	Partition int32
	Offset    int64
	Key       string
	Payload   messagePayload
}

type messagePayload struct {
	OrderId int64     `json:"orderId"`
	Time    time.Time `json:"time"`
	Message string    `json:"message"`
}

var _ sarama.ConsumerGroupHandler = (*NotifierConsumerHandler)(nil)

type NotifierConsumerHandler struct{}

func NewNotifierConsumerHandler() *NotifierConsumerHandler {
	return &NotifierConsumerHandler{}
}

func (h *NotifierConsumerHandler) Setup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (h *NotifierConsumerHandler) Cleanup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (h *NotifierConsumerHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message, ok := <-claim.Messages():
			if !ok {
				return nil
			}

			formattedMessage, err := convertMessage(message)

			if err != nil {
				log.Println("Error unmarshalling message", err)
			}

			log.Printf("Message claimed: %v", formattedMessage.String())

			session.MarkMessage(message, "")

		case <-session.Context().Done():
			return nil
		}
	}
}

func convertMessage(in *sarama.ConsumerMessage) (*MessageEvent, error) {
	var payload messagePayload
	err := json.Unmarshal(in.Value, &payload)

	if err != nil {
		return nil, err
	}

	return &MessageEvent{
		Topic:     in.Topic,
		Partition: in.Partition,
		Offset:    in.Offset,
		Key:       string(in.Key),
		Payload:   payload,
	}, nil
}

func (m *MessageEvent) String() string {
	payload := fmt.Sprintf("time: %v, message: %v", m.Payload.Time, m.Payload.Message)
	return fmt.Sprintf("Topic: %s, Partition: %d, Offset: %d, Key: %v, Message: %v", m.Topic, m.Partition, m.Offset, m.Key, payload)
}
