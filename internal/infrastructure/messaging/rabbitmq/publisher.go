package rabbitmq

import (
	"authentication/internal/application/contracts/messaging"
	"authentication/shared/logging"
	"context"
	"fmt"

	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

type RabbitMQBus struct {
	channel    *amqp.Channel
	connection *amqp.Connection
	logger     logging.Logger
}

func NewRabbitMQBus(url string, logger logging.Logger) (*RabbitMQBus, error) {
	ctx := context.Background()

	conn, err := amqp.Dial(url)
	if err != nil {
		logger.Error(ctx, "failed to connect to RabbitMQ",
			zap.String("url", url),
			zap.Error(err),
		)
		return nil, fmt.Errorf("rabbitmq connection failed: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		logger.Error(ctx, "failed to open RabbitMQ channel", zap.Error(err))
		return nil, fmt.Errorf("rabbitmq channel creation failed: %w", err)
	}

	logger.Info(ctx, "connected to RabbitMQ", zap.String("url", url))

	return &RabbitMQBus{
		channel:    ch,
		connection: conn,
		logger:     logger.With(zap.String("component", "RabbitMQBus")),
	}, nil
}

func (r *RabbitMQBus) Publish(event messaging.Event) error {
	ctx := context.Background()

	err := r.channel.Publish(
		"",
		event.Name,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         event.Payload,
			DeliveryMode: amqp.Persistent,
		},
	)

	if err != nil {
		r.logger.Error(ctx, "failed to publish to RabbitMQ",
			zap.String("queue", event.Name),
			zap.Error(err),
		)
		return fmt.Errorf("rabbitmq publish failed: %w", err)
	}

	return nil
}

func (r *RabbitMQBus) Close() error {
	ctx := context.Background()
	var lastErr error

	if r.channel != nil {
		if err := r.channel.Close(); err != nil {
			r.logger.Error(ctx, "error closing RabbitMQ channel", zap.Error(err))
			lastErr = err
		}
	}

	if r.connection != nil {
		if err := r.connection.Close(); err != nil {
			r.logger.Error(ctx, "error closing RabbitMQ connection", zap.Error(err))
			lastErr = err
		}
	}

	return lastErr
}
