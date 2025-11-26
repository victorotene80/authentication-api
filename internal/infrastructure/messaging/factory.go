package messaging

import (
	"authentication/shared/logging"
	"authentication/internal/application/contracts/messaging"
	"authentication/internal/infrastructure/messaging/brokers/kafka"
	"authentication/internal/infrastructure/messaging/brokers/nats"
	"authentication/internal/infrastructure/messaging/brokers/rabbitmq"
	"fmt"
)

type BrokerType string

const (
	RabbitMQ BrokerType = "rabbitmq"
	Kafka    BrokerType = "kafka"
	NATS     BrokerType = "nats"
)

func NewEventBus(broker BrokerType, url string, logger logging.Logger) (messaging.EventBus, error) {
	switch broker {
	case RabbitMQ:
		return rabbitmq.NewRabbitMQBus(url, logger)
	case Kafka:
		return kafka.NewKafkaBus(url, logger)
	case NATS:
		return nats.NewNATSBus(url, logger)
	default:
		return nil, fmt.Errorf("unsupported broker type: %s", broker)
	}
}