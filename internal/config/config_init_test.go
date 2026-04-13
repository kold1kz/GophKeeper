package config

import (
	"flag"
	"os"
	"testing"
)

func TestInit_Defaults(t *testing.T) {
	oldArgs := os.Args
	oldFlagSet := flag.CommandLine
	defer func() {
		os.Args = oldArgs
		flag.CommandLine = oldFlagSet
	}()

	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	os.Args = []string{"cmd"}

	cfg := Init()
	if cfg == nil {
		t.Fatal("expected config")
	}
}

func TestInit_WithEnv(t *testing.T) {
	oldArgs := os.Args
	oldFlagSet := flag.CommandLine
	defer func() {
		os.Args = oldArgs
		flag.CommandLine = oldFlagSet
	}()

	t.Setenv("GRPC_SERVER_ADDRESS", "127.0.0.1:9000")
	t.Setenv("BASE_URL", "http://localhost:8080")

	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	os.Args = []string{"cmd"}

	cfg := Init()
	if cfg == nil {
		t.Fatal("expected config")
	}
}
