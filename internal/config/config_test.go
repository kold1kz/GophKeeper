package config

import "testing"

func TestApplyEnv(t *testing.T) {
	t.Setenv("GRPC_SERVER_ADDRESS", "127.0.0.1:9000")
	t.Setenv("BASE_URL", "http://localhost:8080")
	t.Setenv("DATABASE_DSN", "postgres://user:pass@localhost:5432/db")
	t.Setenv("ENABLE_HTTPS", "true")

	cfg := &Config{}

	applyEnv(cfg)

	if cfg.GRPCServerAddress != "127.0.0.1:9000" {
		t.Fatalf("unexpected GRPCServerAddress: %q", cfg.GRPCServerAddress)
	}
	if cfg.BaseURL != "http://localhost:8080" {
		t.Fatalf("unexpected BaseURL: %q", cfg.BaseURL)
	}
	if cfg.DatabaseDSN != "postgres://user:pass@localhost:5432/db" {
		t.Fatalf("unexpected DatabaseDSN: %q", cfg.DatabaseDSN)
	}
	if !cfg.EnableHTTPS {
		t.Fatal("expected EnableHTTPS=true")
	}
}

func TestValidate_Success(t *testing.T) {
	cfg := &Config{
		GRPCServerAddress: "localhost:3200",
		BaseURL:           "http://localhost:8080",
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}
}

func TestValidate_EmptyGRPCServerAddress(t *testing.T) {
	cfg := &Config{
		BaseURL: "http://localhost:8080",
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestValidate_EmptyBaseURL(t *testing.T) {
	cfg := &Config{
		GRPCServerAddress: "localhost:3200",
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestClose_NoDB(t *testing.T) {
	cfg := &Config{}

	if err := cfg.Close(); err != nil {
		t.Fatalf("Close returned error: %v", err)
	}
}
