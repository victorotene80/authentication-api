package kafka

import (
	"authentication/internal/application/contracts/messaging"
	"context"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

func (k *KafkaBus) Subscribe(topic string, handler func(messaging.Event)) error {
	ctx := context.Background()

	k.reader = kafka.NewReader(kafka.ReaderConfig{
		Brokers:     k.brokers,
		Topic:       topic,
		GroupID:     "default-group",
		MinBytes:    1,
		MaxBytes:    10e6, 
		StartOffset: kafka.LastOffset,
	})

	k.logger.Info(ctx, "starting kafka consumer",
		zap.String("topic", topic),
		zap.Strings("brokers", k.brokers),
	)

	go func() {
		for {
			msg, err := k.reader.ReadMessage(context.Background())
			if err != nil {
				k.logger.Error(context.Background(), "kafka consumer error",
					zap.String("topic", topic),
					zap.Error(err),
				)
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
