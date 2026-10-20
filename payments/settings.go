package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Settings struct {
	BrokerURI       string
	ExchangeName    string
	LogLevel        slog.Level
	ServiceName     string
	KeysDir         string
	QueueName       string
	ProcessingDelay time.Duration
	PaymentMockURL  string
	WebhookURL      string
	Port            string
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

	processingDelay, err := parseDelaySeconds(os.Getenv("PROCESSING_DELAY_SECONDS"))
	if err != nil {
		return nil, err
	}

	paymentMockURL := os.Getenv("PAYMENT_MOCK_URL")
	if paymentMockURL == "" {
		return nil, errors.New("PAYMENT_MOCK_URL is not set or is empty")
	}

	webhookURL := os.Getenv("WEBHOOK_URL")
	if webhookURL == "" {
		return nil, errors.New("WEBHOOK_URL is not set or is empty")
	}

	port := os.Getenv("PORT")
	if port == "" {
		return nil, errors.New("PORT is not set or is empty")
	}

	return &Settings{
		BrokerURI:       brokerURI,
		ExchangeName:    exchangeName,
		LogLevel:        logLevel,
		ServiceName:     serviceName,
		KeysDir:         keysDir,
		QueueName:       queueName,
		ProcessingDelay: processingDelay,
		PaymentMockURL:  strings.TrimRight(paymentMockURL, "/"),
		WebhookURL:      webhookURL,
		Port:            port,
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

func parseDelaySeconds(value string) (time.Duration, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, nil
	}

	seconds, err := strconv.Atoi(value)
	if err != nil || seconds < 0 {
		return 0, fmt.Errorf("PROCESSING_DELAY_SECONDS %q is not a non-negative integer", value)
	}

	return time.Duration(seconds) * time.Second, nil
}

func SetupLogger(level slog.Level) {
	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})
	slog.SetDefault(slog.New(handler))
}
