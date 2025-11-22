package messaging

import(
	"fmt"
	"authentication/internal/infrastructure/messaging/rabbitmq"
	"authentication/internal/infrastructure/messaging/kafka"
	"authentication/internal/infrastructure/messaging/nats"
	"authentication/internal/application/contracts/messaging"

)

type BrokerType string

const(
	RabbitMQ BrokerType = "rabbitmq"
	Kafka    BrokerType = "kafka"
	NATS 	BrokerType = "nats"
)

func NewEventBus(broker BrokerType, url string) (messaging.EventBus, error) {
	switch broker {
	case RabbitMQ:
		return rabbitmq.NewRabbitMQPublisher(url)
	case Kafka:
		return kafka.NewKafkaBus(url)
	case NATS:
		return nats.NewNATSBus(url)
	default:
		return nil, fmt.Errorf("unsupported broker type: %s", broker)
	}
}