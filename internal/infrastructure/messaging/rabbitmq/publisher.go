package rabbitmq

import (
    "github.com/streadway/amqp"
	"authentication/internal/application/contracts/messaging"
)

type RabbitMQBus struct {
    channel    *amqp.Channel
    connection *amqp.Connection
}

func NewRabbitMQPublisher(url string) (*RabbitMQBus, error) {
    conn, err := amqp.Dial(url)
    if err != nil {
        return nil, err
    }

    ch, err := conn.Channel()
    if err != nil {
        conn.Close()
        return nil, err
    }

    return &RabbitMQBus{
        channel:    ch,
        connection: conn,
    }, nil
}

func (p *RabbitMQBus) Publish(event messaging.Event) error {
    return p.channel.Publish(
        "",
        event.Name,
        false,
        false,
        amqp.Publishing{
            ContentType: "application/json",
            Body:        event.Payload,
        },
    )
}

func (r *RabbitMQBus) Close() error {
    var err error
    if r.channel != nil {
        if e := r.channel.Close(); e != nil {
            err = e
        }
    }
    if r.connection != nil {
        if e := r.connection.Close(); e != nil && err == nil {
            err = e
        }
    }
    return err
}
