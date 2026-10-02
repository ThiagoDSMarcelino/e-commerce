package config

import (
	"errors"
	"fmt"
	"log/slog"
	"main/internal/env"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Addr         string
	BrokerURI    string
	ExchangeName string
	LogLevel     slog.Level
	ServiceName  string
	KeysDir      string
	QueueName    string
	Db           DatabaseConfig
}

type DatabaseConfig struct {
	Dsn string
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		slog.Warn("no .env file found, using environment variables")
	}

	port := env.GetString("PORT", "8080")

	brokerURI := os.Getenv("BROKER_URI")
	if brokerURI == "" {
		return nil, errors.New("BROKER_URI is not set or is empty")
	}

	exchangeName := os.Getenv("EXCHANGE_NAME")
	if exchangeName == "" {
		return nil, errors.New("EXCHANGE_NAME is not set or is empty")
	}

	logLevel, err := parseLogLevel(os.Getenv("LOG_LEVEL"))
	if err != nil {
		return nil, err
	}

	serviceName := os.Getenv("SERVICE_NAME")
	if serviceName == "" {
		return nil, errors.New("SERVICE_NAME is not set or is empty")
	}

	keysDir := os.Getenv("KEYS_DIR")
	if keysDir == "" {
		return nil, errors.New("KEYS_DIR is not set or is empty")
	}

	queueName := os.Getenv("QUEUE_NAME")
	if queueName == "" {
		return nil, errors.New("QUEUE_NAME is not set or is empty")
	}

	dsn := os.Getenv("GOOSE_DBSTRING")
	if queueName == "" {
		return nil, errors.New("CONNECTION_STRING is not set or is empty")
	}

	return &Config{
		Addr:         fmt.Sprintf(":%s", port),
		BrokerURI:    brokerURI,
		ExchangeName: exchangeName,
		LogLevel:     logLevel,
		ServiceName:  serviceName,
		KeysDir:      keysDir,
		QueueName:    queueName,
		Db: DatabaseConfig{
			Dsn: dsn,
		},
	}, nil
}

func parseLogLevel(value string) (slog.Level, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "":
		return slog.LevelInfo, nil
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return 0, fmt.Errorf("LOG_LEVEL %q is not one of debug, info, warn, error", value)
	}
}

func (cfg *Config) SetLoggerLevel() {
	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: cfg.LogLevel})
	slog.SetDefault(slog.New(handler))
}
