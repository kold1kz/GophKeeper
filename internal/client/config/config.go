package config

import (
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	ServerAddress string
	StatePath     string
}

func Load() (*Config, error) {
	serverAddress := os.Getenv("GOPHKEEPER_SERVER_ADDRESS")
	if serverAddress == "" {
		serverAddress = "localhost:3200"
	}

	wd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("get working dir: %w", err)
	}

	statePath := filepath.Join(wd, ".gophkeeper", "state.json")

	return &Config{
		ServerAddress: serverAddress,
		StatePath:     statePath,
	}, nil
}
