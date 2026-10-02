package main

import (
	"context"
	"log/slog"
	"main/internal/config"
	"os"

	"github.com/jackc/pgx/v5"
)

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

	api := NewApplication(cfg, conn)

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	cfg.SetLoggerLevel()

	if err = api.Run(api.Mount()); err != nil {
		slog.Error("Server failed to start", "error", err)
		os.Exit(1)
	}
}
