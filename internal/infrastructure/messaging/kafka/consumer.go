package kafka

import (
	"authentication/internal/application/contracts/messaging"
	"context"
	"log"

	"github.com/segmentio/kafka-go"
)

func (k *KafkaBus) Subscribe(topic string, handler func(messaging.Event)) error {
	k.reader = kafka.NewReader(kafka.ReaderConfig{
		Brokers: k.brokers,
		Topic:   topic,
		GroupID: "default-group",
	})

	go func() {
		for {
			msg, err := k.reader.ReadMessage(context.Background())
			if err != nil {
				log.Printf("Kafka consumer error: %v", err)
				continue
			}

			handler(messaging.Event{
				Name:    topic,
				Payload: msg.Value,
			})
		}
	}()
	return nil
}
