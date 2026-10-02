package main

import (
	"context"
	"log/slog"
	"main/internal/adapters/rabbitmq"
	"main/internal/config"
	"main/internal/signing"
	"os"

	"github.com/jackc/pgx/v5"
)

var bindingKeys = []string{
	"pagamento.aprovado",
	"pagamento.recusado",
	"pedido.enviado",
	"pedido.estoque_ok",
	"estoque.indisponivel",
}

func main() {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("Server failed to load config", "error", err)
		os.Exit(1)
	}

	conn, err := pgx.Connect(ctx, cfg.Db.Dsn)
	if err != nil {
		slog.Error("Server failed to connecto to database", "error", err)
		os.Exit(1)
	}
	defer conn.Close(ctx)

	signer, err := signing.NewSigner(cfg)
	if err != nil {
		slog.Error("Failed to create signer", "error", err)
		os.Exit(1)
	}

	broker, err := rabbitmq.NewBroker(ctx, cfg, signer, bindingKeys)
	if err != nil {
		slog.Error("Failed to create broker", "error", err)
		os.Exit(1)
	}
	defer broker.Close()

	api := NewApplication(cfg, conn, broker)

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	cfg.SetLoggerLevel()

	if err = api.Run(api.Mount(ctx)); err != nil {
		slog.Error("Server failed to start", "error", err)
		os.Exit(1)
	}
}
