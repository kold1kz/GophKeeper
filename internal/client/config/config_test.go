package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoad_Defaults(t *testing.T) {
	t.Setenv("GOPHKEEPER_SERVER_ADDRESS", "")

	wd := t.TempDir()
	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd returned error: %v", err)
	}
	defer func() { _ = os.Chdir(oldWD) }()

	if err := os.Chdir(wd); err != nil {
		t.Fatalf("Chdir returned error: %v", err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.ServerAddress != "localhost:3200" {
		t.Fatalf("expected default server address, got %q", cfg.ServerAddress)
	}

	wantSuffix := filepath.Join(".gophkeeper", "state.json")
	if !strings.HasSuffix(cfg.StatePath, wantSuffix) {
		t.Fatalf("expected state path to end with %q, got %q", wantSuffix, cfg.StatePath)
	}

	if filepath.Base(cfg.StatePath) != "state.json" {
		t.Fatalf("expected file name state.json, got %q", filepath.Base(cfg.StatePath))
	}
}

func TestLoad_FromEnv(t *testing.T) {
	t.Setenv("GOPHKEEPER_SERVER_ADDRESS", "127.0.0.1:9999")

	wd := t.TempDir()
	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd returned error: %v", err)
	}
	defer func() { _ = os.Chdir(oldWD) }()

	if err := os.Chdir(wd); err != nil {
		t.Fatalf("Chdir returned error: %v", err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.ServerAddress != "127.0.0.1:9999" {
		t.Fatalf("expected env server address, got %q", cfg.ServerAddress)
	}
}
