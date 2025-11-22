package nats

import (
	"authentication/internal/application/contracts/messaging"

	"github.com/nats-io/nats.go"
)

type NATSBus struct {
	conn *nats.Conn
}

func NewNATSBus(url string) (*NATSBus, error) {
	conn, err := nats.Connect(url)
	if err != nil {
		return nil, err
	}
	return &NATSBus{conn: conn}, nil
}

func (n *NATSBus) Publish(event messaging.Event) error {
	return n.conn.Publish(event.Name, event.Payload)
}

func (n *NATSBus) Close() error {
	if n.conn != nil && !n.conn.IsClosed() {
		n.conn.Close()
	}
	return nil
}
