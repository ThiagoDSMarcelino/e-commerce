package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"strings"

	rmq "github.com/rabbitmq/rabbitmq-amqp-go-client/pkg/rabbitmqamqp"
)

type Broker struct {
	env          *rmq.Environment
	publisher    *rmq.Publisher
	consumer     *rmq.Consumer
	exchangeName string
	signer       *Signer
	serviceName  string
}

type MessageResponse int

const (
	Accepted MessageResponse = iota
	Requeued
	Rejected
)

// ErrUnroutable reports that the broker accepted the transfer but no queue was
// bound to the routing key, so the message was dropped. Expected for events
// whose consumer does not exist yet.
var ErrUnroutable = errors.New("Message was not routed to any queue")

func NewBroker(ctx context.Context, settings *Settings, signer *Signer, bindingKeys []string) (*Broker, error) {
	env := rmq.NewEnvironment(settings.BrokerURI, nil)
	conn, err := env.NewConnection(ctx)
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to RabbitMQ: %v", err)
	}

	_, err = conn.Management().DeclareExchange(ctx, &rmq.DirectExchangeSpecification{Name: settings.ExchangeName})
	if err != nil {
		return nil, fmt.Errorf("Failed to declare an exchange: %v", err)
	}

	publisher, err := conn.NewPublisher(ctx, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("Failed to create publisher: %v", err)
	}

	_, err = conn.Management().DeclareQueue(ctx, &rmq.QuorumQueueSpecification{Name: settings.QueueName})
	if err != nil {
		return nil, fmt.Errorf("Failed to declare a queue: %v", err)
	}

	for _, bk := range bindingKeys {
		_, err = conn.Management().Bind(ctx, &rmq.ExchangeToQueueBindingSpecification{
			SourceExchange:   settings.ExchangeName,
			DestinationQueue: settings.QueueName,
			BindingKey:       bk,
		})
		if err != nil {
			return nil, fmt.Errorf("Failed to bind a queue: %v", err)
		}
	}

	consumer, err := conn.NewConsumer(ctx, settings.QueueName, &rmq.ConsumerOptions{
		InitialCredits: 1, // One message at a time
	})
	if err != nil {
		return nil, fmt.Errorf("Failed to create consumer: %v", err)
	}

	return &Broker{
		env:          env,
		publisher:    publisher,
		consumer:     consumer,
		signer:       signer,
		serviceName:  settings.ServiceName,
		exchangeName: settings.ExchangeName,
	}, nil
}

func (b *Broker) Close() error {
	return errors.Join(
		b.env.CloseConnections(context.Background()),
		b.publisher.Close(context.Background()),
		b.consumer.Close(context.Background()),
	)
}

func (b *Broker) Publish(ctx context.Context, routingKey string, message []byte) error {
	event, err := rmq.NewMessageWithAddress(message, &rmq.ExchangeAddress{
		Exchange: b.exchangeName,
		Key:      routingKey,
	})
	if err != nil {
		return fmt.Errorf("Failed to create message: %v", err)
	}

	sig, err := b.signer.Sign(message)
	if err != nil {
		return fmt.Errorf("failed to sign message: %w", err)
	}

	event.ApplicationProperties = map[string]any{
		"from":      b.serviceName,
		"signature": sig,
	}

	res, err := b.publisher.Publish(ctx, event)
	if err != nil {
		return fmt.Errorf("Failed to publish message: %v", err)
	}

	switch res.Outcome.(type) {
	case *rmq.StateAccepted:
		slog.Debug("Message was accepted")
		return nil
	case *rmq.StateReleased:
		return fmt.Errorf("%w: %v", ErrUnroutable, res.Outcome)
	case *rmq.StateRejected:
		return fmt.Errorf("Message was rejected: %v", res.Outcome)
	case *rmq.StateModified:
		return fmt.Errorf("Message was modified: %v", res.Outcome)
	default:
		return fmt.Errorf("Unexpected publish outcome: %v", res.Outcome)
	}
}

