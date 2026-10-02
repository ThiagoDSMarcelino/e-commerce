package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"main/internal/config"
	"main/internal/order"
	"main/internal/platform/rabbitmq"
	"main/internal/signing"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	routingKeyCreated = "pedido.criado"
	routingKeyDeleted = "pedido.excluido"
)

var bindingKeys = []string{
	"pagamento.aprovado",
	"pagamento.recusado",
	"pedido.enviado",
	"pedido.estoque_ok",
	"estoque.indisponivel",
}

type readFunc func() (line string, quit bool, err error)

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

	settings, err := config.Load()
	if err != nil {
		return fmt.Errorf("Failed to load settings: %v", err)
	}

	config.SetupLogger(settings.LogLevel)

	signer, err := signing.NewSigner(settings)
	if err != nil {
		return fmt.Errorf("Failed to create signer: %v", err)
	}

	broker, err := rabbitmq.NewBroker(ctx, settings, signer, bindingKeys)
	if err != nil {
		return fmt.Errorf("Failed to create broker: %v", err)
	}
	defer func() {
		_ = broker.Close()
	}()

	ordersRepo := order.NewOrdersRepository()
	ordersHandler := order.NewOrderHandler(ordersRepo, broker)

	consumerDone := make(chan struct{})
	go func() {
		defer close(consumerDone)

		err := broker.Consume(ctx, func(ctx context.Context, event rabbitmq.Event) rabbitmq.MessageResponse {
			return ordersHandler.HandleStatusEvent(ctx, event)
		})
		if err != nil {
			slog.Error("consumer error", "error", err)
		}
	}()

	router := gin.Default()
	router.GET("/orders", ordersHandler.GetOrders)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	srvErrs := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			srvErrs <- err
			return
		}
		srvErrs <- nil
	}()

	slog.Info("HTTP server listening", "addr", srv.Addr)

	select {
	case <-ctx.Done():
		slog.Info("Signal received, shutting down...")
	case err := <-srvErrs:
		if err != nil {
			return fmt.Errorf("HTTP server failed: %v", err)
		}
	}

	shutdownCtx, stopShutdown := context.WithTimeout(context.Background(), 8*time.Second)
	defer stopShutdown()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("HTTP shutdown failed", "error", err)
	}

	select {
	case <-consumerDone:
	case <-time.After(5 * time.Second):
		slog.Warn("consumer did not stop in time")
	}

	return nil
}
