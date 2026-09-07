package main

import (
	"errors"
	"os"
)

type Settings struct {
	BrokerURI    string
	ExchangeName string
}

func LoadSettings() (*Settings, error) {
	brokerURI := os.Getenv("BROKER_URI")
	if brokerURI == "" {
		return nil, errors.New("BROKER_URI is not set or is empty")
	}

	exchangeName := os.Getenv("EXCHANGE_NAME")
	if exchangeName == "" {
		return nil, errors.New("EXCHANGE_NAME is not set or is empty")
	}

	return &Settings{
		BrokerURI:    brokerURI,
		ExchangeName: exchangeName,
	}, nil
}
