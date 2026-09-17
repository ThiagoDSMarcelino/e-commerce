package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	rmq "github.com/rabbitmq/rabbitmq-amqp-go-client/pkg/rabbitmqamqp"
)

var bindingKeys = []string{"promocao.categoria.*"}

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

	if err := godotenv.Load(); err != nil {
		slog.Warn("no .env file found, using environment variables")
	}

	settings, err := LoadSettings()
	if err != nil {
		return fmt.Errorf("Failed to load settings: %v", err)
	}

	SetupLogger(settings.LogLevel)

	env := rmq.NewEnvironment(settings.BrokerURI, nil)
	conn, err := env.NewConnection(ctx)
	if err != nil {
		return fmt.Errorf("Failed to connect to RabbitMQ: %v", err)
	}
	defer func() {
		_ = env.CloseConnections(context.Background())
	}()

	_, err = conn.Management().DeclareExchange(ctx, &rmq.TopicExchangeSpecification{Name: settings.ExchangeName})
	if err != nil {
		return fmt.Errorf("Failed to declare an exchange: %v", err)
	}

	_, err = conn.Management().DeclareQueue(ctx, &rmq.QuorumQueueSpecification{Name: settings.QueueName})
	if err != nil {
		return fmt.Errorf("Failed to declare a queue: %v", err)
	}

	for _, bk := range bindingKeys {
		_, err = conn.Management().Bind(ctx, &rmq.ExchangeToQueueBindingSpecification{
			SourceExchange:   settings.ExchangeName,
			DestinationQueue: settings.QueueName,
			BindingKey:       bk,
		})
		if err != nil {
			return fmt.Errorf("Failed to bind a queue: %v", err)
		}
	}

	consumer, err := conn.NewConsumer(ctx, settings.QueueName, &rmq.ConsumerOptions{
		InitialCredits: 1, // One message at a time
	})
	if err != nil {
		return fmt.Errorf("Failed to create consumer: %v", err)
	}
	defer func() { _ = consumer.Close(context.Background()) }()

	for {
		delivery, err := consumer.Receive(ctx)
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

		data := message.Data[0]

		promotion, err := ParsePromotion(data)
		if err != nil {
			slog.Error("Failed to parse promotion", "error", err)
			_ = delivery.Discard(ctx, nil)
			continue
		}

		println("Client B received promotion category for", promotion.Category, "| discount:", promotion.Discount, "| valid until:", promotion.ValidUntil)

		err = delivery.Accept(ctx)
		if err != nil {
			slog.Error("Failed to handle delivery", "error", err)
		}
	}
}
