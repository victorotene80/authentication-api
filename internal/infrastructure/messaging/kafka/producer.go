package kafka

import (
	"authentication/shared/logging"
	"authentication/internal/application/contracts/messaging"
	"context"
	"fmt"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type KafkaBus struct {
	writer  *kafka.Writer
	reader  *kafka.Reader
	brokers []string
	logger  logging.Logger
}

func NewKafkaBus(brokerURL string, logger logging.Logger) (*KafkaBus, error) {
	brokers := []string{brokerURL}
	
	writer := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireAll,
		Async:        false,
	}

	return &KafkaBus{
		writer:  writer,
		brokers: brokers,
		logger:  logger.With(zap.String("component", "kafka_bus")),
	}, nil
}

func (k *KafkaBus) Publish(event messaging.Event) error {
	ctx := context.Background()
	
	err := k.writer.WriteMessages(ctx,
		kafka.Message{
			Topic: event.Name,
			Value: event.Payload,
		},
	)

	if err != nil {
		k.logger.Error(ctx, "failed to publish message to kafka",
			zap.String("topic", event.Name),
			zap.Error(err),
		)
		return fmt.Errorf("kafka publish failed: %w", err)
	}

	return nil
}

func (k *KafkaBus) Close() error {
	ctx := context.Background()
	var lastErr error

	if k.writer != nil {
		if err := k.writer.Close(); err != nil {
			k.logger.Error(ctx, "error closing kafka writer", zap.Error(err))
			lastErr = err
		}
	}

	if k.reader != nil {
		if err := k.reader.Close(); err != nil {
			k.logger.Error(ctx, "error closing kafka reader", zap.Error(err))
			lastErr = err
		}
	}

	return lastErr
}

