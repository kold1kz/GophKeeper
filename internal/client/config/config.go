// Package config содержит конфигурацию CLI-клиента GophKeeper.
//
// Пакет отвечает за загрузку параметров подключения к серверу
// и путей к локальному состоянию клиента.
package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// Config описывает конфигурацию CLI-клиента.
//
// ServerAddress содержит адрес gRPC-сервера.
// StatePath содержит путь к локальному файлу состояния клиента.
type Config struct {
	ServerAddress string
	StatePath     string
}

// Load загружает конфигурацию клиента.
//
// Значения берутся из переменных окружения и значений по умолчанию.
// Если GOPHKEEPER_SERVER_ADDRESS не задан, используется localhost:3200.
//
// Также функция вычисляет путь к локальному файлу состояния клиента.
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
