package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

const (
	routingKeyApproved = "pagamento.aprovado"
	routingKeyDeclined = "pagamento.recusado"
)

var bindingKeys = []string{"pedido.estoque_ok"}

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

	signer, err := NewSigner(settings)
	if err != nil {
		return fmt.Errorf("Failed to create signer: %v", err)
	}

	broker, err := NewBroker(ctx, settings, signer, bindingKeys)
	if err != nil {
		return fmt.Errorf("Failed to create broker: %v", err)
	}
	defer func() {
		_ = broker.Close()
	}()

	paymentProcessor := NewPaymentProcessor(broker)

	return broker.Consume(ctx, paymentProcessor.ProcessDelivery)
}
