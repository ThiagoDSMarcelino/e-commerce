package main

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"os"
	"os/signal"
	"promotions/internal/adapters/rabbitmq"
	"promotions/internal/config"
	"promotions/internal/registrations"
	"promotions/internal/signing"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	rmq "github.com/rabbitmq/rabbitmq-amqp-go-client/pkg/rabbitmqamqp"
)

const (
	maxDelaySeconds = 10
	minDelaySeconds = 5

	minDiscount  = 5
	maxDiscount  = 30
	validityDays = 30
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

	if err := godotenv.Load(); err != nil {
		slog.Warn("no .env file found, using environment variables")
	}

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

	_, err = conn.Management().DeclareExchange(ctx, &rmq.TopicExchangeSpecification{Name: settings.Rmq.ExchangeName})
	if err != nil {
		return fmt.Errorf("Failed to declare an exchange: %v", err)
	}

	registrationConsumer, err := rabbitmq.NewConsumer[registrations.RegisterPromotionEvent](
		ctx,
		signer,
		conn,
		settings.Rmq.ExchangeName,
		routingKeyPromotionsRegister,
	)
	if err != nil {
		return err
	}

	unregistrationConsumer, err := rabbitmq.NewConsumer[registrations.UnregisterPromotionEvent](
		ctx,
		signer,
		conn,
		settings.Rmq.ExchangeName,
		routingKeyPromotionsUnregister,
	)
	if err != nil {
		return err
	}

	registrationService := registrations.NewService()

	go registrationConsumer.Run(ctx, registrationService.RegisterEmail)
	go unregistrationConsumer.Run(ctx, registrationService.UnregisterEmail)

	// resendClient := resend.NewClient(settings.Resend.ApiKey)

	for {
		registrationService.GetRegistredEmails()

		// TODO: send

		// params := &resend.SendEmailRequest{
		// 	From:    "Acme <onboarding@resend.dev>",
		// 	To:      []string{"delivered@resend.dev"},
		// 	Html:    "<strong>hello world</strong>",
		// 	Subject: "Hello from Golang",
		// 	Cc:      []string{"cc@example.com"},
		// 	Bcc:     []string{"bcc@example.com"},
		// 	ReplyTo: "replyto@example.com",
		// }

		// sent, err := client.Emails.Send(params)
		// if err != nil {
		// 	fmt.Println(err.Error())
		// 	return
		// }
		// fmt.Println(sent.Id)

		delay := rand.IntN(maxDelaySeconds-minDelaySeconds+1) + minDelaySeconds

		select {
		case <-ctx.Done():
			slog.Info("Shutting down")
			return nil
		case <-time.After(time.Duration(delay) * time.Second):
		}
	}
}
