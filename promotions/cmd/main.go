package main

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"os"
	"os/signal"
	"promotions/internal/adapters/rabbitmq"
	"promotions/internal/catalog"
	"promotions/internal/config"
	"promotions/internal/email"
	"promotions/internal/promotions"
	"promotions/internal/registrations"
	"promotions/internal/signing"
	"syscall"
	"time"

	rmq "github.com/rabbitmq/rabbitmq-amqp-go-client/pkg/rabbitmqamqp"
)

const (
	maxDelaySeconds = 10
	minDelaySeconds = 5
)

const (
	routingKeyPromotionsRegister   = "interesse.promocao.register"
	routingKeyPromotionsUnregister = "interesse.promocao.unregister"
)

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

	settings, err := config.LoadSettings()
	if err != nil {
		return fmt.Errorf("Failed to load settings: %v", err)
	}

	config.SetupLogger(settings.Log.Level)

	signer, err := signing.NewSigner(settings)
	if err != nil {
		return fmt.Errorf("Failed to create signer: %v", err)
	}

	env := rmq.NewEnvironment(settings.Rmq.BrokerURI, nil)
	conn, err := env.NewConnection(ctx)
	if err != nil {
		return fmt.Errorf("Failed to connect to RabbitMQ: %v", err)
	}
	defer func() {
		_ = env.CloseConnections(context.Background())
	}()

	_, err = conn.Management().DeclareExchange(ctx, &rmq.DirectExchangeSpecification{Name: settings.Rmq.ExchangeName})
	if err != nil {
		return fmt.Errorf("Failed to declare an exchange: %v", err)
	}

	consumer, err := rabbitmq.NewConsumer(
		ctx,
		signer,
		conn,
		settings.Rmq.ExchangeName,
		settings.Rmq.QueueName,
		[]string{routingKeyPromotionsRegister, routingKeyPromotionsUnregister},
	)
	if err != nil {
		return err
	}

	registrationService := registrations.NewService()

	go consumer.Run(ctx, map[string]rabbitmq.Handler{
		routingKeyPromotionsRegister:   rabbitmq.JSON(registrationService.RegisterEmail),
		routingKeyPromotionsUnregister: rabbitmq.JSON(registrationService.UnregisterEmail),
	})

	catalogClient := catalog.NewClient(settings.Storage.URL)
	sender := email.NewSender(settings.Resend.ApiKey, settings.Resend.From)

	for {
		sendPromotion(ctx, catalogClient, registrationService, sender)

		delay := rand.IntN(maxDelaySeconds-minDelaySeconds+1) + minDelaySeconds

		select {
		case <-ctx.Done():
			slog.Info("Shutting down")
			return nil
		case <-time.After(time.Duration(delay) * time.Second):
		}
	}
}

func sendPromotion(ctx context.Context, catalogClient *catalog.Client, registrationService *registrations.Service, sender *email.Sender) {
	products, err := catalogClient.Products(ctx)
	if err != nil {
		slog.Error("Failed to fetch products", "error", err)
		return
	}

	if len(products) == 0 {
		slog.Debug("No products in stock, skipping promotion")
		return
	}

	promo := promotions.Generate(products)

	recipients := registrationService.Recipients(promo.Category)
	if len(recipients) == 0 {
		slog.Debug("No subscribers for category, skipping promotion", "category", promo.Category, "productId", promo.ProductID)
		return
	}

	msg, err := email.Render(promo)
	if err != nil {
		slog.Error("Failed to render promotion email", "error", err)
		return
	}

	slog.Info("Sending promotion",
		"promotionId", promo.ID,
		"productId", promo.ProductID,
		"category", promo.Category,
		"discount", promo.Discount,
		"recipients", len(recipients),
	)

	if err := sender.Send(ctx, msg, promo.ID, promo.Category, recipients); err != nil {
		slog.Error("Failed to send promotion", "promotionId", promo.ID, "error", err)
	}
}
