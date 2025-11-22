package kafka

import (
    "context"
    "github.com/segmentio/kafka-go"
	"authentication/internal/application/contracts/messaging"
)

type KafkaBus struct {
    writer *kafka.Writer
    reader *kafka.Reader
    brokers []string
}

func NewKafkaBus(brokerURL string) (*KafkaBus, error) {
    brokers := []string{brokerURL}
    return &KafkaBus{
        brokers: brokers,
    }, nil
}

func (k *KafkaBus) Publish(event messaging.Event) error {
    writer := &kafka.Writer{
        Addr:     kafka.TCP(k.brokers...),
        Topic:    event.Name,
        Balancer: &kafka.LeastBytes{},
    }
    defer writer.Close()

    return writer.WriteMessages(context.Background(),
        kafka.Message{Value: event.Payload},
    )
}

func (k *KafkaBus) Close() error {
    if k.writer != nil {
        return k.writer.Close()
    }
    if k.reader != nil {
        return k.reader.Close()
    }
    return nil
}
