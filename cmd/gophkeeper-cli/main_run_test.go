package main

import (
	"os"
	"testing"
)

func TestRun_UnknownCommand(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"gophkeeper-cli", "unknown-command"}

	if err := run(); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestRun_NoArgs(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"gophkeeper-cli"}

	if err := run(); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestRun_Register_InvalidArgs(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"gophkeeper-cli", "register"}

	if err := run(); err == nil {
		t.Fatal("expected error")
	}
}

func TestRun_Login_InvalidArgs(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"gophkeeper-cli", "login"}

	if err := run(); err == nil {
		t.Fatal("expected error")
	}
}

func TestRun_CreateText_InvalidArgs(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"gophkeeper-cli", "create-text", "title"}

	if err := run(); err == nil {
		t.Fatal("expected error")
	}
}

func TestRun_GetLocal_InvalidArgs(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"gophkeeper-cli", "get-local", "id-only"}

	if err := run(); err == nil {
		t.Fatal("expected error")
	}
}

func TestRun_UpdateText_InvalidArgs(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"gophkeeper-cli", "update-text", "id"}

	if err := run(); err == nil {
		t.Fatal("expected error")
	}
}

func TestRun_DeleteItem_InvalidArgs(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"gophkeeper-cli", "delete-item"}

	if err := run(); err == nil {
		t.Fatal("expected error")
	}
}

func TestRun_CreateLogin_InvalidArgs(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"gophkeeper-cli", "create-login", "title"}

	if err := run(); err == nil {
		t.Fatal("expected error")
	}
}

func TestRun_CreateCard_InvalidArgs(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"gophkeeper-cli", "create-card", "title"}

	if err := run(); err == nil {
		t.Fatal("expected error")
	}
}

func TestRun_CreateFile_InvalidArgs(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"gophkeeper-cli", "create-file", "title"}

	if err := run(); err == nil {
		t.Fatal("expected error")
	}
}

func TestRun_GetFile_InvalidArgs(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"gophkeeper-cli", "get-file", "id"}

	if err := run(); err == nil {
		t.Fatal("expected error")
	}
}

func TestRun_Sync_ExtraArgs(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"gophkeeper-cli", "sync", "extra"}

	if err := run(); err == nil {
		t.Fatal("expected error")
	}
}

func TestRun_ListLocal_ExtraArgs(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"gophkeeper-cli", "list-local", "extra"}

	err := run()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

}

func TestRun_ListRemote_ExtraArgs(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"gophkeeper-cli", "list-remote", "extra"}

	if err := run(); err == nil {
		t.Fatal("expected error")
	}
}

func TestRun_Version(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"gophkeeper-cli", "version"}

	if err := run(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRun_UpdateText_InvalidArgs2(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"gophkeeper-cli", "update-text", "id"}

	if err := run(); err == nil {
		t.Fatal("expected error")
	}
}

func TestRun_DeleteItem_InvalidArgs2(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"gophkeeper-cli", "delete-item"}

	if err := run(); err == nil {
		t.Fatal("expected error")
	}
}