// routingKeyFromAddress extracts the routing key from the AMQP 1.0 "to" address
// that rmq.ExchangeAddress produces, e.g. "/exchanges/e_commerce/pedido.criado".
// The client percent-encodes both segments, so neither the exchange name nor the
// key can hold a literal '/': everything after the third '/' is the key.
func routingKeyFromAddress(address string) string {
	const prefix = "/exchanges/"

	if !strings.HasPrefix(address, prefix) {
		return "" // a queue address, or something else entirely
	}

	rest := address[len(prefix):] // "<exchange>/<key>"

	separator := strings.Index(rest, "/")
	if separator < 0 {
		return "" // published to the exchange without a routing key
	}

	encodedKey := rest[separator+1:]
	if encodedKey == "" {
		return ""
	}

	// PathUnescape, not QueryUnescape: the encoder emits %20 for a space and %2B
	// for a literal '+', so a '+' must not be decoded back into a space.
	key, err := url.PathUnescape(encodedKey)
	if err != nil {
		slog.Warn("Failed to decode routing key", "address", address, "error", err)
		return encodedKey // best effort: unescaped keys are already literal
	}

	return key
}

// routingKeyOf resolves the routing key of a received message. Every publisher in
// this system is an AMQP 1.0 client addressing the exchange through Properties.To;
// RabbitMQ fills Properties.Subject with the routing key only for messages that
// came from an AMQP 0.9.1 publisher, so Subject is the fallback.
func routingKeyOf(to, subject *string) string {
	if to != nil {
		if key := routingKeyFromAddress(*to); key != "" {
			return key
		}
	}

	if subject != nil {
		return *subject
	}

	return ""
}

type Event struct {
	RoutingKey string
	Data       []byte
}

func (b *Broker) Consume(ctx context.Context, handler func(ctx context.Context, event Event) MessageResponse) error {
	for {
		delivery, err := b.consumer.Receive(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				slog.Info("Shutting down gracefully...")
				return nil
			}
			return fmt.Errorf("Failed to receive a message: %v", err)
		}

		message := delivery.Message()

		if len(message.Data) == 0 {
			slog.Error("Received message with no data")
			_ = delivery.Discard(ctx, nil)
			continue
		}

		if len(message.Data) > 1 {
			slog.Error("Received message with multiple data parts")
			_ = delivery.Discard(ctx, nil)
			continue
		}

		if len(message.Data[0]) == 0 {
			slog.Error("Received message with empty data")
			_ = delivery.Discard(ctx, nil)
			continue
		}

		from, ok := message.ApplicationProperties["from"].(string)
		if !ok {
			slog.Error("Received message with no 'from' property")
			_ = delivery.Discard(ctx, nil)
			continue
		}

		sig, ok := message.ApplicationProperties["signature"].(string)
		if !ok {
			slog.Error("Received message with no signature")
			_ = delivery.Discard(ctx, nil)
			continue
		}

		data := message.Data[0]

		err = b.signer.Verify(from, sig, data)
		if err != nil {
			slog.Error("Received message with invalid signature", "error", err)
			_ = delivery.Discard(ctx, nil)
			continue
		}

		var to, subject *string
		if message.Properties != nil {
			to = message.Properties.To
			subject = message.Properties.Subject
		}

		var event = Event{
			RoutingKey: routingKeyOf(to, subject),
			Data:       data,
		}

		slog.Debug("Received message", "from", from, "routingKey", event.RoutingKey)

		response := handler(ctx, event)

		switch response {
		case Accepted:
			err = delivery.Accept(ctx)
		case Requeued:
			err = delivery.Requeue(ctx)
		case Rejected:
			err = delivery.Discard(ctx, nil)
		default:
			err = fmt.Errorf("Unknown response type: %v", response)
		}

		if err != nil {
			slog.Error("Failed to handle delivery", "error", err)
		}
	}
}
