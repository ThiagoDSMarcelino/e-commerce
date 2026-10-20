package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	if err := run(); err != nil {
		slog.Error("fatal", "error", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	settings, err := LoadSettings()
	if err != nil {
		return fmt.Errorf("Failed to load settings: %v", err)
	}

	handlers := NewHandlers(settings, NewChargeStore())

	mux := http.NewServeMux()
	mux.HandleFunc("POST /charges", handlers.CreateCharge)
	mux.HandleFunc("GET /checkout/{orderId}", handlers.ShowCheckout)
	mux.HandleFunc("POST /checkout/{orderId}", handlers.Pay)

	server := &http.Server{
		Addr:              ":" + settings.Port,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()

	slog.Info("Mock de pagamento ouvindo", "port", settings.Port, "publicUrl", settings.PublicURL)

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("Failed to serve: %v", err)
	}

	slog.Info("Shutting down gracefully...")

	return nil
}
