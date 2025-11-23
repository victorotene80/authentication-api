package nats

import (
	"authentication/internal/application/contracts/messaging"
	"context"
	"fmt"

	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

func (n *NATSBus) Subscribe(subject string, handler func(messaging.Event)) error {
	ctx := context.Background()

	_, err := n.conn.Subscribe(subject, func(msg *nats.Msg) {
		handler(messaging.Event{
			Name:    subject,
			Payload: msg.Data,
		})
	})

	if err != nil {
		n.logger.Error(ctx, "failed to subscribe to NATS subject",
			zap.String("subject", subject),
			zap.Error(err),
		)
		return fmt.Errorf("nats subscribe failed: %w", err)
	}

	n.logger.Info(ctx, "subscribed to NATS subject", zap.String("subject", subject))
	return nil
}

func (n *NATSBus) Close() error {
	ctx := context.Background()

	if n.conn != nil && !n.conn.IsClosed() {
		n.conn.Close()
		n.logger.Info(ctx, "NATS connection closed")
	}
	return nil
}