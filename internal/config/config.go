package config

import (
	"flag"
	"fmt"
	"gophkeeper/internal/database"
	"log"
	"os"
	"strconv"
)

type Config struct {
	GRPCServerAddress string
	BaseURL           string
	DatabaseDSN       string
	EnableHTTPS       bool

	DB *database.DB
}

type FileConfig struct {
	GRPCServerAddress string `json:"grpc_server_address"`
	BaseURL           string `json:"base_url"`
	DatabaseDSN       string `json:"database_dsn"`
	EnableHTTPS       *bool  `json:"enable_https"`
}

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

func (c *Config) Validate() error {
	if c.GRPCServerAddress == "" {
		return fmt.Errorf("grpc server address cannot be empty")
	}
	if c.BaseURL == "" {
		return fmt.Errorf("base URL cannot be empty")
	}
	return nil
}

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

func (c *Config) Close() error {
	if c.DB != nil {
		_ = c.DB.Close()
	}
	return nil
}
