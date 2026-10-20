package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	routingKeyApproved = "pagamento.aprovado"
	routingKeyDeclined = "pagamento.recusado"
	routingKeyPending  = "pagamento.pendente"
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

	paymentProcessor := NewPaymentProcessor(broker, NewChargeClient(settings), settings.ProcessingDelay)

	listener, err := net.Listen("tcp", ":"+settings.Port)
	if err != nil {
		return fmt.Errorf("Failed to listen on port %s: %v", settings.Port, err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /webhook", paymentProcessor.HandleWebhook)

	server := &http.Server{Handler: mux, ReadHeaderTimeout: 10 * time.Second}
	go func() {
		if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Webhook server stopped", "error", err)
		}
	}()
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()

	slog.Info("Webhook ouvindo", "port", settings.Port)

	return broker.Consume(ctx, paymentProcessor.ProcessDelivery)
}
