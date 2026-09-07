package main

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	rmq "github.com/rabbitmq/rabbitmq-amqp-go-client/pkg/rabbitmqamqp"
)

const (
	maxDelaySeconds = 5
	minDelaySeconds = 1
)

const routingKeyPromotions = "promocao.categoria"

var categories = []string{"A", "B", "C"}

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

	_, err = conn.Management().DeclareExchange(ctx, &rmq.DirectExchangeSpecification{Name: settings.ExchangeName})
	if err != nil {
		return fmt.Errorf("Failed to declare an exchange: %v", err)
	}

	publisher, err := conn.NewPublisher(ctx, nil, nil)
	if err != nil {
		return fmt.Errorf("Failed to create publisher: %v", err)
	}
	defer func() { _ = publisher.Close(context.Background()) }()

	for {
		category := categories[rand.IntN(len(categories))]

		slog.Info("Creating promotion", "category", category)

		routingKey := routingKeyPromotions + "." + category

		outcomeMsg, err := rmq.NewMessageWithAddress([]byte(`{}`), &rmq.ExchangeAddress{
			Exchange: settings.ExchangeName,
			Key:      routingKey,
		})
		if err != nil {
			slog.Error("Failed to create message", "error", err)
			continue
		}

		res, err := publisher.Publish(ctx, outcomeMsg)
		if err != nil {
			slog.Error("Failed to publish", "error", err)
			continue
		}
		switch res.Outcome.(type) {
		case *rmq.StateAccepted:
		case *rmq.StateRejected:
			slog.Error("Message was rejected", "outcome", res.Outcome)
			continue
		case *rmq.StateReleased:
			slog.Info("Message was released", "outcome", res.Outcome)
		case *rmq.StateModified:
			slog.Error("Message was modified", "outcome", res.Outcome)
			continue
		default:
			slog.Error("Unexpected publish outcome", "outcome", res.Outcome)
			continue
		}

		delay := rand.IntN(maxDelaySeconds-minDelaySeconds+1) + minDelaySeconds
		slog.Debug("Promotion sent", "category", category, "delay", delay)

		select {
		case <-ctx.Done():
			slog.Info("Shutting down")
			return nil
		case <-time.After(time.Duration(delay) * time.Second):
		}
	}
}
