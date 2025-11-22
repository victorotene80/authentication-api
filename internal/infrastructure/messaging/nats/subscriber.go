package nats

import (
    "github.com/nats-io/nats.go"
	"authentication/internal/application/contracts/messaging"
)

func (n *NATSBus) Subscribe(subject string, handler func(messaging.Event)) error {
    _, err := n.conn.Subscribe(subject, func(msg *nats.Msg) {
        handler(messaging.Event{
            Name:    subject,
            Payload: msg.Data,
        })
    })
    return err
}
