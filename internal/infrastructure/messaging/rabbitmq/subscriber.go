package rabbitmq

import (
	"authentication/internal/application/contracts/messaging"
)

func (r *RabbitMQBus) Subscribe(queue string, handler func(messaging.Event)) error {
    msgs, err := r.channel.Consume(
        queue, "", true, false, false, false, nil,
    )
    if err != nil {
        return err
    }

    go func() {
        for msg := range msgs {
            handler(messaging.Event{
                Name:    queue,
                Payload: msg.Body,
            })
        }
    }()
    return nil
}
