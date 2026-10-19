package main

import (
	"context"
	"log/slog"
	"time"
)

func waitDelay(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return nil
	}

	slog.Info("Aguardando antes de processar", "delay", d)

	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
