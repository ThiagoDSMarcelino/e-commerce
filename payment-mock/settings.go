package main

import (
	"errors"
	"os"
	"strings"
)

type Settings struct {
	Port      string
	PublicURL string
}

func LoadSettings() (*Settings, error) {
	port := os.Getenv("PORT")
	if port == "" {
		return nil, errors.New("PORT is not set or is empty")
	}

	publicURL := os.Getenv("PUBLIC_URL")
	if publicURL == "" {
		return nil, errors.New("PUBLIC_URL is not set or is empty")
	}

	return &Settings{
		Port:      port,
		PublicURL: strings.TrimRight(publicURL, "/"),
	}, nil
}
