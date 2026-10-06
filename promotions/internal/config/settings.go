package config

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type RabbitSettings struct {
	BrokerURI    string
	ExchangeName string
	QueueName    string
}

type ResendSettings struct {
	ApiKey string
}

type LogSettings struct {
	Level slog.Level
}

type SigningSettings struct {
	ServiceName string
	KeysDir     string
}

type Settings struct {
	Rmq     RabbitSettings
	Resend  ResendSettings
	Log     LogSettings
	Signing SigningSettings
}

func LoadSettings() (*Settings, error) {
	if err := godotenv.Load(); err != nil {
		slog.Warn("no .env file found, using environment variables")
	}

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

	resendApiKey := os.Getenv("RESEND_API_KEY")
	if queueName == "" {
		return nil, errors.New("RESEND_API_KEY is not set or is empty")
	}

	return &Settings{
		Rmq: RabbitSettings{
			BrokerURI:    brokerURI,
			ExchangeName: exchangeName,
			QueueName:    queueName,
		},
		Log: LogSettings{
			Level: logLevel,
		},
		Signing: SigningSettings{
			ServiceName: serviceName,
			KeysDir:     keysDir,
		},
		Resend: ResendSettings{
			resendApiKey,
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

func SetupLogger(level slog.Level) {
	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})
	slog.SetDefault(slog.New(handler))
}
