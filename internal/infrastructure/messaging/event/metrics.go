// internal/infrastructure/messaging/event/metrics.go
package event

import (
	"authentication/internal/application/contracts/messaging"
	domainEvents "authentication/internal/domain/events"
	"authentication/internal/infrastructure/observability/metrics"
	"authentication/shared/utils"
	"context"
	"time"
)

type MetricsEventHandler struct {
	handler messaging.EventHandler
}

func NewMetricsEventHandler(handler messaging.EventHandler) messaging.EventHandler {
	return &MetricsEventHandler{handler: handler}
}

func (m *MetricsEventHandler) Handle(ctx context.Context, event domainEvents.DomainEvent) error {
	eventName := event.EventName()
	start := utils.NowUTC()

	err := m.handler.Handle(ctx, event)

	duration := time.Since(start).Seconds()
	status := "success"
	if err != nil {
		status = "error"
	}

	metrics.EventsProcessed.WithLabelValues(eventName, status).Inc()
	metrics.EventProcessingDuration.WithLabelValues(eventName).Observe(duration)

	return err
}

func (m *MetricsEventHandler) CanHandle(eventName string) bool {
	return m.handler.CanHandle(eventName)
}

type InstrumentedEventBus struct {
	bus        messaging.EventBus
	serializer messaging.EventSerializer
}

func NewInstrumentedEventBus(bus messaging.EventBus, serializer messaging.EventSerializer) *InstrumentedEventBus {
	return &InstrumentedEventBus{
		bus:        bus,
		serializer: serializer,
	}
}

func (i *InstrumentedEventBus) Publish(event messaging.Event) error {
	eventName := event.Name
	start := utils.NowUTC()

	err := i.bus.Publish(event)

	duration := time.Since(start).Seconds()
	status := "success"
	if err != nil {
		status = "error"
	}

	metrics.EventsPublished.WithLabelValues(eventName, status).Inc()
	metrics.MessagePublishDuration.WithLabelValues(eventName).Observe(duration)

	return err
}

func (i *InstrumentedEventBus) Subscribe(topic string, handler func(messaging.Event)) error {
	wrappedHandler := func(e messaging.Event) {
		start := utils.NowUTC()

		handler(e)

		duration := time.Since(start).Seconds()
		metrics.EventProcessingDuration.WithLabelValues(e.Name).Observe(duration)
		metrics.EventsProcessed.WithLabelValues(e.Name, "success").Inc()
	}

	return i.bus.Subscribe(topic, wrappedHandler)
}

func (i *InstrumentedEventBus) Close() error {
	return i.bus.Close()
}

// PublishDomainEvent publishes a domain event with metrics
/*func (i *InstrumentedEventBus) PublishDomainEvent(ctx context.Context, event events.DomainEvent) error {
	eventName := event.EventName()
	start := utils.NowUTC()

	payload, err := i.serializer.Serialize(event)
	if err != nil {
		metrics.EventsPublished.WithLabelValues(eventName, "serialization_error").Inc()
		return err
	}

	busEvent := Event{
		Name:    eventName,
		Payload: payload,
	}

	err = i.bus.Publish(busEvent)

	duration := time.Since(start).Seconds()
	status := "success"
	if err != nil {
		status = "error"
	}

	metrics.EventsPublished.WithLabelValues(eventName, status).Inc()
	metrics.MessagePublishDuration.WithLabelValues(eventName).Observe(duration)

	return err
}*/
