package messaging

type Event struct {
    Name    string
    Payload []byte
}

type Publisher interface {
    Publish(event Event) error
    Close() error
}

type Subscriber interface {
    Subscribe(topic string, handler func(Event)) error
    Close() error
}

type EventBus interface {
    Publisher
    Subscriber
}
