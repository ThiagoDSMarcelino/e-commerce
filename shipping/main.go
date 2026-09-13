package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
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

	return broker.Consume(ctx, func(ctx context.Context, data []byte) MessageResponse {
		order, err := ParseOrder(data)
		if err != nil {
			return Rejected
		}

		slog.Info("Iniciando envio do pedido", "order", order)
		slog.Info("Nota fiscal emitida", "order", order)
		slog.Info("Entrega iniciada", "order", order)

		payload, err := order.Serialize()
		if err != nil {
			return Rejected
		}

		err = broker.Publish(ctx, routingKey, payload)
		if err != nil {
			return Rejected
		}

		return Accepted
	})
}
