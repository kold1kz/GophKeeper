// Package config содержит конфигурацию серверного приложения GophKeeper.
//
// Пакет отвечает за:
//   - чтение конфигурации из флагов командной строки;
//   - применение переменных окружения;
//   - базовую валидацию параметров;
//   - инициализацию подключения к базе данных.
package config

import (
	"flag"
	"fmt"
	"gophkeeper/internal/database"
	"log"
	"os"
	"strconv"
)

// Config описывает конфигурацию серверного приложения.
type Config struct {
	// GRPCServerAddress содержит адрес запуска gRPC-сервера.
	GRPCServerAddress string
	// BaseURL содержит базовый URL приложения.
	BaseURL string
	// DatabaseDSN содержит строку подключения к базе данных PostgreSQL.
	DatabaseDSN string
	// EnableHTTPS определяет, должен ли сервер запускаться с TLS.
	EnableHTTPS bool

	// DB содержит инициализированное подключение к базе данных.
	DB *database.DB
}

// FileConfig описывает структуру конфигурации, которая может быть
// загружена из внешнего файла.
type FileConfig struct {
	// GRPCServerAddress содержит адрес запуска gRPC-сервера.
	GRPCServerAddress string `json:"grpc_server_address"`
	// BaseURL содержит базовый URL приложения.
	BaseURL string `json:"base_url"`
	// DatabaseDSN содержит строку подключения к базе данных.
	DatabaseDSN string `json:"database_dsn"`
	// EnableHTTPS определяет, должен ли сервер запускаться с TLS.
	EnableHTTPS *bool `json:"enable_https"`
}

// Init инициализирует конфигурацию приложения.
//
// Значения по умолчанию читаются из флагов командной строки,
// затем поверх них накладываются значения из переменных окружения.
// После этого выполняется попытка инициализировать подключение к базе данных.
func Init() *Config {
	cfg := &Config{}

	defGRPCAddr := "localhost:3200"
	defBase := "http://localhost:8080"
	defDSN := "postgres://root:root@localhost:5433/gophkeeper"
	defHTTPS := false

	flag.StringVar(&cfg.GRPCServerAddress, "g", defGRPCAddr, "gRPC server address")
	flag.StringVar(&cfg.BaseURL, "b", defBase, "Base URL for short links")
	flag.StringVar(&cfg.DatabaseDSN, "d", defDSN, "Database DSN")
	flag.BoolVar(&cfg.EnableHTTPS, "s", defHTTPS, "Start server with HTTPS")

	flag.Parse()

	applyEnv(cfg)

	cfg.initDB()
	return cfg
}

// applyEnv применяет к конфигурации значения из переменных окружения.
//
// Если соответствующая переменная окружения установлена,
// её значение переопределяет текущее значение конфигурации.
func applyEnv(cfg *Config) {
	if v := os.Getenv("GRPC_SERVER_ADDRESS"); v != "" {
		cfg.GRPCServerAddress = v
	}
	if v := os.Getenv("BASE_URL"); v != "" {
		cfg.BaseURL = v
	}
	if v := os.Getenv("DATABASE_DSN"); v != "" {
		cfg.DatabaseDSN = v
	}
	if v := os.Getenv("ENABLE_HTTPS"); v != "" {
		val, err := strconv.ParseBool(v)
		if err != nil {
			log.Printf("Error: %v", err)
			return
		}
		cfg.EnableHTTPS = val
	}
}

// Validate проверяет обязательные поля конфигурации.
//
// Возвращает ошибку, если обязательные параметры не заданы.
func (c *Config) Validate() error {
	if c.GRPCServerAddress == "" {
		return fmt.Errorf("grpc server address cannot be empty")
	}
	if c.BaseURL == "" {
		return fmt.Errorf("base URL cannot be empty")
	}
	return nil
}

// initDB инициализирует подключение к базе данных,
// если строка подключения задана.
//
// В случае ошибки подключения функция пишет сообщение в лог
// и оставляет поле DB незаполненным.
func (c *Config) initDB() {
	if c.DatabaseDSN == "" {
		return
	}

	db, err := database.NewDB(c.DatabaseDSN)
	if err != nil {
		log.Printf("Failed to connect to PostgreSQL: %v", err)
		return
	}

	c.DB = db
	log.Printf("Connected to PostgreSQL")
}

// Close закрывает все ресурсы, связанные с конфигурацией.
//
// В текущей реализации закрывает подключение к базе данных,
// если оно было инициализировано.
func (c *Config) Close() error {
	if c.DB != nil {
		_ = c.DB.Close()
	}
	return nil
}
