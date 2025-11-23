package nats

import (
	"authentication/shared/logging"
	"authentication/internal/application/contracts/messaging"
	"context"
	"fmt"

	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

type NATSBus struct {
	conn   *nats.Conn
	logger logging.Logger
}

func NewNATSBus(url string, logger logging.Logger) (*NATSBus, error) {
	conn, err := nats.Connect(url,
		nats.MaxReconnects(-1),
		nats.ReconnectWait(2),
	)
	if err != nil {
		logger.Error(context.Background(), "failed to connect to NATS",
			zap.String("url", url),
			zap.Error(err),
		)
		return nil, fmt.Errorf("nats connection failed: %w", err)
	}

	logger.Info(context.Background(), "connected to NATS", zap.String("url", url))

	return &NATSBus{
		conn:   conn,
		logger: logger.With(zap.String("component", "nats_bus")),
	}, nil
}

func (n *NATSBus) Publish(event messaging.Event) error {
	ctx := context.Background()

	err := n.conn.Publish(event.Name, event.Payload)
	if err != nil {
		n.logger.Error(ctx, "failed to publish to NATS",
			zap.String("subject", event.Name),
			zap.Error(err),
		)
		return fmt.Errorf("nats publish failed: %w", err)
	}

	return nil
}