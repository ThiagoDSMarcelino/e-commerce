package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	rmq "github.com/rabbitmq/rabbitmq-amqp-go-client/pkg/rabbitmqamqp"
)

const routingKey = "pedido.enviado"

var bindingKeys = []string{"pagamento.aprovado"}

func main() {
	if err := run(); err != nil {
		slog.Error("fatal", "error", err)
		os.Exit(1)
	}
}

func run() error {
	// Create a context that will be canceled on SIGINT (Ctrl+C) or SIGTERM.
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	settings, err := LoadSettings()
	if err != nil {
		return fmt.Errorf("Failed to load settings: %v", err)
	}

	SetupLogger(settings.LogLevel)

	broker, err := NewBroker(ctx, settings.BrokerURI, settings.ExchangeName, bindingKeys)
	if err != nil {
		return fmt.Errorf("Failed to create broker: %v", err)
	}
	defer func() {
		_ = broker.Close()
	}()

	return broker.Consume(ctx, func(ctx context.Context, delivery rmq.IDeliveryContext) error {
		message := delivery.Message()

		if len(message.Data) == 0 {
			return fmt.Errorf("Received message with no data")
		}

		if len(message.Data) > 1 {
			return fmt.Errorf("Received message with multiple data parts")
		}

		if len(message.Data[0]) == 0 {
			return fmt.Errorf("Received message with empty data")
		}

		order, err := ParseOrder(message.Data[0])
		if err != nil {
			return fmt.Errorf("Failed to parse order: %v", err)
		}

		slog.Info("Iniciando envio do pedido", "order", order)
		slog.Info("Nota fiscal emitida", "order", order)
		slog.Info("Entrega iniciada", "order", order)

		payload, err := order.Serialize()
		if err != nil {
			return fmt.Errorf("Failed to serialize order: %v", err)
		}

		err = broker.Publish(ctx, routingKey, payload)
		if err != nil {
			return fmt.Errorf("Failed to publish message: %v", err)
		}

		err = delivery.Accept(ctx)
		if err != nil {
			return fmt.Errorf("Failed to accept message: %v", err)
		}

		return nil
	})
}
